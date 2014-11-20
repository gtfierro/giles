package giles

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"time"
)

/* Type of the query. Are we selecting
   data or deleting data?
*/
type queryType_T uint

const (
	SELECT_TYPE = iota
	DELETE_TYPE
	SET_TYPE
)

/*
 * What are returning? Is it tags or data?
 */
type targetType_T uint

const (
	TAGS_TARGET = iota
	DATA_TARGET
	SET_TARGET
)

/*
 * direction of data query
**/
type dataqueryType_T uint

const (
	IN = iota
	BEFORE
	AFTER
)

type target_T interface{}

type tagsTarget struct {
	Distinct bool
	Contents []string
}

type setTarget struct {
	Updates bson.M
}

type dataTarget struct {
	Type        dataqueryType_T
	Ref         time.Time
	Start       time.Time
	End         time.Time
	Limit       int32
	Streamlimit int
}

func (tt tagsTarget) ToBson() bson.M {
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
	QueryType  queryType_T
	TargetType targetType_T
	Target     target_T
	Where      *node
}

func (ast *AST) Repr() {
	fmt.Println("QueryType: ", ast.QueryType)
	fmt.Println("TargetType: ", ast.TargetType)
	fmt.Println("Target:")
	switch ast.Target.(type) {
	case (*tagsTarget):
		fmt.Println("  distinct?:", ast.Target.(*tagsTarget).Distinct)
		fmt.Println("  contents:")
		for idx, val := range ast.Target.(*tagsTarget).Contents {
			fmt.Println("    ", idx, ":", val)
		}
	case (*setTarget):
		fmt.Println("  set target")
		fmt.Println("  ", ast.Target.(*setTarget).Updates)
	}
	fmt.Println("Where:")
	fmt.Println(ast.Where.ToBson())
}
