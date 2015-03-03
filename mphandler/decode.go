package mphandler

type MessageType uint

const (
	DATA_WRITE = iota
	DATA_PREV
	DATA_NEXT
	DATA_RANGE
	TAG_GET
	TAG_SET
	QUERY
)

// ^^ to be continued ...

// Given a reference to a byte slice (probably your incoming buffer) and an
// offset into that slice, decode the header and return the MessageType and the
// total packet length
func ParseHeader(input *[]byte, offset int) (MessageType, int) {
	packetlength := int(getUintLE(input, offset, 2))
	return DATA_WRITE, packetlength
}

func getUintLE(input *[]byte, offset, length int) uint64 {
	var value uint64
	for i := 0; i < length; i++ {
		value |= uint64((*input)[offset+i]) << uint(i*8)
	}
	return value
}
