package archiver

import (
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

// Calls function f and then pauses for [pause]. Loops forever.
func periodicCall(pause time.Duration, f func()) {
	for {
		f()
		time.Sleep(pause)
	}
}

// Given a forward-slash delimited path, returns a slice of prefixes, e.g.:
// input: /a/b/c/d
// output: ['/a','/a/b','/a/b/c']
func getPrefixes(s string) []string {
	ret := []string{}
	root := ""
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
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

// Takes a timestamp with accompanying unit of time 'stream_uot' and
// converts it to the unit of time 'target_uot'
func convertTime(time uint64, stream_uot, target_uot UnitOfTime) uint64 {
	unitmultiplier := map[UnitOfTime]uint64{
		UOT_NS: 1000000000,
		UOT_US: 1000000,
		UOT_MS: 1000,
		UOT_S:  1}
	return time / unitmultiplier[stream_uot] * unitmultiplier[target_uot]
}
