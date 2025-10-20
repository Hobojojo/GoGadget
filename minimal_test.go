package main

import (
	"testing"
)

func TestMinimal(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math is broken")
	}
}