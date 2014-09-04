package main

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
)

/* Type of the query. Are we selecting
   data or deleting data?
*/
type QueryType_T uint

const (
	SELECT_TYPE = iota
	DELETE_TYPE
	SET_TYPE
)

/*
 * What are returning? Is it tags or data?
 */
type TargetType_T uint

const (
	TAGS_TARGET = iota
	DATA_TARGET
	SET_TARGET
)

/*
 * direction of data query
**/
type DataQueryType_T uint

const (
	IN = iota
	BEFORE
	AFTER
)

type SmapQuery struct {
	Where      *bson.M
	Type       string
	targettype string // either 'data' or 'tags'
	Contents   []string
}

func (sq *SmapQuery) Repr() {
	fmt.Println("Type: ", sq.Type)
	fmt.Println("Contents:")
	for _, val := range sq.Contents {
		fmt.Println(val)
	}
	fmt.Println("Where:")
	if sq.Where != nil {
		fmt.Println(*(sq.Where))
	}
}

type Target_T interface {
}

type TagsTarget struct {
	Distinct bool
	Contents []string
}

type SetTarget struct {
	Updates bson.M
}

type DataTarget struct {
	Type DataQueryType_T
}

func (tt TagsTarget) ToBson() bson.M {
	var item = bson.M{}
	for _, val := range tt.Contents {
		if val == "*" {
			break
		} else {
			item[val] = 1
		}
	}
	return item
}

type AST struct {
	QueryType  QueryType_T
	TargetType TargetType_T
	Target     Target_T
	Where      *Node
}

func (ast *AST) Repr() {
	fmt.Println("QueryType: ", ast.QueryType)
	fmt.Println("TargetType: ", ast.TargetType)
	fmt.Println("Target:")
	switch ast.Target.(type) {
	case (*TagsTarget):
		fmt.Println("  distinct?:", ast.Target.(*TagsTarget).Distinct)
		fmt.Println("  contents:")
		for idx, val := range ast.Target.(*TagsTarget).Contents {
			fmt.Println("    ", idx, ":", val)
		}
	case (*SetTarget):
		fmt.Println("  set target")
		fmt.Println("  ", ast.Target.(*SetTarget).Updates)
	}
	fmt.Println("Where:")
	fmt.Println(ast.Where.ToBson())
}
