package main

import (
	"gopkg.in/mgo.v2/bson"
)

type NodeType_T uint

const (
	AND_NODE = iota
	OR_NODE
	NOT_NODE
	LIKE_NODE
	HAS_NODE
	EQ_NODE
	NEQ_NODE
	LEAF_NODE
	DEF_NODE // default
)

func getNodeType(inp string) NodeType_T {
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

type Node struct {
	Type  NodeType_T
	Left  interface{}
	Right interface{}
}

func (n Node) ToBson() bson.M {
	switch n.Type {
	case EQ_NODE:
		return bson.M{n.Left.(string): n.Right.(string)}
	case NEQ_NODE:
		return bson.M{"$not": bson.M{n.Left.(string): n.Right.(string)}}
	case AND_NODE:
		return bson.M{"$and": []bson.M{n.Left.(Node).ToBson(), n.Right.(Node).ToBson()}}
	case OR_NODE:
		return bson.M{"$or": []bson.M{n.Left.(Node).ToBson(), n.Right.(Node).ToBson()}}
	case HAS_NODE:
		return bson.M{n.Left.(string): bson.M{"$exists": true}}
	}
	return bson.M{}
}

func (n Node) Length() int {
	return 0
}
