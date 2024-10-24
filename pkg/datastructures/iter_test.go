package datastructures

import (
	"testing"
)

func TestMap(t *testing.T) {
	x := []int{1, 2, 4, 8, 16, 32}
	y := []int{1, 2, 4, 8, 16, 32}

	Map(x, func(el int) int {
		return el * 2
	})

	for i, el := range x {
		if el != y[i]*2 {
			t.Fail()
		}
	}
}

func TestMapCopy(t *testing.T) {
	x := []int{1, 2, 4, 8, 16, 32}

	y := MapCopy(x, func(el int) int {
		return el * 2
	})

	for i, el := range y {
		if el != x[i]*2 {
			t.Fail()
		}
		i++
	}
}

func TestReduceInt(t *testing.T) {
	x := []int{1, 2, 4, 8, 16, 32}

	total := Reduce(x, func(el int, acc int) int {
		return acc + el
	}, 0)

	if total != 63 {
		t.Fail()
	}
}

func TestReduceString(t *testing.T) {
	x := []string{"a", "b", "c"}

	total := Reduce(x, func(el string, acc string) string {
		return acc + el
	}, "")

	if total != "abc" {
		t.Fail()
	}
}

func TestListAny(t *testing.T) {
	x := []string{"a", "b", "c"}

	if !Any(x, func(el string) bool {
		return el == "c"
	}) {
		t.Fail()
	}
}

func TestListAll(t *testing.T) {
	x := []string{"a", "b", "c"}

	if All(x, func(el string) bool {
		return el == "c"
	}) {
		t.Fail()
	}
}
