package unsafe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnsafeAccessor_Field(t *testing.T) {

	type User struct {
		Name string
		Age  int
	}

	u := &User{Name: "george", Age: 18}
	accessor := NewUnsafeAccessor(u)
	val, err := accessor.Field("Age")
	require.NoError(t, err)
	assert.Equal(t, 18, val)

	err = accessor.SetField("Age", 20)
	require.NoError(t, err)
	assert.Equal(t, 20, u.Age)

	err = accessor.SetField("Age", "invalid")
	assert.Error(t, err)
}

func TestUnsafeAccessor_InvalidEntity(t *testing.T) {
	accessor := NewUnsafeAccessor("not a struct pointer")
	_, err := accessor.Field("Name")
	assert.Error(t, err)
}
