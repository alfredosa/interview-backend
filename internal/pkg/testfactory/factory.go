package testfactory

import (
	pb "github.com/coopnorge/interview-backend/internal/generated/logistics/api/v1"
)

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

// NewGetWarehouseRequest creates a new instance of pb.GetWarehouseRequest with the specified warehouse ID.
func NewGetWarehouseRequest(warehouseId int64) *pb.GetWarehouseRequest {
	return &pb.GetWarehouseRequest{
		WarehouseId: warehouseId,
	}
}

// DefaultGetWarehouseRequest generates a default instance of pb.GetWarehouseRequest for testing or initial setup.
func DefaultGetWarehouseRequest() *pb.GetWarehouseRequest {
	return NewGetWarehouseRequest(2001) // Example warehouse ID for default request
}
