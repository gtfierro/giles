// These functions define aggregation functions to be used inside other operators
package archiver

import (
	"math"
)

var opFuncChooser map[string](func([][]interface{}) float64)

func opFuncMean(data [][]interface{}) float64 {
	if len(data) == 0 {
		return float64(0)
	}

	var total float64

	for _, tuple := range data {
		switch tuple[1].(type) {
		case uint64:
			total += float64(tuple[1].(uint64))
		case float64:
			total += tuple[1].(float64)
		}
	}
	return total / float64(len(data))
}

func opFuncMax(data [][]interface{}) float64 {
	if len(data) == 0 {
		return float64(0)
	}

	var datamax = -1 * float64(math.MaxFloat64)

	for _, tuple := range data {
		switch tuple[1].(type) {
		case uint64:
			val := float64(tuple[1].(uint64))
			if datamax < val {
				datamax = val
			}
		case float64:
			val := tuple[1].(float64)
			if datamax < val {
				datamax = val
			}
		}
	}
	return datamax
}

func opFuncMin(data [][]interface{}) float64 {
	if len(data) == 0 {
		return float64(0)
	}

	var datamin = float64(math.MaxFloat64)

	for _, tuple := range data {
		switch tuple[1].(type) {
		case uint64:
			val := float64(tuple[1].(uint64))
			if datamin > val {
				datamin = val
			}
		case float64:
			val := tuple[1].(float64)
			if datamin > val {
				datamin = val
			}
		}
	}
	return datamin
}
