package archiver

import (
	"gopkg.in/mgo.v2/bson"
	"strings"
)

type nodeType_T uint

const (
	DEF_NODE = iota
	AND_NODE
	OR_NODE
	NOT_NODE
	LIKE_NODE
	HAS_NODE
	EQ_NODE
	NEQ_NODE
	LEAF_NODE
)

func getnodeType(inp string) nodeType_T {
	switch inp {
	case "and":
		return AND_NODE
	case "or":
		return OR_NODE
	case "not":
		return NOT_NODE
	case "like":
		return LIKE_NODE
	case "~":
		return LIKE_NODE
	case "has":
		return HAS_NODE
	case "=":
		return EQ_NODE
	case "!=":
		return NEQ_NODE
	default:
		return DEF_NODE
	}
}

type node struct {
	Type  nodeType_T
	Left  interface{}
	Right interface{}
}

func (n node) ToBson() bson.M {
	switch n.Type {
	case EQ_NODE:
		return bson.M{n.Left.(string): n.Right.(string)}
	case LIKE_NODE:
		return bson.M{n.Left.(string): bson.M{"$regex": strings.Replace(n.Right.(string), "%", ".*", -1)}}
	case NEQ_NODE:
		return bson.M{"$not": bson.M{n.Left.(string): n.Right.(string)}}
	case AND_NODE:
		return bson.M{"$and": []bson.M{n.Left.(node).ToBson(), n.Right.(node).ToBson()}}
	case OR_NODE:
		return bson.M{"$or": []bson.M{n.Left.(node).ToBson(), n.Right.(node).ToBson()}}
	case HAS_NODE:
		return bson.M{n.Left.(string): bson.M{"$exists": true}}
	}
	return bson.M{}
}

func (n node) Length() int {
	return 0
}
