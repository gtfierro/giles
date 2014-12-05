package cphandler

// AUTO GENERATED - DO NOT EDIT

import (
	C "github.com/glycerine/go-capnproto"
	"math"
	"unsafe"
)

type Request C.Struct
type Request_Which uint16

const (
	REQUEST_VOID      Request_Which = 0
	REQUEST_WRITEDATA Request_Which = 1
)

func NewRequest(s *C.Segment) Request      { return Request(s.NewStruct(8, 1)) }
func NewRootRequest(s *C.Segment) Request  { return Request(s.NewRootStruct(8, 1)) }
func AutoNewRequest(s *C.Segment) Request  { return Request(s.NewStructAR(8, 1)) }
func ReadRootRequest(s *C.Segment) Request { return Request(s.Root(0).ToStruct()) }
func (s Request) Which() Request_Which     { return Request_Which(C.Struct(s).Get16(0)) }
func (s Request) SetVoid()                 { C.Struct(s).Set16(0, 0) }
func (s Request) WriteData() ReqWriteData  { return ReqWriteData(C.Struct(s).GetObject(0).ToStruct()) }
func (s Request) SetWriteData(v ReqWriteData) {
	C.Struct(s).Set16(0, 1)
	C.Struct(s).SetObject(0, C.Object(v))
}

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s Request) MarshalJSON() (bs []byte, err error) { return }

type Request_List C.PointerList

func NewRequestList(s *C.Segment, sz int) Request_List {
	return Request_List(s.NewCompositeList(8, 1, sz))
}
func (s Request_List) Len() int         { return C.PointerList(s).Len() }
func (s Request_List) At(i int) Request { return Request(C.PointerList(s).At(i).ToStruct()) }
func (s Request_List) ToArray() []Request {
	return *(*[]Request)(unsafe.Pointer(C.PointerList(s).ToArray()))
}
func (s Request_List) Set(i int, item Request) { C.PointerList(s).Set(i, C.Object(item)) }

type ReqWriteData C.Struct

func NewReqWriteData(s *C.Segment) ReqWriteData       { return ReqWriteData(s.NewStruct(0, 1)) }
func NewRootReqWriteData(s *C.Segment) ReqWriteData   { return ReqWriteData(s.NewRootStruct(0, 1)) }
func AutoNewReqWriteData(s *C.Segment) ReqWriteData   { return ReqWriteData(s.NewStructAR(0, 1)) }
func ReadRootReqWriteData(s *C.Segment) ReqWriteData  { return ReqWriteData(s.Root(0).ToStruct()) }
func (s ReqWriteData) Messages() SmapMessage_List     { return SmapMessage_List(C.Struct(s).GetObject(0)) }
func (s ReqWriteData) SetMessages(v SmapMessage_List) { C.Struct(s).SetObject(0, C.Object(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s ReqWriteData) MarshalJSON() (bs []byte, err error) { return }

type ReqWriteData_List C.PointerList

func NewReqWriteDataList(s *C.Segment, sz int) ReqWriteData_List {
	return ReqWriteData_List(s.NewCompositeList(0, 1, sz))
}
func (s ReqWriteData_List) Len() int { return C.PointerList(s).Len() }
func (s ReqWriteData_List) At(i int) ReqWriteData {
	return ReqWriteData(C.PointerList(s).At(i).ToStruct())
}
func (s ReqWriteData_List) ToArray() []ReqWriteData {
	return *(*[]ReqWriteData)(unsafe.Pointer(C.PointerList(s).ToArray()))
}
func (s ReqWriteData_List) Set(i int, item ReqWriteData) { C.PointerList(s).Set(i, C.Object(item)) }

type Response C.Struct

func NewResponse(s *C.Segment) Response           { return Response(s.NewStruct(8, 1)) }
func NewRootResponse(s *C.Segment) Response       { return Response(s.NewRootStruct(8, 1)) }
func AutoNewResponse(s *C.Segment) Response       { return Response(s.NewStructAR(8, 1)) }
func ReadRootResponse(s *C.Segment) Response      { return Response(s.Root(0).ToStruct()) }
func (s Response) Status() StatusCode             { return StatusCode(C.Struct(s).Get16(0)) }
func (s Response) SetStatus(v StatusCode)         { C.Struct(s).Set16(0, uint16(v)) }
func (s Response) Messages() SmapMessage_List     { return SmapMessage_List(C.Struct(s).GetObject(0)) }
func (s Response) SetMessages(v SmapMessage_List) { C.Struct(s).SetObject(0, C.Object(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s Response) MarshalJSON() (bs []byte, err error) { return }

type Response_List C.PointerList

func NewResponseList(s *C.Segment, sz int) Response_List {
	return Response_List(s.NewCompositeList(8, 1, sz))
}
func (s Response_List) Len() int          { return C.PointerList(s).Len() }
func (s Response_List) At(i int) Response { return Response(C.PointerList(s).At(i).ToStruct()) }
func (s Response_List) ToArray() []Response {
	return *(*[]Response)(unsafe.Pointer(C.PointerList(s).ToArray()))
}
func (s Response_List) Set(i int, item Response) { C.PointerList(s).Set(i, C.Object(item)) }

type StatusCode uint16

const (
	STATUSCODE_OK            StatusCode = 0
	STATUSCODE_INTERNALERROR StatusCode = 1
)

func (c StatusCode) String() string {
	switch c {
	case STATUSCODE_OK:
		return "ok"
	case STATUSCODE_INTERNALERROR:
		return "internalError"
	default:
		return ""
	}
}

func StatusCodeFromString(c string) StatusCode {
	switch c {
	case "ok":
		return STATUSCODE_OK
	case "internalError":
		return STATUSCODE_INTERNALERROR
	default:
		return 0
	}
}

type StatusCode_List C.PointerList

func NewStatusCodeList(s *C.Segment, sz int) StatusCode_List {
	return StatusCode_List(s.NewUInt16List(sz))
}
func (s StatusCode_List) Len() int            { return C.UInt16List(s).Len() }
func (s StatusCode_List) At(i int) StatusCode { return StatusCode(C.UInt16List(s).At(i)) }
func (s StatusCode_List) ToArray() []StatusCode {
	return *(*[]StatusCode)(unsafe.Pointer(C.UInt16List(s).ToEnumArray()))
}

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s StatusCode) MarshalJSON() (bs []byte, err error) { return }

type SmapMessage C.Struct

func NewSmapMessage(s *C.Segment) SmapMessage      { return SmapMessage(s.NewStruct(0, 6)) }
func NewRootSmapMessage(s *C.Segment) SmapMessage  { return SmapMessage(s.NewRootStruct(0, 6)) }
func AutoNewSmapMessage(s *C.Segment) SmapMessage  { return SmapMessage(s.NewStructAR(0, 6)) }
func ReadRootSmapMessage(s *C.Segment) SmapMessage { return SmapMessage(s.Root(0).ToStruct()) }
func (s SmapMessage) Path() string                 { return C.Struct(s).GetObject(0).ToText() }
func (s SmapMessage) SetPath(v string)             { C.Struct(s).SetObject(0, s.Segment.NewText(v)) }
func (s SmapMessage) Uuid() []byte                 { return C.Struct(s).GetObject(1).ToData() }
func (s SmapMessage) SetUuid(v []byte)             { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s SmapMessage) Readings() SmapMessageReading_List {
	return SmapMessageReading_List(C.Struct(s).GetObject(2))
}
func (s SmapMessage) SetReadings(v SmapMessageReading_List) { C.Struct(s).SetObject(2, C.Object(v)) }
func (s SmapMessage) Contents() C.TextList                  { return C.TextList(C.Struct(s).GetObject(3)) }
func (s SmapMessage) SetContents(v C.TextList)              { C.Struct(s).SetObject(3, C.Object(v)) }
func (s SmapMessage) Properties() SmapMessagePair_List {
	return SmapMessagePair_List(C.Struct(s).GetObject(4))
}
func (s SmapMessage) SetProperties(v SmapMessagePair_List) { C.Struct(s).SetObject(4, C.Object(v)) }
func (s SmapMessage) Metadata() SmapMessagePair_List {
	return SmapMessagePair_List(C.Struct(s).GetObject(5))
}
func (s SmapMessage) SetMetadata(v SmapMessagePair_List) { C.Struct(s).SetObject(5, C.Object(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s SmapMessage) MarshalJSON() (bs []byte, err error) { return }

type SmapMessage_List C.PointerList

func NewSmapMessageList(s *C.Segment, sz int) SmapMessage_List {
	return SmapMessage_List(s.NewCompositeList(0, 6, sz))
}
func (s SmapMessage_List) Len() int             { return C.PointerList(s).Len() }
func (s SmapMessage_List) At(i int) SmapMessage { return SmapMessage(C.PointerList(s).At(i).ToStruct()) }
func (s SmapMessage_List) ToArray() []SmapMessage {
	return *(*[]SmapMessage)(unsafe.Pointer(C.PointerList(s).ToArray()))
}
func (s SmapMessage_List) Set(i int, item SmapMessage) { C.PointerList(s).Set(i, C.Object(item)) }

type SmapMessageReading C.Struct

func NewSmapMessageReading(s *C.Segment) SmapMessageReading {
	return SmapMessageReading(s.NewStruct(16, 0))
}
func NewRootSmapMessageReading(s *C.Segment) SmapMessageReading {
	return SmapMessageReading(s.NewRootStruct(16, 0))
}
func AutoNewSmapMessageReading(s *C.Segment) SmapMessageReading {
	return SmapMessageReading(s.NewStructAR(16, 0))
}
func ReadRootSmapMessageReading(s *C.Segment) SmapMessageReading {
	return SmapMessageReading(s.Root(0).ToStruct())
}
func (s SmapMessageReading) Time() uint64      { return C.Struct(s).Get64(0) }
func (s SmapMessageReading) SetTime(v uint64)  { C.Struct(s).Set64(0, v) }
func (s SmapMessageReading) Data() float64     { return math.Float64frombits(C.Struct(s).Get64(8)) }
func (s SmapMessageReading) SetData(v float64) { C.Struct(s).Set64(8, math.Float64bits(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s SmapMessageReading) MarshalJSON() (bs []byte, err error) { return }

type SmapMessageReading_List C.PointerList

func NewSmapMessageReadingList(s *C.Segment, sz int) SmapMessageReading_List {
	return SmapMessageReading_List(s.NewCompositeList(16, 0, sz))
}
func (s SmapMessageReading_List) Len() int { return C.PointerList(s).Len() }
func (s SmapMessageReading_List) At(i int) SmapMessageReading {
	return SmapMessageReading(C.PointerList(s).At(i).ToStruct())
}
func (s SmapMessageReading_List) ToArray() []SmapMessageReading {
	return *(*[]SmapMessageReading)(unsafe.Pointer(C.PointerList(s).ToArray()))
}
func (s SmapMessageReading_List) Set(i int, item SmapMessageReading) {
	C.PointerList(s).Set(i, C.Object(item))
}

type SmapMessagePair C.Struct

func NewSmapMessagePair(s *C.Segment) SmapMessagePair { return SmapMessagePair(s.NewStruct(0, 2)) }
func NewRootSmapMessagePair(s *C.Segment) SmapMessagePair {
	return SmapMessagePair(s.NewRootStruct(0, 2))
}
func AutoNewSmapMessagePair(s *C.Segment) SmapMessagePair { return SmapMessagePair(s.NewStructAR(0, 2)) }
func ReadRootSmapMessagePair(s *C.Segment) SmapMessagePair {
	return SmapMessagePair(s.Root(0).ToStruct())
}
func (s SmapMessagePair) Key() string       { return C.Struct(s).GetObject(0).ToText() }
func (s SmapMessagePair) SetKey(v string)   { C.Struct(s).SetObject(0, s.Segment.NewText(v)) }
func (s SmapMessagePair) Value() string     { return C.Struct(s).GetObject(1).ToText() }
func (s SmapMessagePair) SetValue(v string) { C.Struct(s).SetObject(1, s.Segment.NewText(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s SmapMessagePair) MarshalJSON() (bs []byte, err error) { return }

type SmapMessagePair_List C.PointerList

func NewSmapMessagePairList(s *C.Segment, sz int) SmapMessagePair_List {
	return SmapMessagePair_List(s.NewCompositeList(0, 2, sz))
}
func (s SmapMessagePair_List) Len() int { return C.PointerList(s).Len() }
func (s SmapMessagePair_List) At(i int) SmapMessagePair {
	return SmapMessagePair(C.PointerList(s).At(i).ToStruct())
}
func (s SmapMessagePair_List) ToArray() []SmapMessagePair {
	return *(*[]SmapMessagePair)(unsafe.Pointer(C.PointerList(s).ToArray()))
}
func (s SmapMessagePair_List) Set(i int, item SmapMessagePair) {
	C.PointerList(s).Set(i, C.Object(item))
}
