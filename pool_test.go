package pool

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPoolPowerOfTwoRange(t *testing.T) {
	for i, test := range []struct {
		min, max int
		exp      []int
	}{
		{0, 8, []int{1, 2, 4, 8}},
		{0, 7, []int{1, 2, 4}},
		{0, 9, []int{1, 2, 4, 8}},
		{3, 8, []int{4, 8}},
		{1, 7, []int{1, 2, 4}},
		{1, 9, []int{1, 2, 4, 8}},
	} {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			var act []int
			PowerOfTwoRange(test.min, test.max, func(n int) {
				act = append(act, n)
			})
			if !reflect.DeepEqual(act, test.exp) {
				t.Errorf("unexpected range from %d to %d: %v; want %v", test.min, test.max, act, test.exp)
			}
		})
	}
}
