package reflect

import (
	"orm/reflect/types"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterateFunc(t *testing.T) {
	testCases := []struct {
		name    string
		entity  any
		wantErr error
		wantRes map[string]FuncInfo
	}{
		{
			name:   "struct",
			entity: types.NewUser("George", 18),
			wantRes: map[string]FuncInfo{
				"GetAge": FuncInfo{
					Name:        "GetAge",
					InputTypes:  []reflect.Type{reflect.TypeOf(types.User{})},
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{18},
				},
				//"ChangeName": FuncInfo{
				//	Name:       "ChangeName",
				//	InputTypes: []reflect.Type{reflect.TypeOf("")},
				//	// OutputTypes: []reflect.Type{},
				//	// Result: []any{},
				//},
			},
		},
		{
			name:   "pointer",
			entity: types.NewUserPtr("George", 18),
			wantRes: map[string]FuncInfo{
				"GetAge": FuncInfo{
					Name:        "GetAge",
					InputTypes:  []reflect.Type{reflect.TypeOf(&types.User{})},
					OutputTypes: []reflect.Type{reflect.TypeOf(0)},
					Result:      []any{18},
				},
				"ChangeName": FuncInfo{
					Name:        "ChangeName",
					InputTypes:  []reflect.Type{reflect.TypeOf(&types.User{}), reflect.TypeOf("")},
					OutputTypes: []reflect.Type{},
					Result:      []any{},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateFunc(tc.entity)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
