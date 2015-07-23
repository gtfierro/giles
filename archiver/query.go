//line query.y:2
package archiver

import __yyfmt__ "fmt"

//line query.y:3
import (
	"bufio"
	"fmt"
	"github.com/taylorchu/toki"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	_time "time"
)

/**
Notes here
**/

//line query.y:20
type SQSymType struct {
	yys      int
	str      string
	dict     Dict
	oplist   []*OpNode
	op       *OpNode
	data     *dataquery
	limit    datalimit
	timeconv UnitOfTime
	list     List
	time     _time.Time
	timediff _time.Duration
}

const SELECT = 57346
const DISTINCT = 57347
const DELETE = 57348
const SET = 57349
const APPLY = 57350
const WHERE = 57351
const DATA = 57352
const BEFORE = 57353
const AFTER = 57354
const LIMIT = 57355
const STREAMLIMIT = 57356
const NOW = 57357
const LVALUE = 57358
const QSTRING = 57359
const OPERATOR = 57360
const EQ = 57361
const NEQ = 57362
const COMMA = 57363
const ALL = 57364
const LEFTPIPE = 57365
const LIKE = 57366
const AS = 57367
const AND = 57368
const OR = 57369
const HAS = 57370
const NOT = 57371
const IN = 57372
const TO = 57373
const LPAREN = 57374
const RPAREN = 57375
const LBRACK = 57376
const RBRACK = 57377
const NUMBER = 57378
const SEMICOLON = 57379
const NEWLINE = 57380
const TIMEUNIT = 57381

var SQToknames = []string{
	"SELECT",
	"DISTINCT",
	"DELETE",
	"SET",
	"APPLY",
	"WHERE",
	"DATA",
	"BEFORE",
	"AFTER",
	"LIMIT",
	"STREAMLIMIT",
	"NOW",
	"LVALUE",
	"QSTRING",
	"OPERATOR",
	"EQ",
	"NEQ",
	"COMMA",
	"ALL",
	"LEFTPIPE",
	"LIKE",
	"AS",
	"AND",
	"OR",
	"HAS",
	"NOT",
	"IN",
	"TO",
	"LPAREN",
	"RPAREN",
	"LBRACK",
	"RBRACK",
	"NUMBER",
	"SEMICOLON",
	"NEWLINE",
	"TIMEUNIT",
}
var SQStatenames = []string{}

const SQEofCode = 1
const SQErrCode = 2
const SQMaxDepth = 200

//line query.y:436

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
type OpNode struct {
	Operator  string
	Arguments Dict
}
type queryType uint

const (
	SELECT_TYPE queryType = iota
	DELETE_TYPE
	SET_TYPE
	DATA_TYPE
	APPLY_TYPE
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
	qtype queryType
	// information about a data query if we are one
	data *dataquery
	// key-value pairs to add
	set Dict
	// where clause for query
	where Dict
	// are we querying distinct values?
	distinct bool
	// list of tags to target for deletion, selection
	Contents []string
	// formed operator tree
	operators []*OpNode
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
	dtype    dataqueryType
	start    _time.Time
	end      _time.Time
	limit    datalimit
	timeconv UnitOfTime
}

type datalimit struct {
	limit       int64
	streamlimit int64
}

type SQLex struct {
	querystring string
	query       *query
	scanner     *toki.Scanner
	lasttoken   string
	tokens      []string
	error       error
	// all keys that we encounter. Used for republish concerns
	_keys map[string]struct{}
	keys  []string
}

func NewSQLex(s string) *SQLex {
	scanner := toki.NewScanner(
		[]toki.Def{
			{Token: WHERE, Pattern: "where"},
			{Token: SELECT, Pattern: "select"},
			{Token: APPLY, Pattern: "apply"},
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
			{Token: TO, Pattern: "to"},
			{Token: DATA, Pattern: "data"},
			{Token: OR, Pattern: "or"},
			{Token: IN, Pattern: "in"},
			{Token: HAS, Pattern: "has"},
			{Token: NOT, Pattern: "not"},
			{Token: NEQ, Pattern: "!="},
			{Token: EQ, Pattern: "="},
			{Token: LEFTPIPE, Pattern: "<"},
			{Token: LPAREN, Pattern: "\\("},
			{Token: RPAREN, Pattern: "\\)"},
			{Token: LBRACK, Pattern: "\\["},
			{Token: RBRACK, Pattern: "\\]"},
			{Token: SEMICOLON, Pattern: ";"},
			{Token: NEWLINE, Pattern: "\n"},
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

// Parse has been moved to query_processor.go

//line yacctab:1
var SQExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const SQNprod = 65
const SQPrivate = 57344

var SQTokenNames []string
var SQStates []string

const SQLast = 159

var SQAct = []int{

	109, 85, 39, 82, 78, 12, 15, 12, 52, 14,
	73, 17, 19, 25, 54, 35, 49, 7, 22, 24,
	18, 53, 119, 54, 62, 53, 30, 54, 33, 58,
	54, 42, 12, 60, 131, 54, 18, 47, 48, 46,
	34, 59, 51, 70, 55, 56, 51, 101, 31, 120,
	13, 74, 65, 66, 97, 83, 112, 76, 88, 111,
	45, 75, 40, 36, 23, 80, 37, 79, 42, 129,
	104, 41, 27, 28, 99, 43, 95, 96, 98, 92,
	93, 63, 64, 100, 77, 8, 108, 103, 94, 113,
	16, 26, 15, 15, 15, 72, 71, 110, 107, 114,
	115, 116, 128, 117, 61, 63, 64, 44, 127, 123,
	106, 74, 124, 118, 121, 57, 68, 69, 79, 102,
	91, 67, 90, 122, 130, 10, 89, 81, 29, 105,
	11, 134, 132, 133, 32, 135, 13, 18, 54, 125,
	13, 84, 9, 21, 13, 86, 87, 126, 2, 11,
	4, 3, 5, 1, 18, 50, 20, 6, 38,
}
var SQPact = []int{

	144, -1000, 120, 124, 128, 127, 27, 145, -1000, -1000,
	124, 61, 107, -1000, 11, 115, 145, 3, 34, 44,
	84, 28, 2, -1000, 0, -1000, 6, 10, 10, 124,
	-8, -1000, -3, -13, -1000, 79, 34, 34, -1000, 97,
	124, 66, 121, 139, 127, 51, -1000, -1000, 10, 106,
	19, 125, -1000, -1000, -1000, 132, 132, -1000, -1000, 105,
	101, 99, -1000, 34, 34, 79, 55, 121, 18, 121,
	-1000, 124, 53, 12, 98, 145, -1000, -1000, 37, 110,
	89, 10, -1000, 124, -1000, 72, 23, 20, 72, 124,
	124, 124, 79, 79, -1000, -1000, -1000, -1000, -1000, -1000,
	124, -1000, 121, -15, -1000, 13, 10, 132, 19, -1000,
	123, 133, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	87, 81, 36, 72, -1000, -1000, -2, 102, 102, 132,
	-1000, -1000, -1000, -1000, 72, -1000,
}
var SQPgo = []int{

	0, 15, 158, 11, 9, 4, 157, 85, 10, 71,
	12, 156, 17, 16, 155, 3, 1, 0, 8, 2,
	153,
}
var SQR1 = []int{

	0, 20, 20, 20, 20, 20, 20, 20, 20, 7,
	7, 9, 8, 8, 4, 4, 4, 4, 4, 4,
	6, 6, 6, 6, 12, 12, 12, 12, 13, 13,
	14, 14, 14, 14, 15, 15, 16, 16, 16, 16,
	17, 17, 3, 2, 2, 2, 2, 2, 2, 2,
	18, 19, 1, 1, 1, 1, 1, 10, 10, 11,
	11, 5, 5, 5, 5,
}
var SQR2 = []int{

	0, 4, 3, 4, 4, 3, 4, 3, 6, 1,
	3, 3, 1, 3, 3, 3, 3, 5, 5, 5,
	1, 1, 2, 1, 9, 7, 5, 5, 1, 2,
	2, 1, 1, 1, 2, 3, 0, 2, 2, 4,
	0, 2, 2, 3, 3, 3, 3, 2, 3, 4,
	1, 1, 3, 3, 2, 3, 1, 1, 3, 3,
	4, 3, 3, 5, 5,
}
var SQChk = []int{

	-1000, -20, 4, 7, 6, 8, -6, -12, -7, 22,
	5, 10, -19, 16, -4, -19, -7, -3, 9, -10,
	-11, 16, -3, 37, -3, -19, 30, 11, 12, 21,
	-3, 37, 19, -3, 37, -1, 29, 32, -2, -19,
	28, -9, 34, 31, 23, 32, 37, 37, 32, -13,
	-14, 36, -18, 15, 17, -13, -13, -7, 37, -18,
	36, -9, 37, 26, 27, -1, -1, 24, 19, 20,
	-19, 30, 29, -8, -18, -12, -10, 33, -5, 16,
	-13, 21, -15, 36, 16, -16, 13, 14, -16, 21,
	21, 21, -1, -1, 33, -18, -18, 36, -18, -19,
	30, 35, 21, -3, 33, 19, 21, -13, -19, -17,
	25, 36, 36, -17, -4, -4, -4, -19, -8, 37,
	36, -18, -13, -16, -15, 16, 14, 21, 21, 33,
	-17, 36, -5, -5, -16, -17,
}
var SQDef = []int{

	0, -2, 0, 0, 0, 0, 0, 0, 20, 21,
	23, 0, 9, 51, 0, 0, 0, 0, 0, 0,
	57, 0, 0, 2, 0, 22, 0, 0, 0, 0,
	0, 5, 0, 0, 7, 42, 0, 0, 56, 0,
	0, 0, 0, 0, 0, 0, 1, 3, 0, 0,
	28, 31, 32, 33, 50, 36, 36, 10, 4, 14,
	15, 16, 6, 0, 0, 54, 0, 0, 0, 0,
	47, 0, 0, 0, 12, 0, 58, 59, 0, 0,
	0, 0, 29, 0, 30, 40, 0, 0, 40, 0,
	0, 0, 52, 53, 55, 43, 44, 45, 46, 48,
	0, 11, 0, 0, 60, 0, 0, 36, 34, 26,
	0, 37, 38, 27, 17, 18, 19, 49, 13, 8,
	61, 62, 0, 40, 35, 41, 0, 0, 0, 36,
	25, 39, 63, 64, 40, 24,
}
var SQTok1 = []int{

	1,
}
var SQTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39,
}
var SQTok3 = []int{
	0,
}

//line yaccpar:1

/*	parser for yacc output	*/

var SQDebug = 0

type SQLexer interface {
	Lex(lval *SQSymType) int
	Error(s string)
}

const SQFlag = -1000

func SQTokname(c int) string {
	// 4 is TOKSTART above
	if c >= 4 && c-4 < len(SQToknames) {
		if SQToknames[c-4] != "" {
			return SQToknames[c-4]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func SQStatname(s int) string {
	if s >= 0 && s < len(SQStatenames) {
		if SQStatenames[s] != "" {
			return SQStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func SQlex1(lex SQLexer, lval *SQSymType) int {
	c := 0
	char := lex.Lex(lval)
	if char <= 0 {
		c = SQTok1[0]
		goto out
	}
	if char < len(SQTok1) {
		c = SQTok1[char]
		goto out
	}
	if char >= SQPrivate {
		if char < SQPrivate+len(SQTok2) {
			c = SQTok2[char-SQPrivate]
			goto out
		}
	}
	for i := 0; i < len(SQTok3); i += 2 {
		c = SQTok3[i+0]
		if c == char {
			c = SQTok3[i+1]
			goto out
		}
	}

out:
	if c == 0 {
		c = SQTok2[1] /* unknown char */
	}
	if SQDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", SQTokname(c), uint(char))
	}
	return c
}

func SQParse(SQlex SQLexer) int {
	var SQn int
	var SQlval SQSymType
	var SQVAL SQSymType
	SQS := make([]SQSymType, SQMaxDepth)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	SQstate := 0
	SQchar := -1
	SQp := -1
	goto SQstack

ret0:
	return 0

ret1:
	return 1

SQstack:
	/* put a state and value onto the stack */
	if SQDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", SQTokname(SQchar), SQStatname(SQstate))
	}

	SQp++
	if SQp >= len(SQS) {
		nyys := make([]SQSymType, len(SQS)*2)
		copy(nyys, SQS)
		SQS = nyys
	}
	SQS[SQp] = SQVAL
	SQS[SQp].yys = SQstate

SQnewstate:
	SQn = SQPact[SQstate]
	if SQn <= SQFlag {
		goto SQdefault /* simple state */
	}
	if SQchar < 0 {
		SQchar = SQlex1(SQlex, &SQlval)
	}
	SQn += SQchar
	if SQn < 0 || SQn >= SQLast {
		goto SQdefault
	}
	SQn = SQAct[SQn]
	if SQChk[SQn] == SQchar { /* valid shift */
		SQchar = -1
		SQVAL = SQlval
		SQstate = SQn
		if Errflag > 0 {
			Errflag--
		}
		goto SQstack
	}

SQdefault:
	/* default state action */
	SQn = SQDef[SQstate]
	if SQn == -2 {
		if SQchar < 0 {
			SQchar = SQlex1(SQlex, &SQlval)
		}

		/* look through exception table */
		xi := 0
		for {
			if SQExca[xi+0] == -1 && SQExca[xi+1] == SQstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			SQn = SQExca[xi+0]
			if SQn < 0 || SQn == SQchar {
				break
			}
		}
		SQn = SQExca[xi+1]
		if SQn < 0 {
			goto ret0
		}
	}
	if SQn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			SQlex.Error("syntax error")
			Nerrs++
			if SQDebug >= 1 {
				__yyfmt__.Printf("%s", SQStatname(SQstate))
				__yyfmt__.Printf(" saw %s\n", SQTokname(SQchar))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for SQp >= 0 {
				SQn = SQPact[SQS[SQp].yys] + SQErrCode
				if SQn >= 0 && SQn < SQLast {
					SQstate = SQAct[SQn] /* simulate a shift of "error" */
					if SQChk[SQstate] == SQErrCode {
						goto SQstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if SQDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", SQS[SQp].yys)
				}
				SQp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if SQDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", SQTokname(SQchar))
			}
			if SQchar == SQEofCode {
				goto ret1
			}
			SQchar = -1
			goto SQnewstate /* try again in the same state */
		}
	}

	/* reduction by production SQn */
	if SQDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", SQn, SQStatname(SQstate))
	}

	SQnt := SQn
	SQpt := SQp
	_ = SQpt // guard against "declared and not used"

	SQp -= SQR2[SQn]
	SQVAL = SQS[SQp+1]

	/* consult goto table to find next state */
	SQn = SQR1[SQn]
	SQg := SQPgo[SQn]
	SQj := SQg + SQS[SQp].yys + 1

	if SQj >= SQLast {
		SQstate = SQAct[SQg]
	} else {
		SQstate = SQAct[SQj]
		if SQChk[SQstate] != -SQn {
			SQstate = SQAct[SQg]
		}
	}
	// dummy call; replaced with literal code
	switch SQnt {

	case 1:
		//line query.y:63
		{
			SQlex.(*SQLex).query.Contents = SQS[SQpt-2].list
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = SELECT_TYPE
		}
	case 2:
		//line query.y:69
		{
			SQlex.(*SQLex).query.Contents = SQS[SQpt-1].list
			SQlex.(*SQLex).query.qtype = SELECT_TYPE
		}
	case 3:
		//line query.y:74
		{
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.data = SQS[SQpt-2].data
			SQlex.(*SQLex).query.qtype = DATA_TYPE
		}
	case 4:
		//line query.y:80
		{
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.set = SQS[SQpt-2].dict
			SQlex.(*SQLex).query.qtype = SET_TYPE
		}
	case 5:
		//line query.y:86
		{
			SQlex.(*SQLex).query.set = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = SET_TYPE
		}
	case 6:
		//line query.y:91
		{
			SQlex.(*SQLex).query.Contents = SQS[SQpt-2].list
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = DELETE_TYPE
		}
	case 7:
		//line query.y:97
		{
			SQlex.(*SQLex).query.Contents = []string{}
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = DELETE_TYPE
		}
	case 8:
		//line query.y:103
		{
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.data = SQS[SQpt-2].data
			SQlex.(*SQLex).query.operators = SQS[SQpt-4].oplist
			SQlex.(*SQLex).query.qtype = APPLY_TYPE
		}
	case 9:
		//line query.y:112
		{
			SQVAL.list = List{SQS[SQpt-0].str}
		}
	case 10:
		//line query.y:116
		{
			SQVAL.list = append(List{SQS[SQpt-2].str}, SQS[SQpt-0].list...)
		}
	case 11:
		//line query.y:122
		{
			SQVAL.list = SQS[SQpt-1].list
		}
	case 12:
		//line query.y:127
		{
			SQVAL.list = List{SQS[SQpt-0].str}
		}
	case 13:
		//line query.y:131
		{
			SQVAL.list = append(List{SQS[SQpt-2].str}, SQS[SQpt-0].list...)
		}
	case 14:
		//line query.y:137
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 15:
		//line query.y:141
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 16:
		//line query.y:145
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].list}
		}
	case 17:
		//line query.y:149
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 18:
		//line query.y:154
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 19:
		//line query.y:159
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].list
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 20:
		//line query.y:166
		{
			SQVAL.list = SQS[SQpt-0].list
		}
	case 21:
		//line query.y:170
		{
			SQVAL.list = List{}
		}
	case 22:
		//line query.y:174
		{
			SQlex.(*SQLex).query.distinct = true
			SQVAL.list = List{SQS[SQpt-0].str}
		}
	case 23:
		//line query.y:179
		{
			SQlex.(*SQLex).query.distinct = true
			SQVAL.list = List{}
		}
	case 24:
		//line query.y:186
		{
			SQVAL.data = &dataquery{dtype: IN_TYPE, start: SQS[SQpt-5].time, end: SQS[SQpt-3].time, limit: SQS[SQpt-1].limit, timeconv: SQS[SQpt-0].timeconv}
		}
	case 25:
		//line query.y:190
		{
			SQVAL.data = &dataquery{dtype: IN_TYPE, start: SQS[SQpt-4].time, end: SQS[SQpt-2].time, limit: SQS[SQpt-1].limit, timeconv: SQS[SQpt-0].timeconv}
		}
	case 26:
		//line query.y:194
		{
			SQVAL.data = &dataquery{dtype: BEFORE_TYPE, start: SQS[SQpt-2].time, limit: SQS[SQpt-1].limit, timeconv: SQS[SQpt-0].timeconv}
		}
	case 27:
		//line query.y:198
		{
			SQVAL.data = &dataquery{dtype: AFTER_TYPE, start: SQS[SQpt-2].time, limit: SQS[SQpt-1].limit, timeconv: SQS[SQpt-0].timeconv}
		}
	case 28:
		//line query.y:204
		{
			SQVAL.time = SQS[SQpt-0].time
		}
	case 29:
		//line query.y:208
		{
			SQVAL.time = SQS[SQpt-1].time.Add(SQS[SQpt-0].timediff)
		}
	case 30:
		//line query.y:214
		{
			foundtime, err := parseAbsTime(SQS[SQpt-1].str, SQS[SQpt-0].str)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse time \"%v %v\" (%v)", SQS[SQpt-1].str, SQS[SQpt-0].str, err.Error()))
			}
			SQVAL.time = foundtime
		}
	case 31:
		//line query.y:222
		{
			num, err := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", SQS[SQpt-0].str, err.Error()))
			}
			SQVAL.time = _time.Unix(num, 0)
		}
	case 32:
		//line query.y:230
		{
			found := false
			for _, format := range supported_formats {
				t, err := _time.Parse(format, SQS[SQpt-0].str)
				if err != nil {
					continue
				}
				SQVAL.time = t
				found = true
				break
			}
			if !found {
				SQlex.(*SQLex).Error(fmt.Sprintf("No time format matching \"%v\" found", SQS[SQpt-0].str))
			}
		}
	case 33:
		//line query.y:246
		{
			SQVAL.time = _time.Now()
		}
	case 34:
		//line query.y:252
		{
			var err error
			SQVAL.timediff, err = parseReltime(SQS[SQpt-1].str, SQS[SQpt-0].str)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Error parsing relative time \"%v %v\" (%v)", SQS[SQpt-1].str, SQS[SQpt-0].str, err.Error()))
			}
		}
	case 35:
		//line query.y:260
		{
			newDuration, err := parseReltime(SQS[SQpt-2].str, SQS[SQpt-1].str)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Error parsing relative time \"%v %v\" (%v)", SQS[SQpt-2].str, SQS[SQpt-1].str, err.Error()))
			}
			SQVAL.timediff = addDurations(newDuration, SQS[SQpt-0].timediff)
		}
	case 36:
		//line query.y:270
		{
			SQVAL.limit = datalimit{limit: -1, streamlimit: -1}
		}
	case 37:
		//line query.y:274
		{
			num, err := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", SQS[SQpt-0].str, err.Error()))
			}
			SQVAL.limit = datalimit{limit: num, streamlimit: -1}
		}
	case 38:
		//line query.y:282
		{
			num, err := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", SQS[SQpt-0].str, err.Error()))
			}
			SQVAL.limit = datalimit{limit: -1, streamlimit: num}
		}
	case 39:
		//line query.y:290
		{
			limit_num, err := strconv.ParseInt(SQS[SQpt-2].str, 10, 64)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", SQS[SQpt-2].str, err.Error()))
			}
			slimit_num, err := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", SQS[SQpt-2].str, err.Error()))
			}
			SQVAL.limit = datalimit{limit: limit_num, streamlimit: slimit_num}
		}
	case 40:
		//line query.y:304
		{
			SQVAL.timeconv = UOT_MS
		}
	case 41:
		//line query.y:308
		{
			uot, err := parseUOT(SQS[SQpt-0].str)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse unit of time %v (%v)", SQS[SQpt-0].str, err))
			}
			SQVAL.timeconv = uot
		}
	case 42:
		//line query.y:320
		{
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 43:
		//line query.y:327
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: Dict{"$regex": SQS[SQpt-0].str}}
		}
	case 44:
		//line query.y:331
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 45:
		//line query.y:335
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 46:
		//line query.y:339
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: Dict{"$neq": SQS[SQpt-0].str}}
		}
	case 47:
		//line query.y:343
		{
			SQVAL.dict = Dict{SQS[SQpt-0].str: Dict{"$exists": true}}
		}
	case 48:
		//line query.y:347
		{
			SQVAL.dict = Dict{SQS[SQpt-0].str: Dict{"$in": SQS[SQpt-2].list}}
		}
	case 49:
		//line query.y:351
		{
			SQVAL.dict = Dict{SQS[SQpt-1].str: Dict{"$not": Dict{"$in": SQS[SQpt-3].list}}}
		}
	case 50:
		//line query.y:357
		{
			SQVAL.str = SQS[SQpt-0].str[1 : len(SQS[SQpt-0].str)-1]
		}
	case 51:
		//line query.y:363
		{

			SQlex.(*SQLex)._keys[SQS[SQpt-0].str] = struct{}{}
			SQVAL.str = cleantagstring(SQS[SQpt-0].str)
		}
	case 52:
		//line query.y:371
		{
			SQVAL.dict = Dict{"$and": []Dict{SQS[SQpt-2].dict, SQS[SQpt-0].dict}}
		}
	case 53:
		//line query.y:375
		{
			SQVAL.dict = Dict{"$or": []Dict{SQS[SQpt-2].dict, SQS[SQpt-0].dict}}
		}
	case 54:
		//line query.y:379
		{
			tmp := make(Dict)
			for k, v := range SQS[SQpt-0].dict {
				tmp[k] = Dict{"$ne": v}
			}
			SQVAL.dict = tmp
		}
	case 55:
		//line query.y:387
		{
			SQVAL.dict = SQS[SQpt-1].dict
		}
	case 56:
		//line query.y:391
		{
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 57:
		//line query.y:397
		{
			SQVAL.oplist = []*OpNode{SQS[SQpt-0].op}
		}
	case 58:
		//line query.y:401
		{
			SQVAL.oplist = append(SQS[SQpt-0].oplist, SQS[SQpt-2].op)
		}
	case 59:
		//line query.y:407
		{
			SQVAL.op = &OpNode{Operator: SQS[SQpt-2].str}
		}
	case 60:
		//line query.y:411
		{
			SQVAL.op = &OpNode{Operator: SQS[SQpt-3].str, Arguments: SQS[SQpt-1].dict}
		}
	case 61:
		//line query.y:417
		{
			fmt.Printf("op args %v %v\n", SQS[SQpt-2].str, SQS[SQpt-0].str)
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 62:
		//line query.y:422
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 63:
		//line query.y:426
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 64:
		//line query.y:431
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	}
	goto SQstack /* stack new state and value */
}
