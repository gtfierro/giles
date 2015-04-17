%{

package main

import (
	"bufio"
	"fmt"
	"os"
	"github.com/taylorchu/toki"
	"strconv"
)


/**
Notes here
**/
%}

%union{
	str string
	dict Dict
	data *dataquery
	limit datalimit
	list List
	time uint64
}

%token <str> SELECT DISTINCT DELETE
%token <str> WHERE
%token <str> DATA BEFORE AFTER LIMIT STREAMLIMIT NOW
%token <str> LVALUE QSTRING QREGEX
%token <str> EQ NEQ COMMA ALL
%token <str> LIKE
%token <str> AND OR HAS NOT IN
%token <str> LPAREN RPAREN
%token NUMBER
%token SEMICOLON
%token NEWLINE

%type <dict> whereList whereTerm whereClause
%type <list> selector tagList
%type <data> dataClause
%type <time> timeref
%type <limit> limit
%type <str> NUMBER
%type <str> SEMICOLON NEWLINE

%right EQ

%%

query		: SELECT selector whereClause SEMICOLON
			{
				SQlex.(*SQLex).query.contents = $2
				SQlex.(*SQLex).query.where = $3
				SQlex.(*SQLex).query.qtype = SELECT_TYPE
			}
			| SELECT selector SEMICOLON
			{
				SQlex.(*SQLex).query.contents = $2
				SQlex.(*SQLex).query.qtype = SELECT_TYPE
			}
			| SELECT dataClause whereClause SEMICOLON
			{
				SQlex.(*SQLex).query.where = $3
				SQlex.(*SQLex).query.data = $2
				SQlex.(*SQLex).query.qtype = DATA_TYPE
			}
			| DELETE tagList whereClause SEMICOLON
			{
				SQlex.(*SQLex).query.contents = $2
				SQlex.(*SQLex).query.where = $3
				SQlex.(*SQLex).query.qtype = DELETE_TYPE
			}
			;

tagList		: LVALUE
			{
				$$ = List{$1}
			}
			| LVALUE COMMA tagList
			{
				$$ = append(List{$1}, $3...)
			}
			;

selector	: tagList
			{
				$$ = $1
			}
			| ALL
			{
				$$ = List{"*"};
			}
			| DISTINCT LVALUE
			{
				SQlex.(*SQLex).query.distinct = true
				$$ = List{$2}
			}
			| DISTINCT
			{
				SQlex.(*SQLex).query.distinct = true
				$$ = List{}
			}
			;

dataClause : DATA IN LPAREN timeref COMMA timeref RPAREN limit
			{
				$$ = &dataquery{dtype: IN_TYPE, start: $4, end: $6, limit: $8}
			}
		   | DATA IN timeref COMMA timeref limit
			{
				$$ = &dataquery{dtype: IN_TYPE, start: $3, end: $5, limit: $6}
			}
		   | DATA BEFORE timeref limit
			{
				$$ = &dataquery{dtype: BEFORE_TYPE, start: $3, limit: $4}
			}
		   | DATA AFTER timeref limit
			{
				$$ = &dataquery{dtype: AFTER_TYPE, start: $3, limit: $4}
			}
		   ;

timeref		: abstime
			{
				$$ = uint64(0)
			}
			| abstime reltime
			{
				$$ = uint64(0)
			}
			;

abstime		: NUMBER
			| QSTRING
			| NOW
			;

reltime		: NUMBER LVALUE
			| NUMBER LVALUE reltime
			;

limit		: /* empty */
			{
				$$ = datalimit{limit: 0, streamlimit: 0}
			}
			| LIMIT NUMBER
			{
				num, _ := strconv.ParseUint($2, 10, 64)
				$$ = datalimit{limit: num, streamlimit: 0}
			}
			| STREAMLIMIT NUMBER
			{
				num, _ := strconv.ParseUint($2, 10, 64)
				$$ = datalimit{limit: 0, streamlimit: num}
			}
			| LIMIT NUMBER STREAMLIMIT NUMBER
			{
				limit_num, _ := strconv.ParseUint($2, 10, 64)
				slimit_num, _ := strconv.ParseUint($4, 10, 64)
				$$ = datalimit{limit: limit_num, streamlimit: slimit_num}
			}
			;


whereClause : WHERE whereList
			{
			  $$ = $2
			}
			;


whereTerm : LVALUE LIKE QREGEX
			{
				$$ = Dict{$1: Dict{"$like": $3}}
			}
		  | LVALUE EQ QSTRING
			{
				$$ = Dict{$1: Dict{"$eq": $3}}
			}
		  | LVALUE NEQ QSTRING
			{
				$$ = Dict{$1: Dict{"$neq": $3}}
			}
		  | HAS LVALUE
			{
				$$ = Dict{$2: Dict{"$exists": true}}
			}
		  ;

whereList : whereList AND whereList
			{
				$$ = Dict{"$and": []Dict{$1, $3}}
			}
		  | whereList OR whereList
			{
				$$ = Dict{"$or": []Dict{$1, $3}}
			}
		  | NOT whereList
			{
				$$ = Dict{"$not": $2} // fix this to negate all items in $2
			}
		  | LPAREN whereList RPAREN
			{
				$$ = $2
			}
		  | whereTerm
			{
				$$ = $1
			}
		  ;

%%

const eof = 0
type Dict map[string]interface{}
type List []string
type queryType uint
const (
	SELECT_TYPE queryType = iota
	DELETE_TYPE
	SET_TYPE
	DATA_TYPE
)
func (qt queryType) String() string {
	ret := ""
	switch qt {
	case SELECT_TYPE:
		ret = "select"
	case DELETE_TYPE:
		ret = "delete"
	case SET_TYPE:
		ret = "set"
	case DATA_TYPE:
		ret = "data"
	}
	return ret
}

type query struct {
	// the type of query we are doing
	qtype	   queryType
	// information about a data query if we are one
	data	   *dataquery
	// where clause for query
	where	  Dict
	// are we querying distinct values?
	distinct  bool
	// list of tags to target for deletion, selection
	contents  []string
}

type dataqueryType uint
const (
	IN_TYPE dataqueryType = iota
	BEFORE_TYPE
	AFTER_TYPE
)
func (dt dataqueryType) String() string {
	ret := ""
	switch dt {
	case IN_TYPE:
		ret = "in"
	case BEFORE_TYPE:
		ret = "before"
	case AFTER_TYPE:
		ret = "after"
	}
	return ret
}

type dataquery struct {
	dtype		dataqueryType
	start		uint64
	end			uint64
	limit		datalimit
}

type datalimit struct {
	limit		uint64
	streamlimit uint64
}

func (q *query) Print() {
	fmt.Printf("Type: %v\n", q.qtype.String())
	if q.qtype == DATA_TYPE {
		fmt.Printf("Data Query Type: %v\n", q.data.dtype.String())
		fmt.Printf("Start: %v\n", q.data.start)
		fmt.Printf("End: %v\n", q.data.end)
		fmt.Printf("Limit: %v\n", q.data.limit.limit)
		fmt.Printf("Streamlimit: %v\n", q.data.limit.streamlimit)
	}
	fmt.Printf("Contents: %v\n", q.contents)
	fmt.Printf("Distinct? %v\n", q.distinct)
	fmt.Printf("where: %v\n", q.where)
}

type SQLex struct {
	querystring string
	query	*query
	scanner *toki.Scanner
}

func NewSQLex(s string) *SQLex {
	scanner := toki.NewScanner(
		[]toki.Def{
			{Token: WHERE, Pattern: "where"},
			{Token: SELECT, Pattern: "select"},
			{Token: DELETE, Pattern: "delete"},
			{Token: DISTINCT, Pattern: "distinct"},
			{Token: LIMIT, Pattern: "limit"},
			{Token: STREAMLIMIT, Pattern: "streamlimit"},
			{Token: ALL, Pattern: "\\*"},
			{Token: NOW, Pattern: "now"},
			{Token: BEFORE, Pattern: "before"},
			{Token: AFTER, Pattern: "after"},
			{Token: COMMA, Pattern: ","},
			{Token: AND, Pattern: "and"},
			{Token: DATA, Pattern: "data"},
			{Token: OR, Pattern: "or"},
			{Token: IN, Pattern: "in"},
			{Token: HAS, Pattern: "has"},
			{Token: NOT, Pattern: "not"},
			{Token: NEQ, Pattern: "!="},
			{Token: EQ, Pattern: "="},
			{Token: LPAREN, Pattern: "\\("},
			{Token: RPAREN, Pattern: "\\)"},
			{Token: SEMICOLON, Pattern: ";"},
			{Token: NEWLINE, Pattern: "\n"},
			{Token: LIKE, Pattern: "(like)|~"},
			{Token: NUMBER, Pattern: "([+-]?([0-9]*\\.)?[0-9]+)"},
			{Token: LVALUE, Pattern: "[a-zA-Z\\~\\$\\_][a-zA-Z0-9\\/\\%_\\-]*"},
			{Token: QSTRING, Pattern: "(\"[^\"\\\\]*?(\\.[^\"\\\\]*?)*?\")|('[^'\\\\]*?(\\.[^'\\\\]*?)*?')"},
			{Token: QREGEX, Pattern: "%?[a-zA-Z0-9]+%?"},
		})
	scanner.SetInput(s)
	q := &query{contents: []string{}, distinct: false, data: &dataquery{}}
	return &SQLex{query: q, querystring: s, scanner: scanner}
}

func (sq *SQLex) Lex(lval *SQSymType) int {
	r := sq.scanner.Next()
	if r.Pos.Line == 2 {
		return eof
	}
	fmt.Printf("token: %v %v\n", r.String(), int(r.Token))
	lval.str = string(r.Value)
	return int(r.Token)
}

func (sq *SQLex) Error(s string) {
	fmt.Printf("syntax error: %s\n", s)
}

func readline(fi *bufio.Reader) (string, bool) {
	fmt.Printf("smap> ")
	s, err := fi.ReadString('\n')
	if err != nil {
		return "", false
	}
	return s, true
}

func main() {
	fi := bufio.NewReader(os.NewFile(0, "stdin"))
	for {
		if querystring, ok := readline(fi); ok {
			l := NewSQLex(querystring)
			SQParse(l)
			l.query.Print()
		} else {
			break
		}
	}
}
