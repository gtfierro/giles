package mphandler

import (
	"fmt"
)

type MsgPackSmap struct {
	Path       string
	UUID       string `codec:"uuid"`
	Key        string `codec:"key"`
	Properties map[string]interface{}
	Metadata   map[string]interface{}
	Readings   MsgPackReadings
}

type MsgPackReadings struct {
	Readings [][2]interface{}
	UUID     string `codec:"uuid"`
}

func isstring(b byte) bool {
	return (b >= 0xa0 && b <= 0xbf)
}

func ismap(b byte) bool {
	return (b >= 0x80 && b <= 0x8f)
}

func isarray(b byte) bool {
	return (b >= 0x90 && b <= 0x9f)
}

// decodes msgpack string from byteslice
// returns deocded string and number of consumed bytes
func getstring(input *[]byte, offset int) (string, int) {
	length := int((*input)[offset] & 0x1f)
	fmt.Println("found string with length", length)
	if length == 0 {
		return "", 1
	}
	return string((*input)[offset+1 : offset+1+length]), length + 1
}

func getUint32(input *[]byte, offset int) (uint32, int) {
	return uint32((*input)[offset+0])<<24 |
		uint32((*input)[offset+1])<<16 |
		uint32((*input)[offset+2])<<8 |
		uint32((*input)[offset+3]), 4
}

func getStr16(input *[]byte, offset int) (string, int) {
	length := int(uint32((*input)[offset+1])<<8 | uint32((*input)[offset+2]))
	if length == 0 {
		return "", 3
	}
	return string((*input)[offset+3 : offset+3+length]), length + 3
}

func getarray(input *[]byte, offset int) ([]interface{}, int) {
	length := int((*input)[offset] & 0xf)
	initialoffset := offset
	offset += 1
	fmt.Println("array w/ length", length)
	if length == 0 {
		return nil, 1
	}
	var ret []interface{}
	for arridx := 0; arridx < length; arridx++ {
		var value interface{}
		consumed := 0
		if isstring((*input)[offset]) {
			value, consumed = getstring(input, offset)
		} else if isarray((*input)[offset]) {
			value, consumed = getarray(input, offset)
		} else if (*input)[offset] < 0x7f { // positive fixint
			value = uint64((*input)[offset])
			consumed = 1
			fmt.Println("offset", offset)
		} else { // is a number, probably
			switch (*input)[offset] {
			case 0xce:
				offset += 1
				value, consumed = getUint32(input, offset)
			default:
				fmt.Println("don't know what this is", (*input)[offset])
			}
		}
		offset += consumed
		ret = append(ret, value)
		fmt.Println("array value:", value)
	}
	return ret, offset - initialoffset
}

func getmap(input *[]byte, offset int) (map[string]interface{}, int) {
	length := int((*input)[offset] & 0xf)
	initialoffset := offset
	offset += 1
	fmt.Println("got map of length", length, "so it has", length*2, "elements")
	if length == 0 {
		return nil, 1
	}
	ret := map[string]interface{}{}
	for mapidx := 0; mapidx < length; mapidx++ {
		var value interface{}
		var consumed int
		// get key, assuming is string
		fmt.Println("string?", (*input)[offset])
		key, consumed := getstring(input, offset)
		fmt.Println("key:", key)
		offset += consumed
		// get value
		fmt.Println("value byte", (*input)[offset])
		if isstring((*input)[offset]) {
			value, consumed = getstring(input, offset)
		} else if isarray((*input)[offset]) {
			value, consumed = getarray(input, offset)
		} else if ismap((*input)[offset]) {
			value, consumed = getmap(input, offset)
		} else if (*input)[offset] == 0xda {
			value, consumed = getStr16(input, offset)
		} else {
			fmt.Println("actualy is dolan", (*input)[offset], offset)
		}
		fmt.Println("value:", value)
		offset += consumed
		ret[key] = value
	}
	return ret, offset - initialoffset
}

func decode(input []byte) ([]byte, map[string]interface{}) {
	idx := 0 // index into array
	// get wrapper byte: 0x8n
	// length * 2 is number of elements
	if ismap(input[idx]) {
		mymap, consumed := getmap(&input, idx)
		fmt.Println("got map", mymap)
		idx += consumed
		return input[idx:], mymap
	} else {
		fmt.Println("unrecognized beginning:", input[idx])
		return input, nil
	}
}
