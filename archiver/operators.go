package archiver

import (
	"fmt"
	"math"
)

/** Min Node **/

type MinNode struct {
	data []SmapNumbersResponse
}

//TODO: implement min over axis
func NewMinNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	msn := &MinNode{}

	n = NewNode(msn, done)
	n.Tags["out:datatype"] = SCALAR
	n.Tags["out:structure"] = LIST
	n.Tags["in:datatype"] = SCALAR
	n.Tags["in:structure"] = TIMESERIES
	return n
}

// arg0: list of SmapNumbersResponse to compute MIN of. Must be scalars
func (msn *MinNode) Run(input interface{}) (interface{}, error) {
	var (
		err error
		ok  bool
	)
	msn.data, ok = input.([]SmapNumbersResponse)
	var result = make([]*SmapItem, len(msn.data))
	if !ok {
		err = fmt.Errorf("Arg0 to MinNode must be []SmapNumbersResponse")
	}
	if len(msn.data) == 0 {
		err = fmt.Errorf("No data to compute min over")
		return result, err
	}
	for idx, stream := range msn.data {
		item := &SmapItem{UUID: stream.UUID}
		if len(stream.Readings) == 0 {
			result[idx] = item
			continue
		}
		min := float64(math.MaxFloat64)
		for _, reading := range stream.Readings {
			if reading.Value < min {
				min = reading.Value
			}
		}
		item.Data = min
		result[idx] = item
	}

	return result, err
}

type MaxNode struct {
	data []SmapNumbersResponse
}

//** Max Node **/
//TODO: implement max over axis
func NewMaxNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	msn := &MaxNode{}

	n = NewNode(msn, done)
	n.Tags["out:datatype"] = SCALAR
	n.Tags["out:structure"] = LIST
	n.Tags["in:datatype"] = SCALAR
	n.Tags["in:structure"] = TIMESERIES
	return n
}

// arg0: list of SmapNumbersResponse to compute MIN of. Must be scalars
func (msn *MaxNode) Run(input interface{}) (interface{}, error) {
	var (
		err error
		ok  bool
	)
	msn.data, ok = input.([]SmapNumbersResponse)
	var result = make([]*SmapItem, len(msn.data))
	if !ok {
		err = fmt.Errorf("Arg0 to MaxNode must be []SmapNumbersResponse")
	}
	if len(msn.data) == 0 {
		err = fmt.Errorf("No data to compute max over")
		return result, err
	}
	for idx, stream := range msn.data {
		item := &SmapItem{UUID: stream.UUID}
		if len(stream.Readings) == 0 {
			result[idx] = item
			continue
		}
		max := float64(0)
		for _, reading := range stream.Readings {
			if reading.Value > max {
				max = reading.Value
			}
		}
		item.Data = max
		result[idx] = item
	}

	return result, err
}

// The Edge operator essentially takes the 1st order derivative of a stream
type EdgeNode struct {
	data []SmapNumbersResponse
}

func NewEdgeNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	en := &EdgeNode{}
	n = NewNode(en, done)
	n.Tags["out:datatype"] = SCALAR
	n.Tags["out:structure"] = TIMESERIES
	n.Tags["in:datatype"] = SCALAR
	n.Tags["in:structure"] = TIMESERIES
	return n
}

// arg0: list of SmapNumbersResponse to compute max of. Must be scalars
func (en *EdgeNode) Run(input interface{}) (interface{}, error) {
	var (
		ok  bool
		err error
	)
	en.data, ok = input.([]SmapNumbersResponse)
	if !ok {
		err = fmt.Errorf("Arg0 to EdgeNode must be []SmapNumbersResponse")
	}
	if len(en.data) == 0 {
		return nil, fmt.Errorf("No data to compute edge")
	}
	var result = make([]SmapNumbersResponse, len(en.data))

	for idx, stream := range en.data {
		item := SmapNumbersResponse{UUID: stream.UUID, Readings: []*SmapNumberReading{}}
		if len(stream.Readings) == 0 {
			result[idx] = item
			continue
		}
		var last float64
		for idx, reading := range stream.Readings {
			if reading.Value != last {
				toPut := &SmapNumberReading{Time: reading.Time, Value: reading.Value - last}
				if idx > 0 {
					item.Readings = append(item.Readings, toPut)
				}
				last = reading.Value
			}
		}
		result[idx] = item
	}
	return result, err
}

type WindowNode struct {
	data         []SmapNumbersResponse
	window       uint64
	aggFunc      string
	start        uint64
	end          uint64
	fromTimeUnit UnitOfTime
}

func NewWindowNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	wn := &WindowNode{}
	n = NewNode(wn, done)

	var windowSize interface{}
	var aggFunc interface{}
	if args[0] != nil {
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

	dq := args[1].(*dataquery)

	wn.start = uint64(dq.start.UnixNano())
	wn.end = uint64(dq.end.UnixNano())
	wn.fromTimeUnit = dq.timeconv

	// evaluate windowSize
	parsed, err := parseIntoDuration(windowSize.(string))
	if err != nil {
		log.Error("Could not parse window size %v (%v)", windowSize, err)
		return n
	}
	wn.window = uint64(parsed.Nanoseconds())

	log.Debug("time: %v", parsed)
	log.Debug("start: %v end %v", wn.start, wn.end)

	// evaluate aggFunc
	//TODO

	n.Tags["out:datatype"] = SCALAR
	n.Tags["out:structure"] = TIMESERIES
	n.Tags["in:datatype"] = SCALAR
	n.Tags["in:structure"] = TIMESERIES
	return n
}

//TODO: do we assume that data is sorted?
func (wn *WindowNode) Run(input interface{}) (interface{}, error) {
	var (
		ok  bool
		err error
	)
	wn.data, ok = input.([]SmapNumbersResponse)
	if !ok {
		err = fmt.Errorf("Arg0 to EdgeNode must be []SmapNumbersResponse")
	}
	if len(wn.data) == 0 {
		return nil, fmt.Errorf("No data to compute window")
	}
	var result = make([]SmapNumbersResponse, len(wn.data))
	for idx, stream := range wn.data {
		item := SmapNumbersResponse{UUID: stream.UUID, Readings: []*SmapNumberReading{}}
		if len(stream.Readings) == 0 {
			result[idx] = item
			continue
		}

		upperBound := wn.start
		lowerBound := wn.start
		lastIdx := 0
		upperBound += min64(wn.window, wn.end)
		for upperBound < wn.end {
			window := [][]interface{}{}
			for lastIdx < len(stream.Readings) {
				time := stream.Readings[lastIdx].Time
				time = convertTime(time, wn.fromTimeUnit, UOT_NS)
				if time >= lowerBound && time < upperBound {
					window = append(window, []interface{}{time, stream.Readings[lastIdx].Value})
					lastIdx += 1
				} else {
					fmt.Printf("fail: %v %v %v\n", lowerBound, time, upperBound)
					break
				}
			}
			lowerBound += wn.window
			upperBound += min64(wn.window, wn.end)
			fmt.Printf("got window %v out of total %v\n", len(window), len(stream.Readings))
			newTime := convertTime(lowerBound, UOT_NS, wn.fromTimeUnit)
			item.Readings = append(item.Readings, &SmapNumberReading{Time: newTime, Value: opFuncMean(window)})
		}
		result[idx] = item
	}

	return result, err
}
