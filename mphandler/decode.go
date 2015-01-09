package mphandler

import (
	"encoding/binary"
	"math"
)

func isstring(b byte) bool {
	return (b >= 0xa0 && b <= 0xbf)
}

func ismap(b byte) bool {
	return (b >= 0x80 && b <= 0x8f)
}

func isarray(b byte) bool {
	return (b >= 0x90 && b <= 0x9f)
}

func getUint(input *[]byte, offset, length int) uint64 {
	var value uint64
	for i := 0; i < length; i++ {
		value |= uint64((*input)[offset+i]) << uint((length-i-1)*8)
	}
	return value
}

func parseUint(input *[]byte, offset int) (uint64, int) {
	var (
		value    uint64
		consumed int
	)
	c := (*input)[offset]
	switch {
	case c == 0xcc:
		value = uint64((*input)[offset+1])
		consumed = 2
	case c == 0xcd:
		value = getUint(input, offset+1, 2)
		consumed = 3
	case c == 0xce:
		value = getUint(input, offset+1, 4)
		consumed = 5
	case c == 0xcf:
		value = getUint(input, offset+1, 8)
		consumed = 9
	}
	return value, consumed
}

//TODO: handle negative fixint
func parseInt(input *[]byte, offset int) (int64, int) {
	var (
		value    int64
		consumed int
		tmp      uint64 // used for zero-copy conversion
	)
	c := (*input)[offset]
	switch {
	case c <= 0x7f:
		tmp = uint64((*input)[offset])
		consumed = 1
	case c == 0xd0:
		tmp = uint64((*input)[offset+1])
		consumed = 2
	case c == 0xd1:
		tmp = getUint(input, offset+1, 2)
		consumed = 3
	case c == 0xd2:
		tmp = getUint(input, offset+1, 4)
		consumed = 5
	case c == 0xd3:
		tmp = getUint(input, offset+1, 8)
		consumed = 9
	}

	value = int64(tmp >> 1)
	if tmp&1 != 0 {
		value = ^value
	}
	return value, consumed
}

//TODO: parsing bigendian floats is not zero copy! :(
func parseFloat(input *[]byte, offset int) (float64, int) {
	var (
		value    float64
		consumed int
	)
	c := (*input)[offset]
	switch {
	case c == 0xca:
		bits := binary.BigEndian.Uint64((*input)[offset+1 : offset+5])
		value = math.Float64frombits(bits)
		consumed = 5
	case c == 0xcb:
		bits := binary.BigEndian.Uint64((*input)[offset+1 : offset+9])
		value = math.Float64frombits(bits)
		consumed = 9
	}
	return value, consumed
}

func parseString(input *[]byte, offset int) (string, int) {
	var (
		value    string
		consumed int
		length   int
	)
	c := (*input)[offset]
	switch {
	case c >= 0xa0 && c <= 0xbf:
		length = int(c & 0x1f)
		value = string((*input)[offset+1 : offset+1+length])
		consumed = length + 1
	case c == 0xd9:
		length = int((*input)[offset+1])
		value = string((*input)[offset+2 : offset+2+length])
		consumed = length + 2
	case c == 0xda:
		length = int(getUint(input, offset+1, 2))
		value = string((*input)[offset+3 : offset+3+length])
		consumed = length + 3
	case c == 0xdb:
		length = int(getUint(input, offset+1, 4))
		value = string((*input)[offset+5 : offset+5+length])
		consumed = length + 5
	}
	return value, consumed
}

func parseMap(input *[]byte, offset int) (map[string]interface{}, int) {
	var (
		value    map[string]interface{}
		consumed int
		length   int
	)
	initialoffset := offset
	c := (*input)[offset]
	switch {
	case c >= 0x80 && c <= 0x8f:
		length = int((*input)[offset] & 0xf)
		offset += 1
	case c == 0xde:
		length = int(getUint(input, offset+1, 2))
		offset += 3
	case c == 0xdf:
		length = int(getUint(input, offset+1, 4))
		offset += 5
	}
	value = make(map[string]interface{}, length)
	// get both a key and value for [length] elements
	for mapidx := 0; mapidx < length; mapidx++ {
		var key string
		var ok bool
		newoffset, _key := decode(input, offset)
		if key, ok = _key.(string); !ok {
			log.Debug("have key we don't understand: %v", _key)
			return value, consumed
		}
		offset = newoffset
		newoffset, _value := decode(input, offset)
		value[key] = _value
		offset = newoffset
	}
	return value, offset - initialoffset
}

func parseArray(input *[]byte, offset int) ([]interface{}, int) {
	var (
		value  []interface{}
		length int
	)
	initialoffset := offset
	c := (*input)[offset]
	switch {
	case c >= 0x90 && c <= 0x9f:
		length = int((*input)[offset] & 0xf)
		offset += 1
	case c == 0xdc:
		length = int(getUint(input, offset+1, 2))
		offset += 3
	case c == 0xdd:
		length = int(getUint(input, offset+1, 4))
		offset += 5
	}
	value = make([]interface{}, length, length)
	for arridx := 0; arridx < length; arridx++ {
		newoffset, _val := decode(input, offset)
		offset = newoffset
		value[arridx] = _val
	}
	return value, offset - initialoffset
}

func decode(input *[]byte, offset int) (int, interface{}) {
	c := (*input)[offset]
	var (
		value    interface{} // the decoded value
		consumed int         // how many bytes that value used
	)
	switch {
	// int64
	case 0x00 <= c && c <= 0x7f, //positive fixint
		0xe0 <= c && c <= 0xff, //negative fixint
		0xd0 == c,              //int8
		0xd1 == c,              //int16
		0xd2 == c,              //int32
		0xd3 == c:              //int64
		value, consumed = parseInt(input, offset)

	// uint64
	case 0xcc == c, //uint8
		0xcd == c, //uint16
		0xce == c, //uint32
		0xcf == c: //uint64
		value, consumed = parseUint(input, offset)

	// float64
	case 0xca == c, //float32
		0xcb == c: //float64
		value, consumed = parseFloat(input, offset)

	// string
	case 0xa0 <= c && c <= 0xbf, //fixstr
		0xd9 == c, //str8
		0xda == c, //str16
		0xdb == c: //str32
		value, consumed = parseString(input, offset)

	// map[string]interface{}
	case 0x80 <= c && c <= 0x8f, //fixmap
		0xde == c, //map 16
		0xdf == c: //map 32
		value, consumed = parseMap(input, offset)

	// array []interface{}
	case 0x90 <= c && c <= 0x9f, //fixarray
		0xdc == c, //array 16
		0xdd == c: //array 32
		value, consumed = parseArray(input, offset)

	case 0xc0 == c: //nil

	case 0xc2 == c: //false
	case 0xc3 == c: //true

	case 0xc4 == c: //bin8
	case 0xc5 == c: //bin16
	case 0xc6 == c: //bin32

	case 0xc7 == c: //ext8
	case 0xc8 == c: //ext16
	case 0xc9 == c: //ext32

	case 0xd4 == c: //fixext 1
	case 0xd5 == c: //fixext 2
	case 0xd6 == c: //fixext 4
	case 0xd7 == c: //fixext 8
	case 0xd8 == c: //fixext 16

	default:
		log.Debug("actualy is dolan")
	}
	offset += consumed
	return offset, value
}
