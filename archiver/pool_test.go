package archiver

import (
	"testing"
)

type Connection struct {}

func newmap(b *testing.B) map[int]*Connection {
	var x = make(map[int]*Connection)
	for i := 0; i < b.N; i++ {
		x[i] = &Connection{}
	}
	return x
}

func BenchmarkMapDelete(b *testing.B) {
	x := newmap(b)
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		delete(x, k)
	}
}

func BenchmarkMapSetNil(b *testing.B) {
	x := newmap(b)
	b.ResetTimer()
	for k := 0; k < b.N; k++ {
		x[k] = nil
	}
}
