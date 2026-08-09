package reflect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterateArray(t *testing.T) {
	tests := []struct {
		name     string
		entity   any
		wantErr  error
		wantVals []any
	}{
		{
			name:     "[]int",
			entity:   [3]int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
		{
			name:     "slice",
			entity:   []int{1, 2, 3},
			wantVals: []any{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := IterateArrayOrSlice(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			assert.Equal(t, tt.wantVals, res)
		})
	}
}

func TestIterateMap(t *testing.T) {

	tests := []struct {
		name     string
		entity   any
		wantErr  error
		wantKeys []any
		wantVals []any
	}{
		{
			name: "map",
			entity: map[string]string{
				"A": "a",
				"B": "b",
				"C": "c",
			},
			wantKeys: []any{
				"A",
				"B",
				"C",
			},
			wantVals: []any{
				"a",
				"b",
				"c",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, vals, err := IterateMap(tt.entity)
			assert.Equal(t, tt.wantErr, err)
			if tt.wantErr != nil {
				return
			}
			actual := make(map[string]string)
			for i, k := range keys {
				actual[k.(string)] = vals[i].(string)
			}
			assert.Equal(t, tt.entity, actual)
		})
	}
}
