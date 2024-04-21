package server

import (
	"context"

	pb "github.com/coopnorge/interview-backend/internal/generated/logistics/api/v1"
)

func (s *GrpcService) MoveUnit(ctx context.Context, in *pb.MoveUnitRequest) (*pb.DefaultResponse, error) {
	err := s.Controller.MoveUnit(ctx, in)
	if err != nil {
		return &pb.DefaultResponse{}, err
	}
	return &pb.DefaultResponse{}, nil
}

// Handler for the Request UnitReachedWarehouse. It will process the data and return an error if it fails to do so.
func (s *GrpcService) UnitReachedWarehouse(ctx context.Context, in *pb.UnitReachedWarehouseRequest) (*pb.DefaultResponse, error) {
	err := s.Controller.WarehouseReceivedProcessing(ctx, in)

	if err != nil {
		return &pb.DefaultResponse{}, ctx.Err()
	}

	return &pb.DefaultResponse{}, nil
}

// Handler for the GetWarehouse request. It retrieves warehouse information along with its suppliers.
func (s *GrpcService) GetWarehouse(ctx context.Context, in *pb.GetWarehouseRequest) (*pb.GetWarehouseResponse, error) {
	response, err := s.Controller.GetWarehouseResponse(ctx, in)
	if err != nil {
		return nil, err
	}
	return response, nil
}
