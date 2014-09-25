package main

import (
	"testing"
)

func newmap() map[int]*Connection {
	var x = make(map[int]*Connection)
	for i := 0; i < 1000000; i++ {
		x[i] = &Connection{}
	}
	return x
}

func BenchmarkMapDelete(b *testing.B) {
	x := newmap()
	b.ResetTimer()
	for k, _ := range x {
		delete(x, k)
	}
}

func BenchmarkMapSetNil(b *testing.B) {
	x := newmap()
	b.ResetTimer()
	for k, _ := range x {
		x[k] = nil
	}
}
