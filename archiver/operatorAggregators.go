// These functions define aggregation functions to be used inside other operators
package archiver

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
