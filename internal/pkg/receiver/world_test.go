package receiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogisticsController(t *testing.T) {
	wh1 := Warehouse{
		ID: 1235,
		Location: &Location{
			Lattitude: 1234,
			Longitude: 64534,
		},
	}

	assert.NotNil(t, wh1)
}
