package receiver

import "errors"

var ErrFailedToProcessSupplier = errors.New("supplier does not exist")
var ErrFailedToAddSupplierCoordinates = errors.New("failed to add Supplier Coordinates")
var ErrWarehouseDoesNotExist = errors.New("warehouse does not exist")
var ErrSupplierDoesNotExist = errors.New("supplier does not exist")
