package model

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	pb "github.com/coopnorge/interview-backend/internal/generated/logistics/api/v1"
)

type Supplier struct {
	Location *Location

	supplierMu sync.RWMutex
}

type Warehouse struct {
	UnitsReceived uint64
	Suppliers     map[int64]*Supplier
	Location      *Location

	supplierMu sync.RWMutex
}

// Generic Location for both Suppliers and Warehouses
type Location struct {
	Lattitude uint32
	Longitude uint32
}

func NewWarehouse(lat uint32, lon uint32) *Warehouse {
	location := &Location{
		Lattitude: lat,
		Longitude: lon,
	}
	return &Warehouse{
		UnitsReceived: 0,
		Suppliers:     make(map[int64]*Supplier, 256),
		Location:      location,
	}
}

func NewSupplier(ctx context.Context, lat uint32, lon uint32) (*Supplier, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	location := &Location{
		Lattitude: lat,
		Longitude: lon,
	}

	return &Supplier{
		Location: location,
	}, nil
}

var ErrSupplierDoesNotExist = errors.New("supplier does not exist")

// Sets New Coordinates of a Warehouse Supplier
func (wa *Warehouse) SetSafeSupplierCoordinates(ctx context.Context, id int64, lat uint32, lon uint32) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if !wa.SupplierExists(id) {
		return ErrSupplierDoesNotExist
	}

	supplier, err := wa.GetSupplier(ctx, id)
	if err != nil {
		return err
	}
	supplier.SetSafeCoordinates(ctx, lat, lon)
	return nil
}

func (su *Supplier) SetSafeCoordinates(ctx context.Context, lat uint32, lon uint32) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	su.supplierMu.Lock()
	defer su.supplierMu.Unlock()

	su.Location.Lattitude = lat
	su.Location.Longitude = lon

	return nil
}

// Increments the UnitsReceived counter atomically.
func (wa *Warehouse) AddUnit() {
	atomic.AddUint64(&wa.UnitsReceived, 1)
}

// GetUnits returns the current value of the UnitsReceived counter.
func (wa *Warehouse) GetUnits(ctx context.Context) uint64 {
	return atomic.LoadUint64(&wa.UnitsReceived)
}

// Safely add a supplier to the Supplier pool of a given Warehouse
func (wa *Warehouse) SafeSuppliersAdd(ctx context.Context, id int64, supplier *Supplier) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	wa.supplierMu.Lock()
	defer wa.supplierMu.Unlock()
	wa.Suppliers[id] = supplier
	return nil
}

// Check if a Supplier Exists in the Pool of Suppliers delivering to this Warehouse
func (wa *Warehouse) SupplierExists(id int64) bool {
	wa.supplierMu.RLock()
	defer wa.supplierMu.RUnlock()

	_, exists := wa.Suppliers[id]
	return exists
}

// GetSupplier retrieves a supplier from the warehouse safely.
func (wa *Warehouse) GetSupplier(ctx context.Context, id int64) (*Supplier, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	wa.supplierMu.RLock()
	defer wa.supplierMu.RUnlock()
	supplier, exists := wa.Suppliers[id]
	if !exists {
		return nil, ErrSupplierDoesNotExist
	}
	return supplier, nil
}

// GetProtoSuppliers retrieves all Suppliers in protobuf format. It checks for context cancellation to handle long-running or halted requests appropriately.
func (wa *Warehouse) GetProtoSuppliers(ctx context.Context) ([]*pb.Supplier, error) {
	wa.supplierMu.RLock()
	defer wa.supplierMu.RUnlock()

	suppliers := make([]*pb.Supplier, 0, len(wa.Suppliers))

	for supplierID, supplier := range wa.Suppliers {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		suppliers = append(suppliers, &pb.Supplier{
			SupplierId: supplierID,
			Location: &pb.Location{
				Latitude:  supplier.Location.Lattitude,
				Longitude: supplier.Location.Longitude,
			},
		})
	}

	return suppliers, nil
}

func (wa *Warehouse) GetWarehouseSummary(id int64) {
	slog.Info("Warehouse State", "ID", id, "Units", wa.UnitsReceived, "Lat", wa.Location.Lattitude, "Lon", wa.Location.Longitude)
}
