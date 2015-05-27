package archiver

import (
	"fmt"
	"github.com/gtfierro/giles/internal/tree"
	"gopkg.in/mgo.v2/bson"
	"math"
)

type NodeType uint

const (
	SCALAR NodeType = iota
	SCALAR_TS
	OBJECT
	OBJECT_TS
)

type OperationType uint

const (
	WINDOW OperationType = iota
	MIN
)

type NodeConstructor func(map[string]interface{}, ...interface{}) tree.Node

var NodeLookup map[OperationType](map[NodeType]NodeConstructor)

// Populate the NodeLookup table
func init() {
	fmt.Println("Initializing NodeLookup table...")
	NodeLookup = make(map[OperationType](map[NodeType]NodeConstructor))
	NodeLookup[MIN] = make(map[NodeType]NodeConstructor)
	NodeLookup[MIN][SCALAR_TS] = NewMinScalarNode
}

/* These nodes implement the node interface in internal/tree */

/** Where Node **/
// A WhereNode takes a where clause in its constructor.
type WhereNode struct {
	where bson.M
	store MetadataStore
	tree.BaseNode
}

// First argument are the k/v tags for this node, second are the arguments to the constructor
// arg0: BSON where clause, most likely from a parsed query
// arg1: pointer to a metadata store
func NewWhereNode(kv map[string]interface{}, args ...interface{}) (wn *WhereNode) {
	wn = &WhereNode{
		where: args[0].(bson.M),
		store: args[1].(MetadataStore),
	}
	tree.InitBaseNode(&wn.BaseNode, kv)
	return
}

// TODO: called when metadata changes. Should reevaluate where clause if necessary
func (wn *WhereNode) Input(args ...interface{}) error {
	fmt.Println("where node input")
	return nil
}

// Evaluates the where clause into a set of uuids
func (wn *WhereNode) Output() (interface{}, error) {
	return wn.store.GetUUIDs(wn.where)
}

/** Select Tags Node **/
type SelectTagsNode struct {
}

/** Select Data Node **/
type SelectDataNode struct {
	a     *Archiver
	dq    *dataquery
	uuids []string
	tree.BaseNode
}

// arg0: archiver reference
// arg1: query.y dataquery struct
func NewSelectDataNode(kv map[string]interface{}, args ...interface{}) (sn *SelectDataNode) {
	sn = &SelectDataNode{
		a:  args[0].(*Archiver),
		dq: args[1].(*dataquery),
	}
	tree.InitBaseNode(&sn.BaseNode, kv)
	return
}

// arg0: the list of UUIDs to apply the data selector to
func (sn *SelectDataNode) Input(args ...interface{}) (err error) {
	sn.uuids = args[0].([]string)
	return nil
}

func (sn *SelectDataNode) Output() (interface{}, error) {
	var err error
	// limit number of streams
	uuids := sn.uuids
	if sn.dq.limit.streamlimit > 0 && len(uuids) > 0 {
		uuids = uuids[:sn.dq.limit.streamlimit]
	}

	var response []SmapReading
	start := uint64(sn.dq.start.UnixNano())
	end := uint64(sn.dq.end.UnixNano())
	switch sn.dq.dtype {
	case IN_TYPE:
		log.Debug("Data in start %v end %v", start, end)
		if start < end {
			response, err = sn.a.GetData(uuids, start, end, UOT_NS, sn.dq.timeconv)
		} else {
			response, err = sn.a.GetData(uuids, end, start, UOT_NS, sn.dq.timeconv)
		}
	case BEFORE_TYPE:
		log.Debug("Data before time %v", start)
		response, err = sn.a.PrevData(uuids, start, int32(sn.dq.limit.limit), UOT_NS, sn.dq.timeconv)
	case AFTER_TYPE:
		log.Debug("Data after time %v", start)
		response, err = sn.a.NextData(uuids, start, int32(sn.dq.limit.limit), UOT_NS, sn.dq.timeconv)
	}
	return response, err
}

/** Min Scalar Node **/

type MinScalarNode struct {
	data []SmapReading
	tree.BaseNode
}

func NewMinScalarNode(kv map[string]interface{}, args ...interface{}) tree.Node {
	msn := &MinScalarNode{}
	kv["output"] = SCALAR
	kv["input"] = SCALAR_TS
	tree.InitBaseNode(&msn.BaseNode, kv)
	return msn
}

// arg0: list of SmapReading to compute MIN of. Must be scalars
func (msn *MinScalarNode) Input(args ...interface{}) (err error) {
	var ok bool
	msn.data, ok = args[0].([]SmapReading)
	if !ok {
		err = fmt.Errorf("Arg0 to MinScalarNode must be []SmapReading")
	}
	return
}

func (msn *MinScalarNode) Output() (interface{}, error) {
	var (
		err    error
		result = make([]interface{}, len(msn.data))
	)
	if len(msn.data) == 0 {
		err = fmt.Errorf("No data to compute min over")
		return result, err
	}
	for idx, stream := range msn.data {
		if len(stream.Readings) == 0 {
			result[idx] = nil
		}
		switch stream.Readings[0][1].(type) {
		case uint64:
			min := uint64(math.MaxUint64)
			for _, reading := range stream.Readings {
				if reading[1].(uint64) < min {
					min = reading[1].(uint64)
				}
			}
		case float64:
			min := float64(math.MaxFloat64)
			for _, reading := range stream.Readings {
				if reading[1].(float64) < min {
					min = reading[1].(float64)
				}
			}
		default:
			err = fmt.Errorf("Data type in (%v) was not uint64 or float64 (scalar)", msn.data[0])
		}
		result[idx] = min
	}

	return result, err
}
