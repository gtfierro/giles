/*
Package rdb is a generated protocol buffer package.

It is generated from these files:
	rdb.proto

It has these top-level messages:
	Reading
	ReadingSet
	DatabaseDelta
	DatabaseRecord
	Query
	Nearest
	Delete
	Response
*/

package readingdbproto

import proto "github.com/golang/protobuf/proto"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type MessageType int32

const (
	MessageType_QUERY      MessageType = 1
	MessageType_READINGSET MessageType = 2
	MessageType_RESPONSE   MessageType = 3
	MessageType_NEAREST    MessageType = 4
	MessageType_DELETE     MessageType = 5
)

var MessageType_name = map[int32]string{
	1: "QUERY",
	2: "READINGSET",
	3: "RESPONSE",
	4: "NEAREST",
	5: "DELETE",
}
var MessageType_value = map[string]int32{
	"QUERY":      1,
	"READINGSET": 2,
	"RESPONSE":   3,
	"NEAREST":    4,
	"DELETE":     5,
}

func (x MessageType) Enum() *MessageType {
	p := new(MessageType)
	*p = x
	return p
}
func (x MessageType) String() string {
	return proto.EnumName(MessageType_name, int32(x))
}
func (x *MessageType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(MessageType_value, data, "MessageType")
	if err != nil {
		return err
	}
	*x = MessageType(value)
	return nil
}

type Nearest_Direction int32

const (
	Nearest_NEXT Nearest_Direction = 1
	Nearest_PREV Nearest_Direction = 2
)

var Nearest_Direction_name = map[int32]string{
	1: "NEXT",
	2: "PREV",
}
var Nearest_Direction_value = map[string]int32{
	"NEXT": 1,
	"PREV": 2,
}

func (x Nearest_Direction) Enum() *Nearest_Direction {
	p := new(Nearest_Direction)
	*p = x
	return p
}
func (x Nearest_Direction) String() string {
	return proto.EnumName(Nearest_Direction_name, int32(x))
}
func (x *Nearest_Direction) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(Nearest_Direction_value, data, "Nearest_Direction")
	if err != nil {
		return err
	}
	*x = Nearest_Direction(value)
	return nil
}

type Response_ErrorCode int32

const (
	Response_OK         Response_ErrorCode = 1
	Response_FAIL       Response_ErrorCode = 2
	Response_FAIL_PARAM Response_ErrorCode = 3
	Response_FAIL_MEM   Response_ErrorCode = 4
)

var Response_ErrorCode_name = map[int32]string{
	1: "OK",
	2: "FAIL",
	3: "FAIL_PARAM",
	4: "FAIL_MEM",
}
var Response_ErrorCode_value = map[string]int32{
	"OK":         1,
	"FAIL":       2,
	"FAIL_PARAM": 3,
	"FAIL_MEM":   4,
}

func (x Response_ErrorCode) Enum() *Response_ErrorCode {
	p := new(Response_ErrorCode)
	*p = x
	return p
}
func (x Response_ErrorCode) String() string {
	return proto.EnumName(Response_ErrorCode_name, int32(x))
}
func (x *Response_ErrorCode) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(Response_ErrorCode_value, data, "Response_ErrorCode")
	if err != nil {
		return err
	}
	*x = Response_ErrorCode(value)
	return nil
}

type Reading struct {
	Timestamp        *uint64  `protobuf:"varint,1,req,name=timestamp" json:"timestamp,omitempty"`
	Value            *float64 `protobuf:"fixed64,2,req,name=value" json:"value,omitempty"`
	Seqno            *uint64  `protobuf:"varint,3,opt,name=seqno" json:"seqno,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *Reading) Reset()         { *m = Reading{} }
func (m *Reading) String() string { return proto.CompactTextString(m) }
func (*Reading) ProtoMessage()    {}

func (m *Reading) GetTimestamp() uint64 {
	if m != nil && m.Timestamp != nil {
		return *m.Timestamp
	}
	return 0
}

func (m *Reading) GetValue() float64 {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return 0
}

func (m *Reading) GetSeqno() uint64 {
	if m != nil && m.Seqno != nil {
		return *m.Seqno
	}
	return 0
}

type ReadingSet struct {
	Streamid         *uint32    `protobuf:"varint,1,req,name=streamid" json:"streamid,omitempty"`
	Substream        *uint32    `protobuf:"varint,2,req,name=substream" json:"substream,omitempty"`
	Data             []*Reading `protobuf:"bytes,3,rep,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte     `json:"-"`
}

func (m *ReadingSet) Reset()         { *m = ReadingSet{} }
func (m *ReadingSet) String() string { return proto.CompactTextString(m) }
func (*ReadingSet) ProtoMessage()    {}

func (m *ReadingSet) GetStreamid() uint32 {
	if m != nil && m.Streamid != nil {
		return *m.Streamid
	}
	return 0
}

func (m *ReadingSet) GetSubstream() uint32 {
	if m != nil && m.Substream != nil {
		return *m.Substream
	}
	return 0
}

func (m *ReadingSet) GetData() []*Reading {
	if m != nil {
		return m.Data
	}
	return nil
}

type DatabaseDelta struct {
	Timestamp        *int64 `protobuf:"varint,1,opt,name=timestamp" json:"timestamp,omitempty"`
	Value            *int64 `protobuf:"varint,2,opt,name=value" json:"value,omitempty"`
	Seqno            *int64 `protobuf:"varint,3,opt,name=seqno" json:"seqno,omitempty"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *DatabaseDelta) Reset()         { *m = DatabaseDelta{} }
func (m *DatabaseDelta) String() string { return proto.CompactTextString(m) }
func (*DatabaseDelta) ProtoMessage()    {}

func (m *DatabaseDelta) GetTimestamp() int64 {
	if m != nil && m.Timestamp != nil {
		return *m.Timestamp
	}
	return 0
}

func (m *DatabaseDelta) GetValue() int64 {
	if m != nil && m.Value != nil {
		return *m.Value
	}
	return 0
}

func (m *DatabaseDelta) GetSeqno() int64 {
	if m != nil && m.Seqno != nil {
		return *m.Seqno
	}
	return 0
}

type DatabaseRecord struct {
	PeriodLength     *uint32          `protobuf:"varint,1,req,name=period_length" json:"period_length,omitempty"`
	First            *Reading         `protobuf:"bytes,2,opt,name=first" json:"first,omitempty"`
	Deltas           []*DatabaseDelta `protobuf:"bytes,3,rep,name=deltas" json:"deltas,omitempty"`
	XXX_unrecognized []byte           `json:"-"`
}

func (m *DatabaseRecord) Reset()         { *m = DatabaseRecord{} }
func (m *DatabaseRecord) String() string { return proto.CompactTextString(m) }
func (*DatabaseRecord) ProtoMessage()    {}

func (m *DatabaseRecord) GetPeriodLength() uint32 {
	if m != nil && m.PeriodLength != nil {
		return *m.PeriodLength
	}
	return 0
}

func (m *DatabaseRecord) GetFirst() *Reading {
	if m != nil {
		return m.First
	}
	return nil
}

func (m *DatabaseRecord) GetDeltas() []*DatabaseDelta {
	if m != nil {
		return m.Deltas
	}
	return nil
}

type Query struct {
	Streamid         *uint32 `protobuf:"varint,1,req,name=streamid" json:"streamid,omitempty"`
	Substream        *uint32 `protobuf:"varint,2,req,name=substream" json:"substream,omitempty"`
	Starttime        *uint64 `protobuf:"varint,3,req,name=starttime" json:"starttime,omitempty"`
	Endtime          *uint64 `protobuf:"varint,4,req,name=endtime" json:"endtime,omitempty"`
	Action           *uint32 `protobuf:"varint,5,opt,name=action" json:"action,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Query) Reset()         { *m = Query{} }
func (m *Query) String() string { return proto.CompactTextString(m) }
func (*Query) ProtoMessage()    {}

func (m *Query) GetStreamid() uint32 {
	if m != nil && m.Streamid != nil {
		return *m.Streamid
	}
	return 0
}

func (m *Query) GetSubstream() uint32 {
	if m != nil && m.Substream != nil {
		return *m.Substream
	}
	return 0
}

func (m *Query) GetStarttime() uint64 {
	if m != nil && m.Starttime != nil {
		return *m.Starttime
	}
	return 0
}

func (m *Query) GetEndtime() uint64 {
	if m != nil && m.Endtime != nil {
		return *m.Endtime
	}
	return 0
}

func (m *Query) GetAction() uint32 {
	if m != nil && m.Action != nil {
		return *m.Action
	}
	return 0
}

type Nearest struct {
	Streamid         *uint32            `protobuf:"varint,1,req,name=streamid" json:"streamid,omitempty"`
	Substream        *uint32            `protobuf:"varint,2,req,name=substream" json:"substream,omitempty"`
	Reference        *uint64            `protobuf:"varint,3,req,name=reference" json:"reference,omitempty"`
	Direction        *Nearest_Direction `protobuf:"varint,4,req,name=direction,enum=Nearest_Direction" json:"direction,omitempty"`
	N                *uint32            `protobuf:"varint,5,opt,name=n" json:"n,omitempty"`
	XXX_unrecognized []byte             `json:"-"`
}

func (m *Nearest) Reset()         { *m = Nearest{} }
func (m *Nearest) String() string { return proto.CompactTextString(m) }
func (*Nearest) ProtoMessage()    {}

func (m *Nearest) GetStreamid() uint32 {
	if m != nil && m.Streamid != nil {
		return *m.Streamid
	}
	return 0
}

func (m *Nearest) GetSubstream() uint32 {
	if m != nil && m.Substream != nil {
		return *m.Substream
	}
	return 0
}

func (m *Nearest) GetReference() uint64 {
	if m != nil && m.Reference != nil {
		return *m.Reference
	}
	return 0
}

func (m *Nearest) GetDirection() Nearest_Direction {
	if m != nil && m.Direction != nil {
		return *m.Direction
	}
	return Nearest_NEXT
}

func (m *Nearest) GetN() uint32 {
	if m != nil && m.N != nil {
		return *m.N
	}
	return 0
}

type Delete struct {
	Streamid         *uint32 `protobuf:"varint,1,req,name=streamid" json:"streamid,omitempty"`
	Substream        *uint32 `protobuf:"varint,2,req,name=substream" json:"substream,omitempty"`
	Starttime        *uint64 `protobuf:"varint,3,req,name=starttime" json:"starttime,omitempty"`
	Endtime          *uint64 `protobuf:"varint,4,req,name=endtime" json:"endtime,omitempty"`
	XXX_unrecognized []byte  `json:"-"`
}

func (m *Delete) Reset()         { *m = Delete{} }
func (m *Delete) String() string { return proto.CompactTextString(m) }
func (*Delete) ProtoMessage()    {}

func (m *Delete) GetStreamid() uint32 {
	if m != nil && m.Streamid != nil {
		return *m.Streamid
	}
	return 0
}

func (m *Delete) GetSubstream() uint32 {
	if m != nil && m.Substream != nil {
		return *m.Substream
	}
	return 0
}

func (m *Delete) GetStarttime() uint64 {
	if m != nil && m.Starttime != nil {
		return *m.Starttime
	}
	return 0
}

func (m *Delete) GetEndtime() uint64 {
	if m != nil && m.Endtime != nil {
		return *m.Endtime
	}
	return 0
}

type Response struct {
	Error            *Response_ErrorCode `protobuf:"varint,1,req,name=error,enum=Response_ErrorCode" json:"error,omitempty"`
	Data             *ReadingSet         `protobuf:"bytes,2,opt,name=data" json:"data,omitempty"`
	XXX_unrecognized []byte              `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}

func (m *Response) GetError() Response_ErrorCode {
	if m != nil && m.Error != nil {
		return *m.Error
	}
	return Response_OK
}

func (m *Response) GetData() *ReadingSet {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterEnum("MessageType", MessageType_name, MessageType_value)
	proto.RegisterEnum("Nearest_Direction", Nearest_Direction_name, Nearest_Direction_value)
	proto.RegisterEnum("Response_ErrorCode", Response_ErrorCode_name, Response_ErrorCode_value)
}
