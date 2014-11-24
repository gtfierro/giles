package giles

import (
	"reflect"
	"testing"
)

/**
type AST struct {
	QueryType  queryType_T
	TargetType targetType_T
	Target     target_T
	Where      *node
}
*/

func TestSimple(t *testing.T) {
	var query string
	var ast *AST

	var tt_type = reflect.TypeOf(&tagsTarget{})
	var dt_type = reflect.TypeOf(&dataTarget{})
	var st_type = reflect.TypeOf(&setTarget{})

	query = "select *"
	ast = parse(query)
	if ast.QueryType != SELECT_TYPE {
		t.Error(query, "\nshould have query type SELECT_TYPE, but has type", ast.QueryType)
	}
	if ast.TargetType != TAGS_TARGET {
		t.Error(query, "\nshould have target type TAGS_TARGET, but has type", ast.TargetType)
	}
	if reflect.TypeOf(ast.Target) != tt_type {
		t.Error(query, "\nshould have target", tt_type, "but has target", reflect.TypeOf(ast.Target))
	}

	query = "select data before now where has uuid"
	ast = parse(query)
	if ast.QueryType != SELECT_TYPE {
		t.Error(query, "\nshould have query type SELECT_TYPE, but has type", ast.QueryType)
	}
	if ast.TargetType != DATA_TARGET {
		t.Error(query, "\nshould have target type DATA_TARGET, but has type", ast.TargetType)
	}
	if reflect.TypeOf(ast.Target) != dt_type {
		t.Error(query, "\nshould have target", dt_type, "but has target", reflect.TypeOf(ast.Target))
	}

	query = "set Metadata/XYZ = 4 where has uuid"
	ast = parse(query)
	if ast.QueryType != SET_TYPE {
		t.Error(query, "\nshould have query type SET_TYPE, but has type", ast.QueryType)
	}
	if ast.TargetType != SET_TARGET {
		t.Error(query, "\nshould have target type SET_TARGET, but has type", ast.TargetType)
	}
	if reflect.TypeOf(ast.Target) != st_type {
		t.Error(query, "\nshould have target", st_type, "but has target", reflect.TypeOf(ast.Target))
	}
}
