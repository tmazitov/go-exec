package main

import (
	"reflect"
	"slice_intersection/solutions"
	"testing"
)

func TestIntersection(t *testing.T) {
	tests := []struct {
		name string
		x    []int
		y    []int
		want []int
	}{
		{
			name: "both nil",
			x:    nil,
			y:    nil,
			want: []int{},
		},
		{
			name: "one nil",
			x:    []int{1, 2, 3},
			y:    nil,
			want: []int{},
		},
		{	
			name: "empty slices",
			x:    []int{},
			y:    []int{},
			want: []int{},
		},
		{
			name: "no intersection",
			x:    []int{1, 2, 3},
			y:    []int{4, 5, 6},
			want: []int{},
		},
		{
			name: "partial intersection",
			x:    []int{1, 2, 3, 4},
			y:    []int{3, 4, 5, 6},
			want: []int{3, 4},
		},
		{
			name: "full intersection",
			x:    []int{1, 2, 3},
			y:    []int{1, 2, 3},
			want: []int{1, 2, 3},
		},
		{
			name: "repeated elements in x",
			x:    []int{1, 2, 2, 3},
			y:    []int{2, 2, 3},
			want: []int{2, 2, 3},
		},
		{
			name: "repeated elements in y",
			x:    []int{1, 2, 3},
			y:    []int{2, 2, 3, 3},
			want: []int{2, 3},
		},
		{
			name: "different order",
			x:    []int{4, 1, 3, 2},
			y:    []int{2, 3, 4, 5},
			want: []int{4, 3, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := solutions.Intersection(tt.x, tt.y)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Intersection(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}
