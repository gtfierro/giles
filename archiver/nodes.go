package archiver

import (
	"bytes"
	"fmt"
	"github.com/gtfierro/msgpack"
	"gopkg.in/mgo.v2/bson"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	MEAN
	COUNT
	EDGE
	NETWORK
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
	NodeLookup[MEAN] = NewMeanNode
	NodeLookup[EDGE] = NewEdgeNode
	NodeLookup[COUNT] = NewCountNode
	NodeLookup[NETWORK] = NewNetworkNode

	OpLookup = make(map[string]OperationType)
	OpLookup["window"] = WINDOW
	OpLookup["min"] = MIN
	OpLookup["max"] = MAX
	OpLookup["mean"] = MEAN
	OpLookup["edge"] = EDGE
	OpLookup["count"] = COUNT
	OpLookup["network"] = NETWORK

	opFuncChooser = make(map[string](func([][]interface{}) float64))
	opFuncChooser["mean"] = opFuncMean
	opFuncChooser["max"] = opFuncMax
	opFuncChooser["min"] = opFuncMin
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
	a      *Archiver
	dq     *dataquery
	uuids  []string
	notify chan bool
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

/** Streaming Echo Node **/
type StreamingEchoNode struct {
	send Subscriber
}

// arg0: send channel
func NewStreamingEchoNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	sen := &StreamingEchoNode{
		send: args[0].(Subscriber),
	}
	n = NewNode(sen, done)
	n.Tags["in:structure"] = LIST | TIMESERIES
	n.Tags["in:datatype"] = SCALAR | OBJECT
	n.Tags["out:structure"] = LIST | TIMESERIES
	n.Tags["out:datatype"] = SCALAR | OBJECT
	return
}

func (sen *StreamingEchoNode) Run(input interface{}) (interface{}, error) {
	log.Debug("stream ehco %v", input)
	go sen.send.Send(input)
	return nil, nil
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

// Node to subscribe to data

type SubscribeDataNode struct {
	a           *Archiver
	querystring string
	apikey      string
	dq          *dataquery
	uuids       []string
	notify      chan bool
	node        *Node
}

// arg0: archiver reference
// arg1: querystring
// arg2: apikey
// arg3: query.y dataquery struct
func NewSubscribeDataNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	sn := &SubscribeDataNode{
		a:           args[0].(*Archiver),
		querystring: args[1].(string),
		apikey:      args[2].(string),
		dq:          args[3].(*dataquery),
		notify:      make(chan bool),
	}
	sn.querystring = "select data " + strings.SplitN(sn.querystring, "data", 2)[1]
	wherestring := strings.SplitN(sn.querystring, "where", 2)[1]
	go sn.a.HandleSubscriber(sn, wherestring, sn.apikey)
	n = NewNode(sn, done)
	n.Tags["in:structure"] = LIST
	n.Tags["in:datatype"] = SCALAR | OBJECT
	n.Tags["out:structure"] = TIMESERIES
	n.Tags["out:datatype"] = SCALAR | OBJECT
	sn.node = n
	return
}

/** implement the Subscriber interface for SubscribeDataNode **/
func (sn *SubscribeDataNode) Send(msg interface{}) {
	// when we receive a new data point, re-run the data query and send it off
	response, _ := sn.a.HandleQuery(sn.querystring, sn.apikey)
	tosend := make([]SmapNumbersResponse, len(response.([]interface{})))
	for i, snr := range response.([]interface{}) {
		tosend[i] = snr.(SmapNumbersResponse)
	}
	sn.node.In <- tosend
}

func (sn *SubscribeDataNode) SendError(err error) {
	log.Error("SDN got error %v\n", err)
}

func (sn *SubscribeDataNode) GetNotify() <-chan bool {
	return sn.notify
}

func (sn *SubscribeDataNode) Run(input interface{}) (interface{}, error) {
	log.Error("SDN got Run %v\n", input)
	return input, nil
}

type ChunkedStreamingDataNode struct {
	a     *Archiver
	q     *query
	uuids []string
	node  *Node
}

// arg0: archiver reference
// arg3: query.y query struct
func NewChunkedStreamingDataNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	csn := &ChunkedStreamingDataNode{
		a: args[0].(*Archiver),
		q: args[1].(*query),
	}
	csn.uuids, _ = csn.a.GetUUIDs(csn.q.WhereBson())
	go csn.grabChunks()
	n = NewNode(csn, done)
	n.Tags["in:structure"] = LIST
	n.Tags["in:datatype"] = SCALAR | OBJECT
	n.Tags["out:structure"] = TIMESERIES
	n.Tags["out:datatype"] = SCALAR | OBJECT
	csn.node = n
	return
}

func (csn *ChunkedStreamingDataNode) grabChunks() {
	start := csn.q.data.start.UnixNano()
	end := csn.q.data.end.UnixNano()
	if start > end {
		start, end = end, start
	}
	diff := getPositiveDifference(start, end)
	fmt.Printf("Window diff is %v\n", diff)
	fmt.Printf("uuids %v\n", csn.uuids)
	for {
		fmt.Printf("fetch data in %v %v\n", uint64(start), uint64(end))
		res, err := csn.a.GetData(csn.uuids, uint64(start), uint64(end), UOT_NS, csn.q.data.timeconv)
		fmt.Printf("got result %v %v\n", res, err)
		start += diff
		end += diff
		tosend := make([]SmapNumbersResponse, len(res.([]interface{})))
		for i, snr := range res.([]interface{}) {
			tosend[i] = snr.(SmapNumbersResponse)
		}
		csn.node.In <- tosend
		time.Sleep(time.Duration(diff) * time.Nanosecond)
	}
}

func (csn *ChunkedStreamingDataNode) Run(input interface{}) (interface{}, error) {
	return input, nil
}

// The Network node takes whatever input, msgpack-encodes it, and sends it to the requested
// URI. The supported URI forms are:
//  tcp://ipaddress:port -- packet
//  udp://ipaddress:port -- packet
//  http://ipaddress:port/endpoint -- sent as body of POST request

type NetworkNode struct {
	uri      string
	url      *url.URL
	conn     net.Conn
	encoding string
}

// arg0: URI
func NewNetworkNode(done <-chan struct{}, args ...interface{}) (n *Node) {
	var (
		encoding string
		found    bool
	)

	arguments := args[0].(Dict)
	if encoding, found = arguments["encoding"].(string); !found {
		encoding = "msgpack"
	}
	nn := &NetworkNode{
		uri:      arguments["uri"].(string),
		encoding: encoding,
	}
	var err error
	nn.url, err = url.Parse(nn.uri)
	if err != nil {
		log.Panic("Invalid URI %v (%v)", nn.uri, err)
	}

	n = NewNode(nn, done)
	n.Tags["out:datatype"] = SCALAR | OBJECT
	n.Tags["out:structure"] = TIMESERIES | LIST
	n.Tags["in:datatype"] = SCALAR | OBJECT
	n.Tags["in:structure"] = TIMESERIES | LIST
	return
}

func (nn *NetworkNode) Run(input interface{}) (interface{}, error) {
	var buf *bytes.Buffer
	var mybytes = make([]byte, 1024)
	switch input.(type) {
	case []SmapNumbersResponse:
		mpfriendly := transformSmapNumResp(input.([]SmapNumbersResponse))
		length := msgpack.Encode(mpfriendly, &mybytes)
		buf = bytes.NewBuffer(mybytes[:length])
	case []*SmapItem:
		mpfriendly := transformSmapItem(input.([]*SmapItem))
		length := msgpack.Encode(mpfriendly, &mybytes)
		buf = bytes.NewBuffer(mybytes[:length])
	default:
		length := msgpack.Encode(input, &mybytes)
		buf = bytes.NewBuffer(mybytes[:length])
	}

	switch nn.url.Scheme {
	case "http":
		resp, err := http.Post(nn.uri, "application/x-msgpack", buf)
		if err == nil {
			var body []byte
			body, err = ioutil.ReadAll(resp.Body)
			_, res := msgpack.Decode(&body, 0)
			return res, err
		}
		return nil, err
	case "tcp":
		fallthrough
	case "udp":
		fallthrough
	default:
		log.Panic("Unsupported scheme %v", nn.uri)
	}
	return nil, nil
}
