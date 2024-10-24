package datastructures

import (
	"context"
	"testing"
)

func TestSetIntersection(t *testing.T) {
	setA := NewSet[string]()
	setB := NewSet[string]()

	setA.Add("a")
	setB.Add("b")

	intersection := setA.Intersection(setB)
	if intersection.Len() != 0 {
		t.Fail()
	}

	setB.Add("a")

	intersection = setA.Intersection(setB)
	if intersection.Len() != 1 {
		t.Fail()
	}
}

func TestSetDifference(t *testing.T) {
	setA := NewSet[string]()
	setB := NewSet[string]()

	setA.Add("a")
	setB.Add("b")

	difference := setA.Difference(setB)
	if difference.Len() != 1 {
		t.Fail()
	}

	setB.Add("a")

	difference = setA.Difference(setB)
	if difference.Len() != 0 {
		t.Fail()
	}
}

func TestSetUnion(t *testing.T) {
	setA := NewSet[string]()
	setB := NewSet[string]()

	setA.Add("a")
	setB.Add("b")

	union := setA.Union(setB)
	if union.Len() != 2 {
		t.Fail()
	}

	setB.Add("a")

	union = setA.Union(setB)
	if union.Len() != 2 {
		t.Fail()
	}
}

func TestSetIter(t *testing.T) {
	setA := NewSet[string]("a", "a")

	for el := range setA.Iter(context.Background()) {
		if el != "a" {
			t.Fail()
		}
	}
}
