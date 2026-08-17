package valuer

import (
	"testing"
)

func TestNewReflectValue(t *testing.T) {
	value := NewReflectValue(nil, &struct{}{})
	if value == nil {
		t.Fatal("NewReflectValue returned nil")
	}
}
