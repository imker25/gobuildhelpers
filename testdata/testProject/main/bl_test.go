package main

import "testing"

func TestAdd(t *testing.T) {

	res := Add(1, 2)

	if res != 3 {
		t.Errorf("Result expected to be '3', but is '%d'", res)
	}
}
