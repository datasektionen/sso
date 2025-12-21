package handlers_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/datasektionen/sso/handlers"
)

func TestBytesTo11BitInts(t *testing.T) {
	tests := [][]byte{
		[]byte{0, 0, 0, 0},
		[]byte{1, 0, 0, 0},
		[]byte{0, 1, 0, 0},
		[]byte{0, 0, 1, 0},
		[]byte{0, 0, 0, 1},
		[]byte{96, 86, 197, 191},
		[]byte{2, 77, 168, 96, 57, 245, 6, 234, 53, 165, 188, 158, 125, 154, 205, 204},
		[]byte{52, 151, 250, 161, 237, 47, 85, 117, 72, 198, 88, 183, 27, 107, 223, 111},
		[]byte{30, 85, 103, 85, 154, 177, 244, 218, 151, 103, 61, 134, 33, 83, 200, 10},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var expected []bool
			for _, n := range tt {
				for bit := range 8 {
					expected = append(expected, n&(1<<bit) != 0)
				}
			}
			var got []bool
			for n := range handlers.BytesTo11BitInts(tt) {
				for bit := range 11 {
					got = append(got, n&(1<<bit) != 0)
				}
			}
			if len(expected)-len(got) > 11 {
				t.Errorf("BytesTo11BitInts(): lost too many bits. input bits = %v, output bits = %v", len(expected), len(got))
			}
			expected = expected[:len(got)]

			if !slices.Equal(got, expected) {
				t.Errorf("BytesTo11BitInts(): %v => %v", expected, got)
			}
		})
	}
}
