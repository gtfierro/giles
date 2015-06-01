package archiver

import (
	"bytes"
	"fmt"
	"github.com/gtfierro/giles/internal/tree"
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

type NodeConstructor func(...interface{}) tree.Node

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
func NewWhereNode(args ...interface{}) (wn *WhereNode) {
	wn = &WhereNode{
		where: args[0].(bson.M),
		store: args[1].(MetadataStore),
	}
	tree.InitBaseNode(&wn.BaseNode)
	wn.BaseNode.Set("out:structure", LIST)
	wn.BaseNode.Set("out:datatype", SCALAR|OBJECT)
	wn.BaseNode.Set("name", "wherenode")
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
func NewSelectDataNode(args ...interface{}) (sn *SelectDataNode) {
	sn = &SelectDataNode{
		a:  args[0].(*Archiver),
		dq: args[1].(*dataquery),
	}
	tree.InitBaseNode(&sn.BaseNode)
	sn.BaseNode.Set("in:structure", LIST)
	sn.BaseNode.Set("in:datatype", SCALAR|OBJECT)
	sn.BaseNode.Set("out:structure", TIMESERIES)
	sn.BaseNode.Set("out:datatype", SCALAR|OBJECT)
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
	tree.BaseNode
}

func NewEchoNode(args ...interface{}) tree.Node {
	en := &EchoNode{
		w:       args[0].(io.Writer),
		mybytes: make([]byte, 1024),
	}
	tree.InitBaseNode(&en.BaseNode)
	en.BaseNode.Set("in:structure", LIST|TIMESERIES)
	en.BaseNode.Set("in:datatype", SCALAR|OBJECT)
	en.BaseNode.Set("out:structure", LIST|TIMESERIES)
	en.BaseNode.Set("out:datatype", SCALAR|OBJECT)
	return en
}

// Takes the first argument and encodes it as msgpack
func (en *EchoNode) Input(args ...interface{}) (err error) {
	fmt.Printf("encoding %v\n", args[0])
	switch args[0].(type) {
	case []SmapNumbersResponse:
		mpfriendly := transformSmapNumResp(args[0].([]SmapNumbersResponse))
		length := msgpack.Encode(mpfriendly, &en.mybytes)
		en.data = bytes.NewBuffer(en.mybytes[:length])
	case []*SmapItem:
		mpfriendly := transformSmapItem(args[0].([]*SmapItem))
		length := msgpack.Encode(mpfriendly, &en.mybytes)
		en.data = bytes.NewBuffer(en.mybytes[:length])
	default:
		length := msgpack.Encode(args[0], &en.mybytes)
		en.data = bytes.NewBuffer(en.mybytes[:length])
	}
	return nil
}

func (en *EchoNode) Output() (interface{}, error) {
	log.Debug("EchoNode writing out %v", en.data.Len())
	return en.data.WriteTo(en.w)
}
