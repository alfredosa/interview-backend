package receiver

import (
	"context"
	"log/slog"
	"sort"
	"sync"

	pb "github.com/coopnorge/interview-backend/internal/generated/logistics/api/v1"
	"github.com/coopnorge/interview-backend/internal/logistics/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Just going on a hunch here, its nice if we init maps with a given length / capacity
// In a real world scenario we know an estimated amount
var (
	initialNWarehouses = 50
	initialNSuppliers  = 256
)

// LogisticsController manages access to warehouse and supplier data with concurrency control.
type LogisticsController struct {
	Warehouses map[int64]*model.Warehouse
	Suppliers  map[int64]*model.Supplier

	warehouseMu sync.RWMutex
	supplierMu  sync.RWMutex
}

func (lc *LogisticsController) MoveUnit(ctx context.Context, in *pb.MoveUnitRequest) error {
	var (
		supplierID = in.CargoUnitId
		lat        = in.Location.Latitude
		lon        = in.Location.Longitude
	)

	err := lc.UpdateOrCreateSupplier(ctx, supplierID, lat, lon)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (lc *LogisticsController) WarehouseReceivedProcessing(ctx context.Context, in *pb.UnitReachedWarehouseRequest) error {
	var (
		warehouseID = in.Announcement.WarehouseId
		supplierID  = in.Announcement.CargoUnitId
		lat         = in.Location.Latitude
		lon         = in.Location.Longitude
	)

	warehouse, err := lc.GetOrCreateWarehouse(ctx, warehouseID, lat, lon)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	warehouse.AddUnit()

	if !warehouse.SupplierExists(supplierID) {
		supplier, err := model.NewSupplier(ctx, lat, lon)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		warehouse.SafeSuppliersAdd(ctx, supplierID, supplier)
	}

	err = warehouse.SetSafeSupplierCoordinates(ctx, supplierID, lat, lon)
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), "lat", lat, "lon", lon)
		return status.Error(codes.NotFound, ErrFailedToAddSupplierCoordinates.Error())
	}

	return nil
}

func (lc *LogisticsController) GetWarehouseResponse(ctx context.Context, in *pb.GetWarehouseRequest) (*pb.GetWarehouseResponse, error) {
	var warehouseID = in.WarehouseId
	warehouse, err := lc.GetWarehouse(ctx, warehouseID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve warehouse data: %v", err)
	}

	if warehouse == nil {
		return nil, status.Error(codes.NotFound, ErrWarehouseDoesNotExist.Error())
	}

	suppliers, err := warehouse.GetProtoSuppliers(ctx)
	if err != nil {
		return nil, err
	}

	response := &pb.GetWarehouseResponse{
		Warehouse: &pb.Warehouse{
			WarehouseId: warehouseID,
			Location: &pb.Location{
				Latitude:  warehouse.Location.Lattitude,
				Longitude: warehouse.Location.Longitude,
			},
			Suppliers: suppliers,
		},
	}
	return response, nil
}

func NewLogisticsController() *LogisticsController {
	return &LogisticsController{
		Warehouses: make(map[int64]*model.Warehouse, initialNWarehouses),
		Suppliers:  make(map[int64]*model.Supplier, initialNSuppliers),
	}
}

// Get a Warehouse from the existing warehouses or add a warehouse
func (lc *LogisticsController) GetOrCreateWarehouse(ctx context.Context, id int64, lat uint32, lon uint32) (*model.Warehouse, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if lc.WarehouseExists(id) {
		warehouse, err := lc.GetWarehouse(ctx, id)
		if err != nil {
			slog.ErrorContext(ctx, "error retrieving warehouse", "error", err.Error())
		}

		if warehouse == nil {
			return nil, ErrWarehouseDoesNotExist
		}
		return warehouse, nil
	}

	warehouse := model.NewWarehouse(lat, lon)
	lc.SafeAddWarehouse(id, warehouse)
	return warehouse, nil
}

func (lc *LogisticsController) UpdateOrCreateSupplier(ctx context.Context, id int64, lat uint32, lon uint32) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if lc.SupplierExists(id) {
		supplier, err := lc.GetSupplier(ctx, id)
		if err != nil {
			return err
		}

		err = supplier.SetSafeCoordinates(ctx, lat, lon)
		if err != nil {
			return err
		}
	}

	supplier := &model.Supplier{
		Location: &model.Location{
			Lattitude: lat,
			Longitude: lon,
		},
	}

	lc.SafeAddSupplier(id, supplier)
	return nil
}

// Safely add a warehouse to the Logistics controller Map
func (lc *LogisticsController) SafeAddWarehouse(id int64, warehouse *model.Warehouse) {
	lc.warehouseMu.Lock()
	defer lc.warehouseMu.Unlock()
	lc.Warehouses[id] = warehouse
}

// Safely add a Supplier to the Logistics controller Map
func (lc *LogisticsController) SafeAddSupplier(id int64, supplier *model.Supplier) {
	lc.supplierMu.Lock()
	defer lc.supplierMu.Unlock()
	lc.Suppliers[id] = supplier
}

func (lc *LogisticsController) GetSupplier(ctx context.Context, id int64) (*model.Supplier, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	lc.supplierMu.RLock()
	defer lc.supplierMu.RUnlock()
	supplier, exists := lc.Suppliers[id]
	if !exists {
		return nil, ErrSupplierDoesNotExist
	}

	return supplier, nil
}

func (lc *LogisticsController) GetWarehouse(ctx context.Context, id int64) (*model.Warehouse, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	lc.warehouseMu.RLock()
	defer lc.warehouseMu.RUnlock()
	warehouse, exists := lc.Warehouses[id]
	if !exists {
		return nil, nil
	}

	return warehouse, nil
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
func (lc *LogisticsController) GetAllWarehouses() []int64 {
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

func (lc *LogisticsController) PrintWarehousesSummary(ctx context.Context) {
	var totalUnits uint64 = 0

	for _, warehouseID := range lc.GetAllWarehouses() {
		warehouse, err := lc.GetWarehouse(ctx, warehouseID)
		if err != nil {
			slog.ErrorContext(ctx, "error retrieving warehouse", "error", err.Error())
		}

		totalUnits += warehouse.GetUnits(ctx)
		slog.Info("Warehouse", "ID", warehouseID, "Warehouse Units", warehouse.GetUnits(ctx), "Number of Suppliers", len(warehouse.Suppliers))
	}

	slog.Info("Total # of Unique Suppliers", "Suppliers", len(lc.Suppliers))
	slog.Info("Total warehouse units", "Units", totalUnits)
}
