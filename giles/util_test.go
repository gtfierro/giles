package main

import (
	"testing"
)

func TestStringSliceEqual(t *testing.T) {
	var x, y []string

	x = []string{"a", "b"}
	y = []string{"a", "b"}
	if !isStringSliceEqual(x, y) {
		t.Error(x, " should = ", y)
	}

	x = []string{"a", "b"}
	y = []string{"a", "b", "c"}
	if isStringSliceEqual(x, y) {
		t.Error(x, " should != ", y)
	}

	x = []string{"asdf", "a"}
	y = []string{"a", "asdf"}
	if isStringSliceEqual(x, y) {
		t.Error(x, " should != ", y)
	}
}

func TestGetPrefixes(t *testing.T) {
	var x string
	var y, z []string
	x = "/a/b/c"
	y = getPrefixes(x)
	z = []string{"/a", "/a/b"}
	if !isStringSliceEqual(y, z) {
		t.Error("Got ", y, " should be ", z)
	}

	x = "/a/b/c/"
	y = getPrefixes(x)
	z = []string{"/a", "/a/b"}
	if !isStringSliceEqual(y, z) {
		t.Error("Got ", y, " should be ", z)
	}

	x = "a/b/c/"
	y = getPrefixes(x)
	z = []string{"/a", "/a/b"}
	if !isStringSliceEqual(y, z) {
		t.Error("Got ", y, " should be ", z)
	}
}
