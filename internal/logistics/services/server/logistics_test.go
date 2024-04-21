package server

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	pb "github.com/coopnorge/interview-backend/internal/generated/logistics/api/v1"
	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupTestServer(t *testing.T) (*GrpcServer, pb.CoopLogisticsEngineAPIClient, func()) {
	// random available addrs
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	s, grpcServer := NewServiceAndGrpcServer()
	pb.RegisterCoopLogisticsEngineAPIServer(s, grpcServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("Failed to serve: %v", err)
		}
	}()

	// Setup client
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

// NewLocation returns a new instance of pb.Location pre-populated with default values.
func NewLocation(latitude, longitude uint32) *pb.Location {
	return &pb.Location{
		Latitude:  latitude,
		Longitude: longitude,
	}
}

// NewWarehouseAnnouncement returns a new instance of pb.WarehouseAnnouncement pre-populated with default values.
func NewWarehouseAnnouncement(cargoUnitId, warehouseId int64, message string) *pb.WarehouseAnnouncement {
	return &pb.WarehouseAnnouncement{
		CargoUnitId: cargoUnitId,
		WarehouseId: warehouseId,
		Message:     message,
	}
}

// NewUnitReachedWarehouseRequest creates a new instance of pb.UnitReachedWarehouseRequest with given location and announcement.
func NewUnitReachedWarehouseRequest(location *pb.Location, announcement *pb.WarehouseAnnouncement) *pb.UnitReachedWarehouseRequest {
	return &pb.UnitReachedWarehouseRequest{
		Location:     location,
		Announcement: announcement,
	}
}

// DefaultUnitReachedWarehouseRequest generates a default instance of pb.UnitReachedWarehouseRequest for testing or initial setup.
func DefaultUnitReachedWarehouseRequest() *pb.UnitReachedWarehouseRequest {
	location := NewLocation(123456789, 987654321)
	announcement := NewWarehouseAnnouncement(1001, 5001, "New cargo unit received at warehouse.")
	return NewUnitReachedWarehouseRequest(location, announcement)
}

// NewMoveUnitRequest creates a new instance of pb.MoveUnitRequest with the specified cargo unit ID and location.
func NewMoveUnitRequest(cargoUnitId int64, location *pb.Location) *pb.MoveUnitRequest {
	return &pb.MoveUnitRequest{
		CargoUnitId: cargoUnitId,
		Location:    location,
	}
}

// DefaultMoveUnitRequest generates a default instance of pb.MoveUnitRequest for testing or initial setup.
func DefaultMoveUnitRequest() *pb.MoveUnitRequest {
	defaultLocation := NewLocation(123456789, 987654321)
	return NewMoveUnitRequest(101, defaultLocation)
}

func TestMoveUnit(t *testing.T) {
	s, client, cleanup := SetupTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := DefaultMoveUnitRequest()

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

	req := DefaultUnitReachedWarehouseRequest()
	resp, err := client.UnitReachedWarehouse(ctx, req)
	fmt.Println(resp)
	if err != nil {
		fmt.Printf("Failed to reach warehouse with %s", err)
	}

	assert.Equal(t, uint64(1), s.Stats.apiHits)
	assert.Nil(t, err)
	assert.Equal(t, len(s.Controller.Warehouses), 1)
	assert.Equal(t, len(s.Controller.Warehouses[5001].Suppliers), 1)
}
