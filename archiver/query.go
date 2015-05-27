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
const NUMBER = 57376
const SEMICOLON = 57377
const NEWLINE = 57378
const TIMEUNIT = 57379

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
	"NUMBER",
	"SEMICOLON",
	"NEWLINE",
	"TIMEUNIT",
}
var SQStatenames = []string{}

const SQEofCode = 1
const SQErrCode = 2
const SQMaxDepth = 200

//line query.y:394
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

// Parse has been moved to query_processor.go

//line yacctab:1
var SQExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const SQNprod = 56
const SQPrivate = 57344

var SQTokenNames []string
var SQStates []string

const SQLast = 138

var SQAct = []int{

	93, 74, 19, 45, 67, 48, 71, 38, 17, 14,
	12, 15, 12, 34, 7, 21, 23, 107, 24, 73,
	100, 50, 49, 29, 50, 32, 49, 18, 50, 50,
	51, 52, 18, 57, 50, 54, 12, 55, 102, 44,
	43, 47, 41, 42, 33, 47, 85, 65, 69, 60,
	61, 56, 112, 30, 77, 66, 58, 59, 22, 114,
	72, 96, 95, 82, 88, 40, 26, 27, 83, 84,
	86, 94, 80, 81, 91, 87, 8, 101, 97, 111,
	92, 16, 58, 59, 13, 25, 15, 15, 98, 99,
	110, 90, 79, 105, 104, 103, 39, 35, 78, 106,
	36, 70, 63, 64, 109, 53, 113, 62, 10, 28,
	89, 31, 50, 11, 117, 115, 116, 18, 118, 13,
	68, 20, 13, 108, 13, 9, 75, 76, 2, 11,
	4, 3, 5, 1, 18, 46, 6, 37,
}
var SQPact = []int{

	124, -1000, 103, 106, 108, 105, 23, 125, -1000, -1000,
	106, 55, 88, -1000, 18, 92, 125, 9, 68, 34,
	10, 8, -1000, 5, -1000, 7, 11, 11, 106, 0,
	-1000, 17, -2, -1000, 56, 68, 68, -1000, 83, 106,
	119, 104, -1000, -1000, 11, 80, 26, -18, -1000, -1000,
	-1000, 113, 113, -1000, -1000, 77, 71, -1000, 68, 68,
	56, 30, 95, 12, 95, -1000, 125, 31, 91, 70,
	11, -1000, 106, -1000, 46, 28, 27, 46, 106, 106,
	56, 56, -1000, -1000, -1000, -1000, -1000, -15, 54, 4,
	11, 113, 26, -1000, -20, 109, -1000, -1000, -1000, -1000,
	-1000, 105, 69, 58, 19, 46, -1000, -1000, 25, -1000,
	104, 104, 113, -1000, -1000, -1000, -1000, 46, -1000,
}
var SQPgo = []int{

	0, 13, 137, 8, 9, 4, 136, 76, 2, 14,
	3, 135, 6, 1, 0, 5, 7, 133,
}
var SQR1 = []int{

	0, 17, 17, 17, 17, 17, 17, 17, 17, 7,
	7, 4, 4, 4, 4, 6, 6, 6, 6, 9,
	9, 9, 9, 10, 10, 11, 11, 11, 11, 12,
	12, 13, 13, 13, 13, 14, 14, 3, 2, 2,
	2, 2, 2, 15, 16, 1, 1, 1, 1, 1,
	8, 8, 5, 5, 5, 5,
}
var SQR2 = []int{

	0, 4, 3, 4, 4, 3, 4, 3, 6, 1,
	3, 3, 3, 5, 5, 1, 1, 2, 1, 9,
	7, 5, 5, 1, 2, 2, 1, 1, 1, 2,
	3, 0, 2, 2, 4, 0, 2, 2, 3, 3,
	3, 3, 2, 1, 1, 3, 3, 2, 3, 1,
	4, 6, 3, 3, 5, 5,
}
var SQChk = []int{

	-1000, -17, 4, 7, 6, 8, -6, -9, -7, 22,
	5, 10, -16, 16, -4, -16, -7, -3, 9, -8,
	16, -3, 35, -3, -16, 30, 11, 12, 21, -3,
	35, 19, -3, 35, -1, 29, 32, -2, -16, 28,
	31, 32, 35, 35, 32, -10, -11, 34, -15, 15,
	17, -10, -10, -7, 35, -15, 34, 35, 26, 27,
	-1, -1, 24, 19, 20, -16, -9, -5, 16, -10,
	21, -12, 34, 37, -13, 13, 14, -13, 21, 21,
	-1, -1, 33, -15, -15, 34, -15, -3, 33, 19,
	21, -10, -16, -14, 25, 34, 34, -14, -4, -4,
	35, 23, 34, -15, -10, -13, -12, 37, 14, -8,
	21, 21, 33, -14, 34, -5, -5, -13, -14,
}
var SQDef = []int{

	0, -2, 0, 0, 0, 0, 0, 0, 15, 16,
	18, 0, 9, 44, 0, 0, 0, 0, 0, 0,
	0, 0, 2, 0, 17, 0, 0, 0, 0, 0,
	5, 0, 0, 7, 37, 0, 0, 49, 0, 0,
	0, 0, 1, 3, 0, 0, 23, 26, 27, 28,
	43, 31, 31, 10, 4, 11, 12, 6, 0, 0,
	47, 0, 0, 0, 0, 42, 0, 0, 0, 0,
	0, 24, 0, 25, 35, 0, 0, 35, 0, 0,
	45, 46, 48, 38, 39, 40, 41, 0, 50, 0,
	0, 31, 29, 21, 0, 32, 33, 22, 13, 14,
	8, 0, 52, 53, 0, 35, 30, 36, 0, 51,
	0, 0, 31, 20, 34, 54, 55, 35, 19,
}
var SQTok1 = []int{

	1,
}
var SQTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37,
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
		//line query.y:61
		{
			SQlex.(*SQLex).query.Contents = SQS[SQpt-2].list
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = SELECT_TYPE
		}
	case 2:
		//line query.y:67
		{
			SQlex.(*SQLex).query.Contents = SQS[SQpt-1].list
			SQlex.(*SQLex).query.qtype = SELECT_TYPE
		}
	case 3:
		//line query.y:72
		{
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.data = SQS[SQpt-2].data
			SQlex.(*SQLex).query.qtype = DATA_TYPE
		}
	case 4:
		//line query.y:78
		{
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.set = SQS[SQpt-2].dict
			SQlex.(*SQLex).query.qtype = SET_TYPE
		}
	case 5:
		//line query.y:84
		{
			SQlex.(*SQLex).query.set = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = SET_TYPE
		}
	case 6:
		//line query.y:89
		{
			SQlex.(*SQLex).query.Contents = SQS[SQpt-2].list
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = DELETE_TYPE
		}
	case 7:
		//line query.y:95
		{
			SQlex.(*SQLex).query.Contents = []string{}
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = DELETE_TYPE
		}
	case 8:
		//line query.y:101
		{
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.data = SQS[SQpt-2].data
			SQlex.(*SQLex).query.operators = SQS[SQpt-4].oplist
			SQlex.(*SQLex).query.qtype = APPLY_TYPE
		}
	case 9:
		//line query.y:110
		{
			SQVAL.list = List{SQS[SQpt-0].str}
		}
	case 10:
		//line query.y:114
		{
			SQVAL.list = append(List{SQS[SQpt-2].str}, SQS[SQpt-0].list...)
		}
	case 11:
		//line query.y:120
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 12:
		//line query.y:124
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 13:
		//line query.y:128
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 14:
		//line query.y:133
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 15:
		//line query.y:140
		{
			SQVAL.list = SQS[SQpt-0].list
		}
	case 16:
		//line query.y:144
		{
			SQVAL.list = List{}
		}
	case 17:
		//line query.y:148
		{
			SQlex.(*SQLex).query.distinct = true
			SQVAL.list = List{SQS[SQpt-0].str}
		}
	case 18:
		//line query.y:153
		{
			SQlex.(*SQLex).query.distinct = true
			SQVAL.list = List{}
		}
	case 19:
		//line query.y:160
		{
			SQVAL.data = &dataquery{dtype: IN_TYPE, start: SQS[SQpt-5].time, end: SQS[SQpt-3].time, limit: SQS[SQpt-1].limit, timeconv: SQS[SQpt-0].timeconv}
		}
	case 20:
		//line query.y:164
		{
			SQVAL.data = &dataquery{dtype: IN_TYPE, start: SQS[SQpt-4].time, end: SQS[SQpt-2].time, limit: SQS[SQpt-1].limit, timeconv: SQS[SQpt-0].timeconv}
		}
	case 21:
		//line query.y:168
		{
			SQVAL.data = &dataquery{dtype: BEFORE_TYPE, start: SQS[SQpt-2].time, limit: SQS[SQpt-1].limit, timeconv: SQS[SQpt-0].timeconv}
		}
	case 22:
		//line query.y:172
		{
			SQVAL.data = &dataquery{dtype: AFTER_TYPE, start: SQS[SQpt-2].time, limit: SQS[SQpt-1].limit, timeconv: SQS[SQpt-0].timeconv}
		}
	case 23:
		//line query.y:178
		{
			SQVAL.time = SQS[SQpt-0].time
		}
	case 24:
		//line query.y:182
		{
			SQVAL.time = SQS[SQpt-1].time.Add(SQS[SQpt-0].timediff)
		}
	case 25:
		//line query.y:188
		{
			foundtime, err := parseAbsTime(SQS[SQpt-1].str, SQS[SQpt-0].str)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse time \"%v %v\" (%v)", SQS[SQpt-1].str, SQS[SQpt-0].str, err.Error()))
			}
			SQVAL.time = foundtime
		}
	case 26:
		//line query.y:196
		{
			num, err := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", SQS[SQpt-0].str, err.Error()))
			}
			SQVAL.time = _time.Unix(num, 0)
		}
	case 27:
		//line query.y:204
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
	case 28:
		//line query.y:220
		{
			SQVAL.time = _time.Now()
		}
	case 29:
		//line query.y:226
		{
			var err error
			SQVAL.timediff, err = parseReltime(SQS[SQpt-1].str, SQS[SQpt-0].str)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Error parsing relative time \"%v %v\" (%v)", SQS[SQpt-1].str, SQS[SQpt-0].str, err.Error()))
			}
		}
	case 30:
		//line query.y:234
		{
			newDuration, err := parseReltime(SQS[SQpt-2].str, SQS[SQpt-1].str)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Error parsing relative time \"%v %v\" (%v)", SQS[SQpt-2].str, SQS[SQpt-1].str, err.Error()))
			}
			SQVAL.timediff = addDurations(newDuration, SQS[SQpt-0].timediff)
		}
	case 31:
		//line query.y:244
		{
			SQVAL.limit = datalimit{limit: -1, streamlimit: -1}
		}
	case 32:
		//line query.y:248
		{
			num, err := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", SQS[SQpt-0].str, err.Error()))
			}
			SQVAL.limit = datalimit{limit: num, streamlimit: -1}
		}
	case 33:
		//line query.y:256
		{
			num, err := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse integer \"%v\" (%v)", SQS[SQpt-0].str, err.Error()))
			}
			SQVAL.limit = datalimit{limit: -1, streamlimit: num}
		}
	case 34:
		//line query.y:264
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
	case 35:
		//line query.y:278
		{
			SQVAL.timeconv = UOT_MS
		}
	case 36:
		//line query.y:282
		{
			uot, err := parseUOT(SQS[SQpt-0].str)
			if err != nil {
				SQlex.(*SQLex).Error(fmt.Sprintf("Could not parse unit of time %v (%v)", SQS[SQpt-0].str, err))
			}
			SQVAL.timeconv = uot
		}
	case 37:
		//line query.y:294
		{
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 38:
		//line query.y:301
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: Dict{"$regex": SQS[SQpt-0].str}}
		}
	case 39:
		//line query.y:305
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 40:
		//line query.y:309
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 41:
		//line query.y:313
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: Dict{"$neq": SQS[SQpt-0].str}}
		}
	case 42:
		//line query.y:317
		{
			SQVAL.dict = Dict{SQS[SQpt-0].str: Dict{"$exists": true}}
		}
	case 43:
		//line query.y:323
		{
			SQVAL.str = SQS[SQpt-0].str[1 : len(SQS[SQpt-0].str)-1]
		}
	case 44:
		//line query.y:329
		{

			SQlex.(*SQLex)._keys[SQS[SQpt-0].str] = struct{}{}
			SQVAL.str = cleantagstring(SQS[SQpt-0].str)
		}
	case 45:
		//line query.y:337
		{
			SQVAL.dict = Dict{"$and": []Dict{SQS[SQpt-2].dict, SQS[SQpt-0].dict}}
		}
	case 46:
		//line query.y:341
		{
			SQVAL.dict = Dict{"$or": []Dict{SQS[SQpt-2].dict, SQS[SQpt-0].dict}}
		}
	case 47:
		//line query.y:345
		{
			tmp := make(Dict)
			for k, v := range SQS[SQpt-0].dict {
				tmp[k] = Dict{"$ne": v}
			}
			SQVAL.dict = tmp
		}
	case 48:
		//line query.y:353
		{
			SQVAL.dict = SQS[SQpt-1].dict
		}
	case 49:
		//line query.y:357
		{
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 50:
		//line query.y:363
		{
			o := &OpNode{Operator: SQS[SQpt-3].str, Arguments: SQS[SQpt-1].dict}
			SQVAL.oplist = []*OpNode{o}
		}
	case 51:
		//line query.y:368
		{
			o := &OpNode{Operator: SQS[SQpt-5].str, Arguments: SQS[SQpt-3].dict}
			SQVAL.oplist = append(SQS[SQpt-0].oplist, o)
		}
	case 52:
		//line query.y:375
		{
			fmt.Printf("op args %v %v\n", SQS[SQpt-2].str, SQS[SQpt-0].str)
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 53:
		//line query.y:380
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 54:
		//line query.y:384
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 55:
		//line query.y:389
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	}
	goto SQstack /* stack new state and value */
}
