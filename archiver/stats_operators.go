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

	log.Error("new MIN node")
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
	//for _, snr := range msn.data {
	//	for _, rdg := range snr.Readings {
	//		fmt.Printf("rdg %v\n", rdg)
	//	}
	//}
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

type MeanNode struct {
}

func NewMeanNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	mn := &MeanNode{}
	n = NewNode(mn, done)
	n.Tags["out:datatype"] = SCALAR
	n.Tags["out:structure"] = LIST
	n.Tags["in:datatype"] = SCALAR
	n.Tags["in:structure"] = TIMESERIES
	return
}

func (mn *MeanNode) Run(input interface{}) (interface{}, error) {
	data, ok := input.([]SmapNumbersResponse)
	var result = make([]*SmapItem, len(data))
	if !ok {
		return nil, fmt.Errorf("Arg0 to MaxNode must be []SmapNumbersResponse")
	}
	if len(data) == 0 {
		return result, fmt.Errorf("No data to compute max over")
	}
	for idx, stream := range data {
		total := float64(0)
		count := len(stream.Readings)
		item := &SmapItem{UUID: stream.UUID}
		if len(stream.Readings) == 0 {
			result[idx] = item
			continue
		}
		for _, reading := range stream.Readings {
			total += reading.Value
		}
		item.Data = float64(total) / float64(count)
		result[idx] = item
	}

	return result, nil
}

/** Count Node **/
type CountNode struct {
}

func NewCountNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	cn := &CountNode{}
	n = NewNode(cn, done)
	n.Tags["out:datatype"] = SCALAR
	n.Tags["out:structure"] = LIST
	n.Tags["in:datatype"] = SCALAR
	n.Tags["in:structure"] = TIMESERIES
	return
}

func (cn *CountNode) Run(input interface{}) (interface{}, error) {
	list, ok := input.([]SmapNumbersResponse)
	if !ok {
		return -1, fmt.Errorf("Input was not []SmapNumbersResponse")
	}
	var result = make([]*SmapItem, len(list))
	for idx, stream := range list {
		result[idx] = &SmapItem{UUID: stream.UUID, Data: len(stream.Readings)}
	}
	return result, nil
}
