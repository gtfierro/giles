package cphandler

// AUTO GENERATED - DO NOT EDIT

import (
	C "github.com/glycerine/go-capnproto"
	"math"
	"unsafe"
)

type Message C.Struct

func NewMessage(s *C.Segment) Message               { return Message(s.NewStruct(0, 6)) }
func NewRootMessage(s *C.Segment) Message           { return Message(s.NewRootStruct(0, 6)) }
func AutoNewMessage(s *C.Segment) Message           { return Message(s.NewStructAR(0, 6)) }
func ReadRootMessage(s *C.Segment) Message          { return Message(s.Root(0).ToStruct()) }
func (s Message) Path() string                      { return C.Struct(s).GetObject(0).ToText() }
func (s Message) SetPath(v string)                  { C.Struct(s).SetObject(0, s.Segment.NewText(v)) }
func (s Message) Uuid() []byte                      { return C.Struct(s).GetObject(1).ToData() }
func (s Message) SetUuid(v []byte)                  { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s Message) Readings() MessageReading_List     { return MessageReading_List(C.Struct(s).GetObject(2)) }
func (s Message) SetReadings(v MessageReading_List) { C.Struct(s).SetObject(2, C.Object(v)) }
func (s Message) Contents() C.TextList              { return C.TextList(C.Struct(s).GetObject(3)) }
func (s Message) SetContents(v C.TextList)          { C.Struct(s).SetObject(3, C.Object(v)) }
func (s Message) Properties() MessagePair_List      { return MessagePair_List(C.Struct(s).GetObject(4)) }
func (s Message) SetProperties(v MessagePair_List)  { C.Struct(s).SetObject(4, C.Object(v)) }
func (s Message) Metadata() MessagePair_List        { return MessagePair_List(C.Struct(s).GetObject(5)) }
func (s Message) SetMetadata(v MessagePair_List)    { C.Struct(s).SetObject(5, C.Object(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s Message) MarshalJSON() (bs []byte, err error) { return }

type Message_List C.PointerList

func NewMessageList(s *C.Segment, sz int) Message_List {
	return Message_List(s.NewCompositeList(0, 6, sz))
}
func (s Message_List) Len() int         { return C.PointerList(s).Len() }
func (s Message_List) At(i int) Message { return Message(C.PointerList(s).At(i).ToStruct()) }
func (s Message_List) ToArray() []Message {
	return *(*[]Message)(unsafe.Pointer(C.PointerList(s).ToArray()))
}
func (s Message_List) Set(i int, item Message) { C.PointerList(s).Set(i, C.Object(item)) }

type MessageReading C.Struct

func NewMessageReading(s *C.Segment) MessageReading      { return MessageReading(s.NewStruct(16, 0)) }
func NewRootMessageReading(s *C.Segment) MessageReading  { return MessageReading(s.NewRootStruct(16, 0)) }
func AutoNewMessageReading(s *C.Segment) MessageReading  { return MessageReading(s.NewStructAR(16, 0)) }
func ReadRootMessageReading(s *C.Segment) MessageReading { return MessageReading(s.Root(0).ToStruct()) }
func (s MessageReading) Time() uint64                    { return C.Struct(s).Get64(0) }
func (s MessageReading) SetTime(v uint64)                { C.Struct(s).Set64(0, v) }
func (s MessageReading) Data() float64                   { return math.Float64frombits(C.Struct(s).Get64(8)) }
func (s MessageReading) SetData(v float64)               { C.Struct(s).Set64(8, math.Float64bits(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s MessageReading) MarshalJSON() (bs []byte, err error) { return }

type MessageReading_List C.PointerList

func NewMessageReadingList(s *C.Segment, sz int) MessageReading_List {
	return MessageReading_List(s.NewCompositeList(16, 0, sz))
}
func (s MessageReading_List) Len() int { return C.PointerList(s).Len() }
func (s MessageReading_List) At(i int) MessageReading {
	return MessageReading(C.PointerList(s).At(i).ToStruct())
}
func (s MessageReading_List) ToArray() []MessageReading {
	return *(*[]MessageReading)(unsafe.Pointer(C.PointerList(s).ToArray()))
}
func (s MessageReading_List) Set(i int, item MessageReading) { C.PointerList(s).Set(i, C.Object(item)) }

type MessagePair C.Struct

func NewMessagePair(s *C.Segment) MessagePair      { return MessagePair(s.NewStruct(0, 2)) }
func NewRootMessagePair(s *C.Segment) MessagePair  { return MessagePair(s.NewRootStruct(0, 2)) }
func AutoNewMessagePair(s *C.Segment) MessagePair  { return MessagePair(s.NewStructAR(0, 2)) }
func ReadRootMessagePair(s *C.Segment) MessagePair { return MessagePair(s.Root(0).ToStruct()) }
func (s MessagePair) Key() string                  { return C.Struct(s).GetObject(0).ToText() }
func (s MessagePair) SetKey(v string)              { C.Struct(s).SetObject(0, s.Segment.NewText(v)) }
func (s MessagePair) Value() string                { return C.Struct(s).GetObject(1).ToText() }
func (s MessagePair) SetValue(v string)            { C.Struct(s).SetObject(1, s.Segment.NewText(v)) }

// capn.JSON_enabled == false so we stub MarshallJSON().
func (s MessagePair) MarshalJSON() (bs []byte, err error) { return }

type MessagePair_List C.PointerList

func NewMessagePairList(s *C.Segment, sz int) MessagePair_List {
	return MessagePair_List(s.NewCompositeList(0, 2, sz))
}
func (s MessagePair_List) Len() int             { return C.PointerList(s).Len() }
func (s MessagePair_List) At(i int) MessagePair { return MessagePair(C.PointerList(s).At(i).ToStruct()) }
func (s MessagePair_List) ToArray() []MessagePair {
	return *(*[]MessagePair)(unsafe.Pointer(C.PointerList(s).ToArray()))
}
func (s MessagePair_List) Set(i int, item MessagePair) { C.PointerList(s).Set(i, C.Object(item)) }
