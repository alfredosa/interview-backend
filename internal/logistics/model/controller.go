package model

import (
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
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

func NewSupplier(lat uint32, lon uint32) *Supplier {
	location := &Location{
		Lattitude: lat,
		Longitude: lon,
	}

	return &Supplier{
		Location: location,
	}
}

var ErrSupplierDoesNotExist = errors.New("supplier does not exist")

// Sets New Coordinates of a supplier
func (wa *Warehouse) SetSafeSupplierCoordinates(id int64, lat uint32, lon uint32) error {
	if !wa.SupplierExists(id) {
		return ErrSupplierDoesNotExist
	}

	supplier := wa.GetSupplier(id)
	supplier.SetSafeCoordinates(lat, lon)
	return nil
}

func (su *Supplier) SetSafeCoordinates(lat uint32, lon uint32) {
	su.supplierMu.Lock()
	defer su.supplierMu.Unlock()

	su.Location.Lattitude = lat
	su.Location.Longitude = lon
}

// Increments the UnitsReceived counter atomically.
func (wa *Warehouse) AddUnit() {
	atomic.AddUint64(&wa.UnitsReceived, 1)
}

// GetUnits returns the current value of the UnitsReceived counter.
func (wa *Warehouse) GetUnits() uint64 {
	return atomic.LoadUint64(&wa.UnitsReceived)
}

// Safely add a supplier to the Supplier pool of a given Warehouse
func (wa *Warehouse) SafeSuppliersAdd(id int64, supplier *Supplier) {
	wa.supplierMu.Lock()
	defer wa.supplierMu.Unlock()
	wa.Suppliers[id] = supplier
}

// Check if a Supplier Exists in the Pool of Suppliers delivering to this Warehouse
func (wa *Warehouse) SupplierExists(id int64) bool {
	wa.supplierMu.RLock()
	defer wa.supplierMu.RUnlock()

	_, exists := wa.Suppliers[id]
	return exists
}

// GetSupplier retrieves a supplier from the warehouse safely.
func (wa *Warehouse) GetSupplier(id int64) *Supplier {
	wa.supplierMu.RLock()
	defer wa.supplierMu.RUnlock()
	supplier, exists := wa.Suppliers[id]
	if !exists {
		return nil
	}
	return supplier
}

func (wa *Warehouse) GetWarehouseSummary(id int64) {
	slog.Info("Warehouse State", "ID", id, "Units", wa.UnitsReceived, "Lat", wa.Location.Lattitude, "Lon", wa.Location.Longitude)
}
