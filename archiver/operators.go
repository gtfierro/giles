package archiver

import (
	"fmt"
	"github.com/gtfierro/giles/internal/tree"
	"math"
)

/** Min Node **/

type MinNode struct {
	data []SmapReading
	tree.BaseNode
}

//TODO: implement min over axis
func NewMinNode(args ...interface{}) tree.Node {
	msn := &MinNode{}
	tree.InitBaseNode(&msn.BaseNode)

	msn.BaseNode.Set("out:datatype", SCALAR)
	msn.BaseNode.Set("out:structure", LIST)
	msn.BaseNode.Set("in:datatype", SCALAR)
	msn.BaseNode.Set("in:structure", TIMESERIES)
	return msn
}

// arg0: list of SmapReading to compute MIN of. Must be scalars
func (msn *MinNode) Input(args ...interface{}) (err error) {
	var ok bool
	msn.data, ok = args[0].([]SmapReading)
	if !ok {
		err = fmt.Errorf("Arg0 to MinNode must be []SmapReading")
	}
	return
}

func (msn *MinNode) Output() (interface{}, error) {
	var (
		err    error
		result = make([]*ListItem, len(msn.data))
	)
	if len(msn.data) == 0 {
		err = fmt.Errorf("No data to compute min over")
		return result, err
	}
	for idx, stream := range msn.data {
		item := &ListItem{UUID: stream.UUID}
		if len(stream.Readings) == 0 {
			result[idx] = item
			continue
		}
		switch stream.Readings[0][1].(type) {
		case uint64:
			min := uint64(math.MaxUint64)
			for _, reading := range stream.Readings {
				if reading[1].(uint64) < min {
					min = reading[1].(uint64)
				}
			}
			item.Data = min
		case float64:
			min := float64(math.MaxFloat64)
			for _, reading := range stream.Readings {
				if reading[1].(float64) < min {
					min = reading[1].(float64)
				}
			}
			item.Data = min
		default:
			err = fmt.Errorf("Data type in (%v) was not uint64 or float64 (scalar)", msn.data[0])
		}
		result[idx] = item
	}

	return result, err
}

/** Max Node **/

type MaxNode struct {
	data []SmapReading
	tree.BaseNode
}

//TODO: implement max over axis
func NewMaxNode(args ...interface{}) tree.Node {
	msn := &MaxNode{}
	tree.InitBaseNode(&msn.BaseNode)

	msn.BaseNode.Set("out:datatype", SCALAR)
	msn.BaseNode.Set("out:structure", LIST)
	msn.BaseNode.Set("in:datatype", SCALAR)
	msn.BaseNode.Set("in:structure", TIMESERIES)
	return msn
}

// arg0: list of SmapReading to compute max of. Must be scalars
func (msn *MaxNode) Input(args ...interface{}) (err error) {
	var ok bool
	msn.data, ok = args[0].([]SmapReading)
	if !ok {
		err = fmt.Errorf("Arg0 to MaxNode must be []SmapReading")
	}
	return
}

func (msn *MaxNode) Output() (interface{}, error) {
	var (
		err    error
		result = make([]*ListItem, len(msn.data))
	)
	if len(msn.data) == 0 {
		err = fmt.Errorf("No data to compute max over")
		return result, err
	}
	for idx, stream := range msn.data {
		item := &ListItem{UUID: stream.UUID}
		if len(stream.Readings) == 0 {
			result[idx] = item
			continue
		}
		switch stream.Readings[0][1].(type) {
		case uint64:
			max := uint64(0)
			for _, reading := range stream.Readings {
				if reading[1].(uint64) < max {
					max = reading[1].(uint64)
				}
			}
			item.Data = max
		case float64:
			max := float64(0)
			for _, reading := range stream.Readings {
				if reading[1].(float64) < max {
					max = reading[1].(float64)
				}
			}
			item.Data = max
		default:
			err = fmt.Errorf("Data type in (%v) was not uint64 or float64 (scalar)", msn.data[0])
		}
		result[idx] = item
	}

	return result, err
}

// The Edge operator essentially takes the 1st order derivative of a stream
type EdgeNode struct {
	data []SmapReading
	tree.BaseNode
}

func NewEdgeNode(args ...interface{}) tree.Node {
	en := &EdgeNode{}
	tree.InitBaseNode(&en.BaseNode)

	en.BaseNode.Set("out:datatype", SCALAR)
	en.BaseNode.Set("out:structure", TIMESERIES)
	en.BaseNode.Set("in:datatype", SCALAR)
	en.BaseNode.Set("in:structure", TIMESERIES)
	return en
}

// arg0: list of SmapReading to compute max of. Must be scalars
func (en *EdgeNode) Input(args ...interface{}) (err error) {
	var ok bool
	en.data, ok = args[0].([]SmapReading)
	if !ok {
		err = fmt.Errorf("Arg0 to EdgeNode must be []SmapReading")
	}
	return
}

func (en *EdgeNode) Output() (interface{}, error) {
	if len(en.data) == 0 {
		return nil, fmt.Errorf("No data to compute edge")
	}
	var result = make([]SmapReading, len(en.data))

	for idx, stream := range en.data {
		item := SmapReading{UUID: stream.UUID, Readings: [][]interface{}{}}
		if len(stream.Readings) == 0 {
			result[idx] = item
			continue
		}
		switch stream.Readings[0][1].(type) {
		case uint64:
			var last uint64
			for _, reading := range stream.Readings {
				if reading[1].(uint64) != last {
					toPut := []interface{}{reading[0], reading[1].(uint64) - last}
					item.Readings = append(item.Readings, toPut)
					last = reading[1].(uint64)
				}
			}
		case float64:
			var last float64
			for _, reading := range stream.Readings {
				if reading[1].(float64) != last {
					toPut := []interface{}{reading[0], reading[1].(float64) - last}
					item.Readings = append(item.Readings, toPut)
					last = reading[1].(float64)
				}
			}
		default:
			return nil, fmt.Errorf("Data type in (%v) was not uint64 or float64 (scalar)", en.data[0])
		}
		result[idx] = item
	}
	return result, nil
}

type WindowNode struct {
	data    []SmapReading
	window  uint64
	aggFunc string
	tree.BaseNode
}

func NewWindowNode(args ...interface{}) tree.Node {
	wn := &WindowNode{}
	tree.InitBaseNode(&wn.BaseNode)

	var windowSize interface{}
	var aggFunc interface{}
	if len(args) > 0 {
		kv := args[0].(Dict)
		windowSize = kv["size"]
		aggFunc = kv["func"]
	}

	if windowSize == nil {
		windowSize = "5min"
	}

	if aggFunc == nil {
		aggFunc = "mean"
	}

	// evaluate windowSize
	parsed, err := parseIntoDuration(windowSize.(string))
	if err != nil {
		log.Error("Could not parse window size %v (%v)", windowSize, err)
		return wn
	}
	fmt.Printf("time: %v\n", parsed)

	// evaluate aggFunc

	wn.BaseNode.Set("out:datatype", SCALAR)
	wn.BaseNode.Set("out:structure", TIMESERIES)
	wn.BaseNode.Set("in:datatype", SCALAR)
	wn.BaseNode.Set("in:structure", TIMESERIES)
	return wn
}

func (wn *WindowNode) Input(args ...interface{}) (err error) {
	var ok bool
	wn.data, ok = args[0].([]SmapReading)
	if !ok {
		err = fmt.Errorf("Arg0 to EdgeNode must be []SmapReading")
	}
	return
}

func (wn *WindowNode) Output() (interface{}, error) {
	if len(wn.data) == 0 {
		return nil, fmt.Errorf("No data to compute window")
	}
	var result = make([]SmapReading, len(wn.data))
	for idx, stream := range wn.data {
		item := SmapReading{UUID: stream.UUID, Readings: [][]interface{}{}}
		if len(stream.Readings) == 0 {
			result[idx] = item
			continue
		}

		result[idx] = item

	}

	return result, nil
}
