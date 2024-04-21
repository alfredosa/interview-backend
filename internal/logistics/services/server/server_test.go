package server

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	pb "github.com/coopnorge/interview-backend/internal/generated/logistics/api/v1"
	factory "github.com/coopnorge/interview-backend/internal/pkg/testfactory"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func SetupTestServer(t *testing.T) (*GrpcService, pb.CoopLogisticsEngineAPIClient, func()) {
	// random available addrs
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	s, grpcServer := NewGrpcServerAndService()
	pb.RegisterCoopLogisticsEngineAPIServer(s, grpcServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}()

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}

	client := pb.NewCoopLogisticsEngineAPIClient(conn)

	return grpcServer, client, func() {
		s.Stop()
		conn.Close()
		lis.Close()
	}
}

func TestMoveUnit(t *testing.T) {
	s, client, cleanup := SetupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := factory.DefaultMoveUnitRequest()

	_, err := client.MoveUnit(ctx, req)
	if err != nil {
		t.Errorf("MoveUnit failed: %v", err)
	}

	assert.Nil(t, err)
	assert.Equal(t, uint64(1), s.Stats.apiHits)
	assert.Equal(t, len(s.Controller.Suppliers), 1)
	assert.Equal(t, s.Controller.Suppliers[101].Location.Lattitude, uint32(123456789))
}

func TestUnitReachedWarehouse(t *testing.T) {
	s, client, cleanup := SetupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := factory.DefaultUnitReachedWarehouseRequest()
	_, err := client.UnitReachedWarehouse(ctx, req)

	if err != nil {
		fmt.Printf("Failed to reach warehouse with %s", err)
	}

	assert.Equal(t, uint64(1), s.Stats.apiHits)
	assert.Nil(t, err)
	assert.Equal(t, len(s.Controller.Warehouses), 1)
	assert.Equal(t, len(s.Controller.Warehouses[5001].Suppliers), 1)
}

func TestGetWarehouse(t *testing.T) {
	_, client, cleanup := SetupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := factory.DefaultGetWarehouseRequest()
	resp, err := client.GetWarehouse(ctx, req)

	assert.Nil(t, resp)
	if assert.Error(t, err) {
		grpcStatus, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, grpcStatus.Code())
		assert.Contains(t, grpcStatus.Message(), "warehouse does not exist")
	}

	warehousereq := factory.DefaultUnitReachedWarehouseRequest()
	_, err = client.UnitReachedWarehouse(ctx, warehousereq)
	if err != nil {
		fmt.Printf("Failed to reach warehouse with %s", err)
	}

	req = factory.NewGetWarehouseRequest(warehousereq.Announcement.WarehouseId)
	resp, err = client.GetWarehouse(ctx, req)

	assert.Nil(t, err)
	assert.Equal(t, len(resp.Warehouse.Suppliers), 1)
	assert.Equal(t, resp.Warehouse.WarehouseId, warehousereq.Announcement.WarehouseId)

	fmt.Println(resp)
}
