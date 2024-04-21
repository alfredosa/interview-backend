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

type GrpcService struct {
	pb.UnimplementedCoopLogisticsEngineAPIServer
	Stats      *ServerStatistics
	Controller *receiver.LogisticsController

	destroych chan bool
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

	logInterval := 1 * time.Second
	go grpcService.setupSignalHandler()

	go func() {
		log.Printf("Server listening at %v", lis.Addr())
		if err := server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	go grpcService.printServerStats(logInterval)
	<-grpcService.destroych
}

func NewGrpcServerAndService() (*grpc.Server, *GrpcService) {
	stats := &ServerStatistics{}
	controller := receiver.NewLogisticsController()

	grpcServer := &GrpcService{
		Stats:      stats,
		Controller: controller,
		destroych:  make(chan bool),
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(apiHitsInterceptor(grpcServer.Stats)),
	)

	return server, grpcServer
}

type ServerStatistics struct {
	apiHits uint64
}

// Increment API Hits from Server Statistics
func (s *ServerStatistics) HitIncrement() {
	atomic.AddUint64(&s.apiHits, 1)
}

// If we need to read the hits at any point in time
func (s *GrpcService) GetHits() uint64 {
	return atomic.LoadUint64(&s.Stats.apiHits)
}

// Prints server stats per Bucket of time.
func (s *GrpcService) printServerStats(t time.Duration) {
	var consecutiveZeros int64
	ctx := context.Background()
	for {
		time.Sleep(t)
		hitAmount := s.Stats.resetAndPrintHits()

		if hitAmount == 0 {
			consecutiveZeros++
		} else {
			consecutiveZeros = 0
		}

		if consecutiveZeros == 5 {
			s.Controller.PrintWarehousesSummary(ctx)
			slog.Info("Gracefully destroying the server, Goodbye")
			s.destroych <- true
			break
		}
	}
}

// Gets the current Hit Count and Resets (Swaps) the value to 0
func (st *ServerStatistics) resetAndPrintHits() uint64 {
	currentHits := atomic.SwapUint64(&st.apiHits, 0)
	slog.Info("Statistics", "Server API hits", currentHits)
	return currentHits
}

// Middleware / Interceptor to capture API Statistics
func apiHitsInterceptor(stats *ServerStatistics) grpc.UnaryServerInterceptor {
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
		s.destroych <- true
	}()
}
