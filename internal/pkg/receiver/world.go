package receiver

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"sync"

	pb "github.com/coopnorge/interview-backend/internal/generated/logistics/api/v1"
	"github.com/coopnorge/interview-backend/internal/logistics/model"
)

// Just going on a hunch here, its nice if we init maps with a given length / capacity
// In a real world scenario we know an estimated amount
var (
	initialNWarehouses = 50
	initialNSuppliers  = 256
)

type LogisticsController struct {
	Warehouses map[int64]*model.Warehouse
	Suppliers  map[int64]*model.Supplier

	warehouseMu sync.RWMutex
	supplierMu  sync.RWMutex
}

func NewLogisticsController() *LogisticsController {
	return &LogisticsController{
		Warehouses: make(map[int64]*model.Warehouse, initialNWarehouses),
		Suppliers:  make(map[int64]*model.Supplier, initialNSuppliers),
	}
}

func (lc *LogisticsController) SafeAddWarehouse(id int64, warehouse *model.Warehouse) {
	lc.warehouseMu.Lock()
	defer lc.warehouseMu.Unlock()
	lc.Warehouses[id] = warehouse
}

func (lc *LogisticsController) SafeAddSupplier(id int64, supplier *model.Supplier) {
	lc.supplierMu.Lock()
	defer lc.supplierMu.Unlock()
	lc.Suppliers[id] = supplier
}

func (lc *LogisticsController) MoveUnit(ctx context.Context, in *pb.MoveUnitRequest) error {
	var (
		supplierID = in.CargoUnitId
		lat        = in.Location.Latitude
		lon        = in.Location.Longitude
	)

	lc.UpdateOrCreateSupplier(ctx, supplierID, lat, lon)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

var ErrFailedToAddSupplierCoordinates = errors.New("Failed to add Supplier Coordinates")

func (lc *LogisticsController) WarehouseReceivedProcessing(ctx context.Context, in *pb.UnitReachedWarehouseRequest) error {
	var (
		warehouseID = in.Announcement.WarehouseId
		supplierID  = in.Announcement.CargoUnitId
		lat         = in.Location.Latitude
		lon         = in.Location.Longitude
	)

	// TODO: Add error
	warehouse := lc.GetOrCreateWarehouse(ctx, warehouseID, lat, lon)

	warehouse.AddUnit()

	if !warehouse.SupplierExists(supplierID) {
		supplier := model.NewSupplier(lat, lon)

		warehouse.SafeSuppliersAdd(supplierID, supplier)
	}

	err := warehouse.SetSafeSupplierCoordinates(supplierID, lat, lon)
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), "lat", lat, "lon", lon)
		return ErrFailedToAddSupplierCoordinates
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	// warehouse.GetWarehouseSummary(warehouseID)
	return nil
}

// Get a Warehouse from the existing warehouses or add a warehouse
func (lc *LogisticsController) GetOrCreateWarehouse(ctx context.Context, id int64, lat uint32, lon uint32) *model.Warehouse {
	if lc.WarehouseExists(id) {
		return lc.GetWarehouse(id)
	}

	warehouse := model.NewWarehouse(lat, lon)
	lc.SafeAddWarehouse(id, warehouse)
	return warehouse
}

func (lc *LogisticsController) UpdateOrCreateSupplier(ctx context.Context, id int64, lat uint32, lon uint32) {
	if lc.SupplierExists(id) {
		supplier := lc.GetSupplier(id)
		supplier.SetSafeCoordinates(lat, lon)
	}

	supplier := &model.Supplier{
		Location: &model.Location{
			Lattitude: lat,
			Longitude: lon,
		},
	}

	lc.SafeAddSupplier(id, supplier)
}

// BUG: WE MIGHT WANT TO RETURN error as well to check if it actually supplied a value, check for nil pointer.
func (lc *LogisticsController) GetSupplier(id int64) *model.Supplier {
	lc.supplierMu.RLock()
	defer lc.supplierMu.RUnlock()
	supplier, exists := lc.Suppliers[id]
	if !exists {
		return nil
	}

	return supplier
}

func (lc *LogisticsController) GetWarehouse(id int64) *model.Warehouse {
	lc.warehouseMu.RLock()
	defer lc.warehouseMu.RUnlock()
	warehouse, exists := lc.Warehouses[id]
	if !exists {
		return nil
	}

	return warehouse
}

func (lc *LogisticsController) WarehouseExists(id int64) bool {
	lc.warehouseMu.RLock()
	defer lc.warehouseMu.RUnlock()

	_, exists := lc.Warehouses[id]
	return exists
}

func (lc *LogisticsController) SupplierExists(id int64) bool {
	lc.supplierMu.RLock()
	defer lc.supplierMu.RUnlock()

	_, exists := lc.Suppliers[id]
	return exists
}

// Get all available Warehouses (Sorted) in the Logistics Controller
func (lc *LogisticsController) getAllWarehouses() []int64 {
	keys := make([]int64, 0, len(lc.Warehouses))
	for k := range lc.Warehouses {
		keys = append(keys, k)
	}

	// Quality of reading experience :D
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return keys
}

func (lc *LogisticsController) PrintWarehousesSummary() {
	var totalUnits uint64 = 0

	for _, warehouseID := range lc.getAllWarehouses() {
		warehouse := lc.GetWarehouse(warehouseID)
		totalUnits += warehouse.GetUnits()
		slog.Info("Warehouse", "ID", warehouseID, "Warehouse Units", warehouse.GetUnits(), "Number of Suppliers", len(warehouse.Suppliers))
	}

	slog.Info("Total warehouse units", "Units", totalUnits)
}

func (lc *LogisticsController) parseUnitMovement(ctx context.Context, in *pb.MoveUnitRequest) {
}
