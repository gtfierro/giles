package archiver

import (
	"bytes"
	"fmt"
	"github.com/gtfierro/msgpack"
	"gopkg.in/mgo.v2/bson"
	"io"
)

type StructureType uint

const (
	LIST StructureType = 1 << iota
	TIMESERIES
)

type DataType uint

const (
	SCALAR DataType = 1 << iota
	OBJECT
)

type OperationType uint

const (
	WINDOW OperationType = iota
	MIN
	MAX
	EDGE
)

type NodeConstructor func(<-chan struct{}, ...interface{}) *Node

var NodeLookup map[OperationType]NodeConstructor
var OpLookup map[string]OperationType

// Populate the NodeLookup table and OpLookup
func init() {
	fmt.Println("Initializing NodeLookup table...")
	NodeLookup = make(map[OperationType]NodeConstructor)
	NodeLookup[WINDOW] = NewWindowNode
	NodeLookup[MIN] = NewMinNode
	NodeLookup[MAX] = NewMaxNode
	NodeLookup[EDGE] = NewEdgeNode

	OpLookup = make(map[string]OperationType)
	OpLookup["window"] = WINDOW
	OpLookup["min"] = MIN
	OpLookup["max"] = MAX
	OpLookup["edge"] = EDGE
}

/** Where Node **/
// A WhereNode takes a where clause in its constructor.
type WhereNode struct {
	where bson.M
	store MetadataStore
}

// First argument are the k/v tags for this node, second are the arguments to the constructor
// arg0: BSON where clause, most likely from a parsed query
// arg1: pointer to a metadata store
func NewWhereNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	wn := &WhereNode{
		where: args[0].(bson.M),
		store: args[1].(MetadataStore),
	}
	n = NewNode(wn, done)
	n.Tags["out:structure"] = LIST
	n.Tags["out:datatype"] = SCALAR | OBJECT
	n.Tags["name"] = "wherenode"
	return
}

// Evaluates the where clause into a set of uuids
func (wn *WhereNode) Run(input interface{}) (interface{}, error) {
	log.Debug("running where node with %v", wn.where)
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
}

// arg0: archiver reference
// arg1: query.y dataquery struct
func NewSelectDataNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	sn := &SelectDataNode{
		a:  args[0].(*Archiver),
		dq: args[1].(*dataquery),
	}
	n = NewNode(sn, done)
	n.Tags["in:structure"] = LIST
	n.Tags["in:datatype"] = SCALAR | OBJECT
	n.Tags["out:structure"] = TIMESERIES
	n.Tags["out:datatype"] = SCALAR | OBJECT
	return
}

func (sn *SelectDataNode) Run(input interface{}) (interface{}, error) {
	var err error
	log.Debug("running select data node with %v", input)
	sn.uuids = input.([]string)
	// limit number of streams
	uuids := sn.uuids
	if sn.dq.limit.streamlimit > 0 && len(uuids) > 0 {
		uuids = uuids[:sn.dq.limit.streamlimit]
	}

	var response interface{}
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
	//TODO: make this work for objects too
	var toreturn = make([]SmapNumbersResponse, len(response.([]interface{})))
	for idx, resp := range response.([]interface{}) {
		if snr, ok := resp.(SmapNumbersResponse); ok {
			toreturn[idx] = snr
		}
	}
	return toreturn, err
}

/** Echo Node **/

type EchoNode struct {
	// writes its Input to the writer when Output() is called
	w       io.Writer
	data    *bytes.Buffer
	mybytes []byte
}

func NewEchoNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	en := &EchoNode{
		w:       args[0].(io.Writer),
		mybytes: make([]byte, 1024),
	}
	n = NewNode(en, done)
	n.Tags["in:structure"] = LIST | TIMESERIES
	n.Tags["in:datatype"] = SCALAR | OBJECT
	n.Tags["out:structure"] = LIST | TIMESERIES
	n.Tags["out:datatype"] = SCALAR | OBJECT
	return
}

// Takes the first argument and encodes it as msgpack
func (en *EchoNode) Run(input interface{}) (interface{}, error) {
	fmt.Printf("encoding %v\n", input)
	switch input.(type) {
	case []SmapNumbersResponse:
		mpfriendly := transformSmapNumResp(input.([]SmapNumbersResponse))
		length := msgpack.Encode(mpfriendly, &en.mybytes)
		en.data = bytes.NewBuffer(en.mybytes[:length])
	case []*SmapItem:
		mpfriendly := transformSmapItem(input.([]*SmapItem))
		length := msgpack.Encode(mpfriendly, &en.mybytes)
		en.data = bytes.NewBuffer(en.mybytes[:length])
	default:
		length := msgpack.Encode(input, &en.mybytes)
		en.data = bytes.NewBuffer(en.mybytes[:length])
	}
	return en.data.WriteTo(en.w)
}

// Node to pause a pipeline
type NopNode struct {
	Wait chan struct{}
}

func NewNopNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	nop := &NopNode{args[0].(chan struct{})}
	n = NewNode(nop, done)
	return
}

func (nop *NopNode) Run(input interface{}) (interface{}, error) {
	nop.Wait <- struct{}{}
	return nil, nil
}
