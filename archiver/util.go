package archiver

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

// remove trailing commas, replace all / with .
func cleantagstring(inp string) string {
	tmp := strings.TrimSuffix(inp, ",")
	tmp = strings.Replace(tmp, "/", ".", -1)
	return tmp
}

// Go doesn't provide min for uint32
func min(a, b uint32) uint32 {
	if a < b {
		return a
	} else {
		return b
	}
}

// Go doesn't provide min for uint32
func min64(a, b uint64) uint64 {
	if a < b {
		return a
	} else {
		return b
	}
}

// Calls function f and then pauses for [pause]. Loops forever.
func periodicCall(pause time.Duration, f func()) {
	for {
		f()
		time.Sleep(pause)
	}
}

// Given a forward-slash delimited path, returns a slice of prefixes, e.g.:
// input: /a/b/c/d
// output: ['/', '/a','/a/b','/a/b/c']
func getPrefixes(s string) []string {
	ret := []string{"/"}
	root := ""
	s = "/" + s
	for _, prefix := range strings.Split(s, "/") {
		if len(prefix) > 0 { //skip empty strings created by Split
			root += "/" + prefix
			ret = append(ret, root)
		}
	}
	if len(ret) > 1 {
		return ret[:len(ret)-1]
	}
	return ret
}

// returns true if the two slices are equal
func isStringSliceEqual(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	for idx := range x {
		if x[idx] != y[idx] {
			return false
		}
	}
	return true
}

// unescapes the HTML code for "="
func unescape(s string) string {
	return strings.Replace(s, "%3D", "=", -1)
}

func getPositiveDifference(x1, x2 int64) int64 {
	if x1 > x2 {
		return x1 - x2
	}
	return x2 - x1
}

func prettyPrintJSON(x interface{}) {
	bytes, _ := json.MarshalIndent(x, "", "  ")
	fmt.Println(string(bytes))
}

// Takes a dictionary that contains nested dictionaries and
// transforms it to a 1-level map with fields separated by periods k.kk.kkk = v
func flatten(m bson.M) bson.M {
	var ret = make(bson.M)
	for k, v := range m {
		if vb, ok := v.(map[string]interface{}); ok {
			for kk, vv := range flatten(vb) {
				ret[k+"."+kk] = vv
			}
		} else {
			ret[k] = v
		}
	}
	return ret
}
