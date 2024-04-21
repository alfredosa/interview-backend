package server

import (
	"context"
	"fmt"
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

type GrpcServer struct {
	pb.UnimplementedCoopLogisticsEngineAPIServer
	Stats      *ServerStatistics
	Controller *receiver.LogisticsController

	Destroyer chan bool
}

func (s *GrpcServer) MoveUnit(ctx context.Context, in *pb.MoveUnitRequest) (*pb.DefaultResponse, error) {
	err := s.Controller.MoveUnit(ctx, in)
	if err != nil {
		return &pb.DefaultResponse{}, err
	}
	return &pb.DefaultResponse{}, nil
}

func (s *GrpcServer) UnitReachedWarehouse(ctx context.Context, in *pb.UnitReachedWarehouseRequest) (*pb.DefaultResponse, error) {
	s.Controller.WarehouseReceivedProcessing(ctx, in)

	if ctx.Err() != nil {
		return &pb.DefaultResponse{}, ctx.Err()
	}

	return &pb.DefaultResponse{}, nil
}

func ListendAndAccept(cfg *config.ServerConfig) {
	lis, err := net.Listen("tcp", cfg.GetCombinedAddress())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	service, grpcServer := NewServiceAndGrpcServer()

	pb.RegisterCoopLogisticsEngineAPIServer(service, grpcServer)

	// NOTE: env var?
	logInterval := 1 * time.Second
	go grpcServer.printServerStats(logInterval)
	go grpcServer.setupSignalHandler()

	log.Printf("Server listening at %v", lis.Addr())
	go func() {
		if err := service.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-grpcServer.Destroyer
	lis.Close()
}

func NewServiceAndGrpcServer() (*grpc.Server, *GrpcServer) {
	stats := &ServerStatistics{}
	controller := receiver.NewLogisticsController()

	grpcServer := &GrpcServer{
		Stats:      stats,
		Controller: controller,
		Destroyer:  make(chan bool),
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
func (s *GrpcServer) GetHits() uint64 {
	return atomic.LoadUint64(&s.Stats.apiHits)
}

// Prints server stats per Bucket of time.
func (s *GrpcServer) printServerStats(t time.Duration) {
	var consecutiveZeros int64
	for {
		time.Sleep(t)
		hitAmount := s.Stats.resetAndPrintHits()

		if hitAmount == 0 {
			consecutiveZeros++
		} else {
			consecutiveZeros = 0
		}

		if consecutiveZeros == 5 {
			s.Controller.PrintWarehousesSummary()
			slog.Info("Gracefully destroying the server, Goodbye")
			s.Destroyer <- true
			break
		}
	}
}

// Gets the current Hit Count and Resets (Swaps) the value to 0
func (st *ServerStatistics) resetAndPrintHits() uint64 {
	currentHits := atomic.SwapUint64(&st.apiHits, 0)
	fmt.Printf("Server API hits: %d\n", currentHits)
	return currentHits
}

// Middleware to capture API Statistics
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

func (s *GrpcServer) setupSignalHandler() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		s.Destroyer <- true
	}()
}
