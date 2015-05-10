%{

package archiver

import (
	"bufio"
	"fmt"
	"os"
	"github.com/taylorchu/toki"
	"strconv"
	"gopkg.in/mgo.v2/bson"
    _time "time"
    "strings"
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
    timeconv UnitOfTime
	list List
	time _time.Time
    timediff _time.Duration
}

%token <str> SELECT DISTINCT DELETE SET
%token <str> WHERE
%token <str> DATA BEFORE AFTER LIMIT STREAMLIMIT NOW
%token <str> LVALUE QSTRING
%token <str> EQ NEQ COMMA ALL
%token <str> LIKE AS
%token <str> AND OR HAS NOT IN
%token <str> LPAREN RPAREN
%token NUMBER
%token SEMICOLON
%token NEWLINE
%token TIMEUNIT

%type <dict> whereList whereTerm whereClause setList
%type <list> selector tagList
%type <data> dataClause
%type <time> timeref abstime
%type <timediff> reltime
%type <limit> limit
%type <timeconv> timeconv
%type <str> NUMBER qstring lvalue TIMEUNIT
%type <str> SEMICOLON NEWLINE

%right EQ

%%

query		: SELECT selector whereClause SEMICOLON
			{
				SQlex.(*SQLex).query.Contents = $2
				SQlex.(*SQLex).query.where = $3
				SQlex.(*SQLex).query.qtype = SELECT_TYPE
			}
			| SELECT selector SEMICOLON
			{
				SQlex.(*SQLex).query.Contents = $2
				SQlex.(*SQLex).query.qtype = SELECT_TYPE
			}
			| SELECT dataClause whereClause SEMICOLON
			{
				SQlex.(*SQLex).query.where = $3
				SQlex.(*SQLex).query.data = $2
				SQlex.(*SQLex).query.qtype = DATA_TYPE
			}
            | SET setList whereClause SEMICOLON
            {
				SQlex.(*SQLex).query.where = $3
				SQlex.(*SQLex).query.set = $2
                SQlex.(*SQLex).query.qtype = SET_TYPE
            }
            | SET setList SEMICOLON
            {
				SQlex.(*SQLex).query.set = $2
                SQlex.(*SQLex).query.qtype = SET_TYPE
            }
			| DELETE tagList whereClause SEMICOLON
			{
				SQlex.(*SQLex).query.Contents = $2
				SQlex.(*SQLex).query.where = $3
				SQlex.(*SQLex).query.qtype = DELETE_TYPE
			}
			| DELETE whereClause SEMICOLON
			{
				SQlex.(*SQLex).query.Contents = []string{}
				SQlex.(*SQLex).query.where = $2
				SQlex.(*SQLex).query.qtype = DELETE_TYPE
			}
			;

tagList		: lvalue
			{
				$$ = List{$1}
			}
			| lvalue COMMA tagList
			{
				$$ = append(List{$1}, $3...)
			}
			;

setList     : lvalue EQ qstring
            {
                $$ = Dict{$1: $3}
            }
            | lvalue EQ NUMBER
            {
                $$ = Dict{$1: $3}
            }
            | lvalue EQ qstring COMMA setList
            {
                $5[$1] = $3
                $$ = $5
            }
            | lvalue EQ NUMBER COMMA setList
            {
                $5[$1] = $3
                $$ = $5
            }
            ;

selector	: tagList
			{
				$$ = $1
			}
			| ALL
			{
				$$ = List{};
			}
			| DISTINCT lvalue
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

dataClause : DATA IN LPAREN timeref COMMA timeref RPAREN limit timeconv
			{
				$$ = &dataquery{dtype: IN_TYPE, start: $4, end: $6, limit: $8, timeconv: $9}
			}
		   | DATA IN timeref COMMA timeref limit timeconv
			{
				$$ = &dataquery{dtype: IN_TYPE, start: $3, end: $5, limit: $6, timeconv: $7}
			}
		   | DATA BEFORE timeref limit timeconv
			{
				$$ = &dataquery{dtype: BEFORE_TYPE, start: $3, limit: $4, timeconv: $5}
			}
		   | DATA AFTER timeref limit timeconv
			{
				$$ = &dataquery{dtype: AFTER_TYPE, start: $3, limit: $4, timeconv: $5}
			}
		   ;

timeref		: abstime
			{
				$$ = $1
			}
			| abstime reltime
			{
                $$ = $1.Add($2)
			}
			;

abstime		: NUMBER TIMEUNIT
            {
                foundtime, err := parseAbsTime($1, $2)
                if err != nil {
				    SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse time \"%v %v\" (%v)", $1, $2, err.Error()))
                }
                $$ = foundtime
            }
            | NUMBER
            {
                num, err := strconv.ParseInt($1, 10, 64)
                if err != nil {
				    SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", $1, err.Error()))
                }
                $$ = _time.Unix(num, 0)
            }
			| qstring
            {
                found := false
                for _, format := range supported_formats {
                    t, err := _time.Parse(format, $1)
                    if err != nil {
                        continue
                    }
                    $$ = t
                    found = true
                    break
                }
                if !found {
				    SQlex.(*SQLex).Error(fmt.Sprintf("No time format matching \"%v\" found", $1))
                }
            }
			| NOW
            {
                $$ = _time.Now()
            }
			;

reltime		: NUMBER lvalue
            {
                var err error
                $$, err = parseReltime($1, $2)
                if err != nil {
				    SQlex.(*SQLex).Error(fmt.Sprintf("Error parsing relative time \"%v %v\" (%v)", $1, $2, err.Error()))
                }
            }
			| NUMBER lvalue reltime
            {
                newDuration, err := parseReltime($1, $2)
                if err != nil {
				    SQlex.(*SQLex).Error(fmt.Sprintf("Error parsing relative time \"%v %v\" (%v)", $1, $2, err.Error()))
                }
                $$ = addDurations(newDuration, $3)
            }
			;

limit		: /* empty */
			{
				$$ = datalimit{limit: -1, streamlimit: -1}
			}
			| LIMIT NUMBER
			{
				num, err := strconv.ParseInt($2, 10, 64)
                if err != nil {
				    SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", $2, err.Error()))
                }
				$$ = datalimit{limit: num, streamlimit: -1}
			}
			| STREAMLIMIT NUMBER
			{
				num, err := strconv.ParseInt($2, 10, 64)
                if err != nil {
				    SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", $2, err.Error()))
                }
				$$ = datalimit{limit: -1, streamlimit: num}
			}
			| LIMIT NUMBER STREAMLIMIT NUMBER
			{
				limit_num, err := strconv.ParseInt($2, 10, 64)
                if err != nil {
				    SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", $2, err.Error()))
                }
				slimit_num, err := strconv.ParseInt($4, 10, 64)
                if err != nil {
				    SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", $2, err.Error()))
                }
				$$ = datalimit{limit: limit_num, streamlimit: slimit_num}
			}
			;

timeconv    : /* empty */
            {
                $$ = UOT_MS
            }
            | AS TIMEUNIT
            {
                uot, err := parseUOT($2)
                if err != nil {
                    SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse unit of time %v (%v)", $2, err))
                }
                $$ = uot
            }
            ;



whereClause : WHERE whereList
			{
			  $$ = $2
			}
			;


whereTerm : lvalue LIKE qstring
			{
				$$ = Dict{$1: Dict{"$regex": $3}}
			}
		  | lvalue EQ qstring
			{
				$$ = Dict{$1: $3}
			}
          | lvalue EQ NUMBER
            {
				$$ = Dict{$1: $3}
            }
		  | lvalue NEQ qstring
			{
				$$ = Dict{$1: Dict{"$neq": $3}}
			}
		  | HAS lvalue
			{
				$$ = Dict{$2: Dict{"$exists": true}}
			}
		  ;

qstring   : QSTRING
          {
            $$ = $1[1:len($1)-1]
          }
          ;

lvalue    : LVALUE
          {
            
		    SQlex.(*SQLex)._keys[$1] = struct{}{}
            $$ = cleantagstring($1)
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
var supported_formats = []string{"1/2/2006",
                                 "1-2-2006",
                                 "1/2/2006 03:04:05 PM MST",
                                 "1-2-2006 03:04:05 PM MST",
                                 "1/2/2006 15:04:05 MST",
                                 "1-2-2006 15:04:05 MST",
                                 "2006-1-2 15:04:05 MST"}
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
    // key-value pairs to add
    set         Dict
	// where clause for query
	where	  Dict
	// are we querying distinct values?
	distinct  bool
	// list of tags to target for deletion, selection
	Contents  []string
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
	fmt.Printf("Contents: %v\n", q.Contents)
	fmt.Printf("Distinct? %v\n", q.distinct)
	fmt.Printf("where: %v\n", q.where)
}

func (q *query) ContentsBson() bson.M {
    ret := bson.M{}
    for _, tag := range q.Contents {
        ret[tag] = 1
    }
    return ret
}

func (q *query) WhereBson() bson.M {
    return bson.M(q.where)
}

func (q *query) SetBson() bson.M {
    return bson.M(q.set)
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
	start		_time.Time
	end			_time.Time
	limit		datalimit
    timeconv  UnitOfTime
}

type datalimit struct {
	limit		int64
	streamlimit int64
}


type SQLex struct {
	querystring string
	query	*query
	scanner *toki.Scanner
    lasttoken string
    tokens  []string
    error   error
    // all keys that we encounter. Used for republish concerns
    _keys    map[string]struct{}
    keys    []string
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
			{Token: SET, Pattern: "set"},
			{Token: BEFORE, Pattern: "before"},
			{Token: AFTER, Pattern: "after"},
			{Token: COMMA, Pattern: ","},
			{Token: AND, Pattern: "and"},
			{Token: AS, Pattern: "as"},
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
            {Token: TIMEUNIT, Pattern: "(ns|us|ms|s)"},
			{Token: LIKE, Pattern: "(like)|~"},
			{Token: NUMBER, Pattern: "([+-]?([0-9]*\\.)?[0-9]+)"},
			{Token: LVALUE, Pattern: "[a-zA-Z\\~\\$\\_][a-zA-Z0-9\\/\\%_\\-]*"},
			{Token: QSTRING, Pattern: "(\"[^\"\\\\]*?(\\.[^\"\\\\]*?)*?\")|('[^'\\\\]*?(\\.[^'\\\\]*?)*?')"},
		})
	scanner.SetInput(s)
	q := &query{Contents: []string{}, distinct: false, data: &dataquery{}}
	return &SQLex{query: q, querystring: s, scanner: scanner, error: nil, lasttoken: "", _keys: map[string]struct{}{}, tokens: []string{}}
}

func (sq *SQLex) Lex(lval *SQSymType) int {
	r := sq.scanner.Next()
    sq.lasttoken = r.String()
	if r.Pos.Line == 2 || len(r.Value) == 0 {
		return eof
	}
	lval.str = string(r.Value)
    sq.tokens = append(sq.tokens, lval.str)
	return int(r.Token)
}

func (sq *SQLex) Error(s string) {
    sq.error = fmt.Errorf(s)
}

func readline(fi *bufio.Reader) (string, bool) {
	fmt.Printf("smap> ")
	s, err := fi.ReadString('\n')
	if err != nil {
		return "", false
	}
	return s, true
}

func Parse(querystring string) *SQLex {
    if !strings.HasSuffix(querystring, ";") {
        querystring = querystring + ";"
    }
	l := NewSQLex(querystring)
    fmt.Printf("Query: %v\n", querystring)
	SQParse(l)
	l.query.Print()
    l.keys = make([]string, len(l._keys))
    i := 0
    for key, _ := range l._keys {
        l.keys[i] = cleantagstring(key)
        i += 1
    }
    return l
}

func main() {
	fi := bufio.NewReader(os.NewFile(0, "stdin"))
	for {
		if querystring, ok := readline(fi); ok {
			Parse(querystring)
		} else {
			break
		}
	}
}
