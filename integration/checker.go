package main

import (
	"encoding/json"
	"reflect"
	"strings"
)

func checkString(found, expected string) bool {
	return found == expected
}

func checkJSON(found, expected string) bool {
	dec := json.NewDecoder(strings.NewReader(found))
	dec.UseNumber()
	var foundContents interface{}
	dec.Decode(&foundContents)

	dec2 := json.NewDecoder(strings.NewReader(expected))
	dec2.UseNumber()
	var expectedContents interface{}
	dec2.Decode(&expectedContents)

	return reflect.DeepEqual(foundContents, expectedContents)
}
