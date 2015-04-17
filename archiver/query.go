//line query.y:2
package archiver

import __yyfmt__ "fmt"

//line query.y:3
import (
	"bufio"
	"fmt"
	"github.com/taylorchu/toki"
	"gopkg.in/mgo.v2/bson"
	"os"
	"strconv"
	_time "time"
)

/**
Notes here
**/

//line query.y:21
type SQSymType struct {
	yys      int
	str      string
	dict     Dict
	data     *dataquery
	limit    datalimit
	list     List
	time     _time.Time
	timediff _time.Duration
}

const SELECT = 57346
const DISTINCT = 57347
const DELETE = 57348
const SET = 57349
const WHERE = 57350
const DATA = 57351
const BEFORE = 57352
const AFTER = 57353
const LIMIT = 57354
const STREAMLIMIT = 57355
const NOW = 57356
const LVALUE = 57357
const QSTRING = 57358
const QREGEX = 57359
const EQ = 57360
const NEQ = 57361
const COMMA = 57362
const ALL = 57363
const LIKE = 57364
const AND = 57365
const OR = 57366
const HAS = 57367
const NOT = 57368
const IN = 57369
const LPAREN = 57370
const RPAREN = 57371
const NUMBER = 57372
const SEMICOLON = 57373
const NEWLINE = 57374

var SQToknames = []string{
	"SELECT",
	"DISTINCT",
	"DELETE",
	"SET",
	"WHERE",
	"DATA",
	"BEFORE",
	"AFTER",
	"LIMIT",
	"STREAMLIMIT",
	"NOW",
	"LVALUE",
	"QSTRING",
	"QREGEX",
	"EQ",
	"NEQ",
	"COMMA",
	"ALL",
	"LIKE",
	"AND",
	"OR",
	"HAS",
	"NOT",
	"IN",
	"LPAREN",
	"RPAREN",
	"NUMBER",
	"SEMICOLON",
	"NEWLINE",
}
var SQStatenames = []string{}

const SQEofCode = 1
const SQErrCode = 2
const SQMaxDepth = 200

//line query.y:292
const eof = 0

var supported_formats = []string{"1/2/2006",
	"1-2-2006",
	"1/2/2006 04:15",
	"1-2-2006 04:15",
	"2006-1-2 15:04:05"}

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
	contents []string
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

func (q *query) ContentsBson() bson.M {
	ret := bson.M{}
	for _, tag := range q.contents {
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
	dtype dataqueryType
	start _time.Time
	end   _time.Time
	limit datalimit
}

type datalimit struct {
	limit       int64
	streamlimit int64
}

type SQLex struct {
	querystring string
	query       *query
	scanner     *toki.Scanner
	syntaxError bool
	lasttoken   string
	tokens      []string
	// all keys that we encounter. Used for republish concerns
	_keys map[string]struct{}
	keys  []string
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
	return &SQLex{query: q, querystring: s, scanner: scanner, syntaxError: false, lasttoken: "", _keys: map[string]struct{}{}, tokens: []string{}}
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
	sq.syntaxError = true
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

func Parse(querystring string) *SQLex {
	l := NewSQLex(querystring)
	SQParse(l)
	l.query.Print()
	l.keys = make([]string, len(l._keys))
	i := 0
	for key, _ := range l._keys {
		l.keys[i] = key
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

//line yacctab:1
var SQExca = []int{
	-1, 1,
	1, -1,
	-2, 0,
}

const SQNprod = 45
const SQPrivate = 57344

var SQTokenNames []string
var SQStates []string

const SQLast = 105

var SQAct = []int{

	63, 61, 38, 34, 41, 13, 11, 14, 11, 43,
	18, 18, 30, 20, 42, 42, 43, 43, 51, 52,
	50, 43, 47, 74, 71, 44, 45, 36, 11, 37,
	40, 40, 48, 26, 17, 49, 88, 12, 29, 58,
	59, 62, 80, 79, 53, 54, 66, 35, 31, 87,
	32, 22, 23, 51, 52, 56, 57, 27, 76, 55,
	68, 73, 75, 77, 69, 70, 78, 67, 21, 60,
	7, 14, 14, 81, 82, 15, 24, 9, 84, 83,
	85, 10, 16, 72, 43, 12, 86, 12, 89, 19,
	64, 65, 18, 8, 1, 46, 25, 2, 28, 4,
	3, 39, 6, 5, 33,
}
var SQPact = []int{

	93, -1000, 72, 70, 70, 3, 84, -1000, -1000, 70,
	41, 56, -1000, 2, 39, 84, 7, -1000, 22, -4,
	-1000, 1, 0, 0, 70, -9, -1000, 5, -11, -1000,
	30, 22, 22, -1000, 37, 70, -1000, 0, 49, 11,
	-1000, -1000, -1000, -1000, 78, 78, -1000, -1000, 47, 40,
	-1000, 22, 22, 30, -5, 66, -7, 68, -1000, 38,
	0, -1000, 70, -1000, 13, 12, -1000, 70, 70, 30,
	30, -1000, -1000, -1000, -1000, -1000, 0, 78, 11, 73,
	-1000, -1000, -1000, 20, -1000, -1000, 6, 78, -1000, -1000,
}
var SQPgo = []int{

	0, 12, 104, 82, 5, 103, 70, 102, 2, 101,
	1, 0, 4, 3, 94,
}
var SQR1 = []int{

	0, 14, 14, 14, 14, 14, 14, 6, 6, 4,
	4, 4, 4, 5, 5, 5, 5, 7, 7, 7,
	7, 8, 8, 9, 9, 9, 10, 10, 11, 11,
	11, 11, 3, 2, 2, 2, 2, 2, 12, 13,
	1, 1, 1, 1, 1,
}
var SQR2 = []int{

	0, 4, 3, 4, 4, 3, 4, 1, 3, 3,
	3, 5, 5, 1, 1, 2, 1, 8, 6, 4,
	4, 1, 2, 1, 1, 1, 2, 3, 0, 2,
	2, 4, 2, 3, 3, 3, 3, 2, 1, 1,
	3, 3, 2, 3, 1,
}
var SQChk = []int{

	-1000, -14, 4, 7, 6, -5, -7, -6, 21, 5,
	9, -13, 15, -4, -13, -6, -3, 31, 8, -3,
	-13, 27, 10, 11, 20, -3, 31, 18, -3, 31,
	-1, 26, 28, -2, -13, 25, 31, 28, -8, -9,
	30, -12, 14, 16, -8, -8, -6, 31, -12, 30,
	31, 23, 24, -1, -1, 22, 18, 19, -13, -8,
	20, -10, 30, -11, 12, 13, -11, 20, 20, -1,
	-1, 29, 17, -12, 30, -12, 20, -8, -13, 30,
	30, -4, -4, -8, -11, -10, 13, 29, 30, -11,
}
var SQDef = []int{

	0, -2, 0, 0, 0, 0, 0, 13, 14, 16,
	0, 7, 39, 0, 0, 0, 0, 2, 0, 0,
	15, 0, 0, 0, 0, 0, 5, 0, 0, 1,
	32, 0, 0, 44, 0, 0, 3, 0, 0, 21,
	23, 24, 25, 38, 28, 28, 8, 4, 9, 10,
	6, 0, 0, 42, 0, 0, 0, 0, 37, 0,
	0, 22, 0, 19, 0, 0, 20, 0, 0, 40,
	41, 43, 33, 34, 35, 36, 0, 28, 26, 29,
	30, 11, 12, 0, 18, 27, 0, 28, 31, 17,
}
var SQTok1 = []int{

	1,
}
var SQTok2 = []int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32,
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
		//line query.y:57
		{
			SQlex.(*SQLex).query.contents = SQS[SQpt-2].list
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = SELECT_TYPE
		}
	case 2:
		//line query.y:63
		{
			SQlex.(*SQLex).query.contents = SQS[SQpt-1].list
			SQlex.(*SQLex).query.qtype = SELECT_TYPE
		}
	case 3:
		//line query.y:68
		{
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.data = SQS[SQpt-2].data
			SQlex.(*SQLex).query.qtype = DATA_TYPE
		}
	case 4:
		//line query.y:74
		{
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.set = SQS[SQpt-2].dict
			SQlex.(*SQLex).query.qtype = SET_TYPE
		}
	case 5:
		//line query.y:80
		{
			SQlex.(*SQLex).query.set = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = SET_TYPE
		}
	case 6:
		//line query.y:85
		{
			SQlex.(*SQLex).query.contents = SQS[SQpt-2].list
			SQlex.(*SQLex).query.where = SQS[SQpt-1].dict
			SQlex.(*SQLex).query.qtype = DELETE_TYPE
		}
	case 7:
		//line query.y:93
		{
			SQVAL.list = List{SQS[SQpt-0].str}
		}
	case 8:
		//line query.y:97
		{
			SQVAL.list = append(List{SQS[SQpt-2].str}, SQS[SQpt-0].list...)
		}
	case 9:
		//line query.y:103
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 10:
		//line query.y:107
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 11:
		//line query.y:111
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 12:
		//line query.y:116
		{
			SQS[SQpt-0].dict[SQS[SQpt-4].str] = SQS[SQpt-2].str
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 13:
		//line query.y:123
		{
			SQVAL.list = SQS[SQpt-0].list
		}
	case 14:
		//line query.y:127
		{
			SQVAL.list = List{}
		}
	case 15:
		//line query.y:131
		{
			SQlex.(*SQLex).query.distinct = true
			SQVAL.list = List{SQS[SQpt-0].str}
		}
	case 16:
		//line query.y:136
		{
			SQlex.(*SQLex).query.distinct = true
			SQVAL.list = List{}
		}
	case 17:
		//line query.y:143
		{
			SQVAL.data = &dataquery{dtype: IN_TYPE, start: SQS[SQpt-4].time, end: SQS[SQpt-2].time, limit: SQS[SQpt-0].limit}
		}
	case 18:
		//line query.y:147
		{
			SQVAL.data = &dataquery{dtype: IN_TYPE, start: SQS[SQpt-3].time, end: SQS[SQpt-1].time, limit: SQS[SQpt-0].limit}
		}
	case 19:
		//line query.y:151
		{
			SQVAL.data = &dataquery{dtype: BEFORE_TYPE, start: SQS[SQpt-1].time, limit: SQS[SQpt-0].limit}
		}
	case 20:
		//line query.y:155
		{
			SQVAL.data = &dataquery{dtype: AFTER_TYPE, start: SQS[SQpt-1].time, limit: SQS[SQpt-0].limit}
		}
	case 21:
		//line query.y:161
		{
			SQVAL.time = SQS[SQpt-0].time
		}
	case 22:
		//line query.y:165
		{
			SQVAL.time = SQS[SQpt-1].time.Add(SQS[SQpt-0].timediff)
		}
	case 23:
		//line query.y:171
		{
			//TODO: handle error?
			num, _ := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			SQVAL.time = _time.Unix(num, 0)
		}
	case 24:
		//line query.y:177
		{
			for _, format := range supported_formats {
				t, err := _time.Parse(format, SQS[SQpt-0].str)
				if err != nil {
					continue
				}
				SQVAL.time = t
				break
			}
		}
	case 25:
		//line query.y:188
		{
			SQVAL.time = _time.Now()
		}
	case 26:
		//line query.y:194
		{
			SQVAL.timediff, _ = parseReltime(SQS[SQpt-1].str, SQS[SQpt-0].str)
		}
	case 27:
		//line query.y:198
		{
			newDuration, _ := parseReltime(SQS[SQpt-2].str, SQS[SQpt-1].str)
			SQVAL.timediff = addDurations(newDuration, SQS[SQpt-0].timediff)
		}
	case 28:
		//line query.y:205
		{
			SQVAL.limit = datalimit{limit: -1, streamlimit: -1}
		}
	case 29:
		//line query.y:209
		{
			num, _ := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			SQVAL.limit = datalimit{limit: num, streamlimit: -1}
		}
	case 30:
		//line query.y:214
		{
			num, _ := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			SQVAL.limit = datalimit{limit: -1, streamlimit: num}
		}
	case 31:
		//line query.y:219
		{
			limit_num, _ := strconv.ParseInt(SQS[SQpt-2].str, 10, 64)
			slimit_num, _ := strconv.ParseInt(SQS[SQpt-0].str, 10, 64)
			SQVAL.limit = datalimit{limit: limit_num, streamlimit: slimit_num}
		}
	case 32:
		//line query.y:228
		{
			SQVAL.dict = SQS[SQpt-0].dict
		}
	case 33:
		//line query.y:235
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: Dict{"$like": SQS[SQpt-0].str}}
		}
	case 34:
		//line query.y:239
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 35:
		//line query.y:243
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: SQS[SQpt-0].str}
		}
	case 36:
		//line query.y:247
		{
			SQVAL.dict = Dict{SQS[SQpt-2].str: Dict{"$neq": SQS[SQpt-0].str}}
		}
	case 37:
		//line query.y:251
		{
			SQVAL.dict = Dict{SQS[SQpt-0].str: Dict{"$exists": true}}
		}
	case 38:
		//line query.y:257
		{
			SQVAL.str = SQS[SQpt-0].str[1 : len(SQS[SQpt-0].str)-1]
		}
	case 39:
		//line query.y:263
		{

			SQlex.(*SQLex)._keys[SQS[SQpt-0].str] = struct{}{}
			SQVAL.str = cleantagstring(SQS[SQpt-0].str)
		}
	case 40:
		//line query.y:271
		{
			SQVAL.dict = Dict{"$and": []Dict{SQS[SQpt-2].dict, SQS[SQpt-0].dict}}
		}
	case 41:
		//line query.y:275
		{
			SQVAL.dict = Dict{"$or": []Dict{SQS[SQpt-2].dict, SQS[SQpt-0].dict}}
		}
	case 42:
		//line query.y:279
		{
			SQVAL.dict = Dict{"$not": SQS[SQpt-0].dict} // fix this to negate all items in $2
		}
	case 43:
		//line query.y:283
		{
			SQVAL.dict = SQS[SQpt-1].dict
		}
	case 44:
		//line query.y:287
		{
			SQVAL.dict = SQS[SQpt-0].dict
		}
	}
	goto SQstack /* stack new state and value */
}
