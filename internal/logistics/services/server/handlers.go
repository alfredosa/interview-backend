package server

import (
	"context"

	pb "github.com/coopnorge/interview-backend/internal/generated/logistics/api/v1"
)

// MoveUnit is a Handler for the MoveUnirRequest,
// It will process the Cargo Unit location and update our tracker with the information or create one if it's a new Cargo Unit
func (s *GrpcService) MoveUnit(ctx context.Context, in *pb.MoveUnitRequest) (*pb.DefaultResponse, error) {
	err := s.Controller.MoveUnit(ctx, in)
	if err != nil {
		return &pb.DefaultResponse{}, err
	}
	return &pb.DefaultResponse{}, nil
}

// UnitReachedWarehouse is a Handler for the Request UnitReachedWarehouse. It will process Creating or Updating a Warehouse, keeping track of delivery units.
// returns an error if it fails to do so.
func (s *GrpcService) UnitReachedWarehouse(ctx context.Context, in *pb.UnitReachedWarehouseRequest) (*pb.DefaultResponse, error) {
	err := s.Controller.WarehouseReceivedProcessing(ctx, in)

	if err != nil {
		return &pb.DefaultResponse{}, err
	}

	return &pb.DefaultResponse{}, nil
}

// GetWarehouse is a Handler for the GetWarehouse request. It retrieves warehouse information along with its suppliers.
func (s *GrpcService) GetWarehouse(ctx context.Context, in *pb.GetWarehouseRequest) (*pb.GetWarehouseResponse, error) {
	response, err := s.Controller.GetWarehouseResponse(ctx, in)
	if err != nil {
		return nil, err
	}
	return response, nil
}
