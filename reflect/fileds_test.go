package reflect

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterateFields(t *testing.T) {
	type User struct {
		Name string
		age  int
	}

	testCases := []struct {
		name    string
		entity  any
		wantErr error
		wantRes map[string]any
	}{
		{
			name: "struct",
			entity: User{
				Name: "george",
				age:  20,
			},
			wantErr: nil,
			wantRes: map[string]any{
				"Name": "george",
				"age":  0,
			},
		},
		{
			name:    "basic type",
			entity:  18,
			wantErr: errors.New("entity must be a struct"),
		},
		{
			name: "multiple pointer",
			entity: func() **User {
				res := &User{
					Name: "george",
					age:  20,
				}
				return &res
			},
			wantErr: errors.New("entity must be a struct"),
		},
		{
			name:    "nil",
			entity:  nil,
			wantErr: errors.New("entity cannot be nil"),
		},
		{
			name:    "val nil",
			entity:  (*User)(nil),
			wantErr: errors.New("value cannot be zero"),
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := IterateFields(test.entity)
			assert.Equal(t, test.wantErr, err)
			if test.wantErr != nil {
				return
			}
			assert.Equal(t, test.wantRes, res)
		})
	}
}

func TestSetField(t *testing.T) {
	type User struct {
		Name string
		age  int
	}

	testCases := []struct {
		name       string
		entity     any
		field      string
		newVal     any
		wantErr    error
		wantEntity any
	}{
		{
			name: "struct",
			entity: User{
				Name: "george",
				age:  20,
			},
			field:   "Name",
			newVal:  "Jacky",
			wantErr: errors.New("value cannot be set"),
		},
		{
			name: "pointer",
			entity: &User{
				Name: "george",
				age:  20,
			},
			field:   "Name",
			newVal:  "Jacky",
			wantErr: nil,
			wantEntity: &User{
				Name: "Jacky",
				age:  20,
			},
		},
		{
			name: "private",
			entity: &User{
				Name: "george",
				age:  20,
			},
			field:   "age",
			newVal:  18,
			wantErr: errors.New("value cannot be set"),
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := SetField(test.entity, test.field, test.newVal)
			assert.Equal(t, test.wantErr, err)
			if test.wantErr != nil {
				return
			}
			assert.Equal(t, test.wantEntity, test.entity)
		})
	}

	//var i = 0
	//ptr := &i
	//reflect.ValueOf(ptr).Elem().Set(reflect.ValueOf(12))
	//assert.Equal(t, 12, i)
}
