package unsafe

import (
	"testing"
)

func TestPrintFieldOffset(t *testing.T) {
	testCases := []struct {
		name   string
		entity any
	}{
		{
			name:   "User",
			entity: User{},
		},
		{
			name:   "UserV1",
			entity: UserV1{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			PrintFieldOffset(tc.entity)
		})
	}
}

type User struct {
	Name    string
	Age     int32
	Alias   []string
	Address string
}

type UserV1 struct {
	Name    string
	Age     int32
	Num     int32
	Alias   []string
	Address string
}
