package additional

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMyStruct(t *testing.T) {

	res := NewMyStruct("001", "George", WithMyStructAddress("ShenZhen"))
	wantRes := &MyStruct{
		id:      "001",
		name:    "George",
		address: "ShenZhen",
	}
	assert.Equal(t, wantRes, res, "NewMyStruct returns the correct value")
}
