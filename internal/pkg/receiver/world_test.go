package receiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogisticsControllerSupplier(t *testing.T) {
	lc := NewLogisticsController()
	ctx := context.Background()
	var (
		supplierID  int64  = 123
		warehouseId int64  = 123
		lat         uint32 = 123
		lon         uint32 = 123
	)
	err := lc.UpdateOrCreateSupplier(ctx, supplierID, lat, lon)

	assert.ErrorIs(t, err, nil)

	supp := lc.GetSupplier(supplierID)

	assert.Equal(t, supp.Location.Lattitude, lat)
	assert.Equal(t, supp.Location.Longitude, lon)

	warehouse, err := lc.GetOrCreateWarehouse(ctx, warehouseId, lat, lon)

	assert.ErrorIs(t, err, nil)
	assert.Equal(t, warehouse.Location.Lattitude, lat)
	assert.Equal(t, warehouse.Location.Longitude, lon)

}
