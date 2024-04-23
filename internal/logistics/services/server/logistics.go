package server

import (
	"context"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	pb "github.com/coopnorge/interview-backend/internal/generated/logistics/api/v1"
	"github.com/coopnorge/interview-backend/internal/logistics/config"
	"github.com/coopnorge/interview-backend/internal/pkg/receiver"

	"google.golang.org/grpc"
)

// GrpcService provides the necessary setup for handling gRPC API calls for the Coop Logistics Engine.
// It includes server statistics, a logistics controller, and a channel for server shutdown.
type GrpcService struct {
	pb.UnimplementedCoopLogisticsEngineAPIServer
	Stats      *ServerStatistics
	Controller *receiver.LogisticsController

	stopch chan bool
}

// Server Statistics represents all values that we want to track about our api, from performance to usage
type ServerStatistics struct {
	apiHits uint64
}

// ListenAndAccept starts the grpc server and beginns accepting incomign requests
// It will initiallize our logger and listen for signal terminations.
func ListendAndAccept(cfg *config.ServerConfig) {
	lis, err := net.Listen("tcp", cfg.GetCombinedAddress())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server, grpcService := NewGrpcServerAndService()
	pb.RegisterCoopLogisticsEngineAPIServer(server, grpcService)
	go grpcService.setupSignalHandler()

	parentCtx := context.Background()
	ctxWithTime, cancel := context.WithTimeout(parentCtx, 20*time.Second)
	defer cancel()

	go func() {
		log.Printf("Server listening at %v", lis.Addr())
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	logInterval := 1 * time.Second
	go grpcService.printServerStats(ctxWithTime, logInterval)

	<-grpcService.stopch
	grpcService.Controller.PrintWarehousesSummary(parentCtx)
	server.GracefulStop()
}

func NewGrpcServerAndService() (*grpc.Server, *GrpcService) {
	stats := NewServerStatistics()
	controller := receiver.NewLogisticsController()

	grpcService := &GrpcService{
		Stats:      stats,
		Controller: controller,
		stopch:     make(chan bool),
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcService.Stats.apiHitsInterceptor()),
	)

	return server, grpcService
}

// Creates new Server Statistics struct
func NewServerStatistics() *ServerStatistics {
	return &ServerStatistics{
		apiHits: uint64(0),
	}
}

// Increment API Hits from Server Statistics
func (s *ServerStatistics) HitIncrement() {
	atomic.AddUint64(&s.apiHits, 1)
}

// If we need to read the hits at any point in time
func (s *GrpcService) GetHits() uint64 {
	return atomic.LoadUint64(&s.Stats.apiHits)
}

// Prints server stats at regular intervals until the context is done.
func (s *GrpcService) printServerStats(ctx context.Context, t time.Duration) {
	ticker := time.NewTicker(t)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Gracefully destroying the server, Goodbye")
			s.stopch <- true
			return

		case <-ticker.C:
			s.Stats.resetAndPrintHits()
		}
	}
}

// Gets the current Hit Count and Resets (Swaps) the value to 0
func (st *ServerStatistics) resetAndPrintHits() {
	currentHits := atomic.SwapUint64(&st.apiHits, 0)
	slog.Info("Statistics", "Server API hits", currentHits)
}

// Middleware / Interceptor to capture API Statistics
func (stats *ServerStatistics) apiHitsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		stats.HitIncrement()
		resp, err := handler(ctx, req)
		return resp, err
	}
}

func (s *GrpcService) setupSignalHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		s.stopch <- true
	}()
}
