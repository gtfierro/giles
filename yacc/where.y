%{

package main

import (
	"bufio"
	"fmt"
	"os"
    "github.com/taylorchu/toki"
)


/**
Notes here
**/
%}

%union{
    str string
    dict Dict
}

%token <str> WHERE
%token <str> LVALUE QSTRING QREGEX
%token <str> EQ NEQ
%token <str> LIKE
%token <str> AND OR HAS NOT
%token <str> LPAREN RPAREN
%token NUMBER
%token SEMICOLON
%token NEWLINE

%type <dict> whereList
%type <dict> whereTerm
%type <str> NUMBER
%type <str> SEMICOLON NEWLINE

%right EQ

%%

whereClause : WHERE whereList SEMICOLON
            {
              fmt.Printf("finished: %v \n", $2)
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

type SQLex struct {
    query string
    scanner *toki.Scanner
    pos int
}

const (
    S_NUMBER toki.Token = iota + 1
    S_LVALUE
    S_EQ
    S_EOF
)

func NewSQLex(s string) *SQLex {
    scanner := toki.NewScanner(
        []toki.Def{
            {Token: WHERE, Pattern: "where"},
            {Token: AND, Pattern: "and"},
            {Token: OR, Pattern: "or"},
            {Token: HAS, Pattern: "has"},
            {Token: NOT, Pattern: "not"},
            {Token: NEQ, Pattern: "!="},
            {Token: EQ, Pattern: "="},
            {Token: LPAREN, Pattern: "\\("},
            {Token: RPAREN, Pattern: "\\)"},
            {Token: SEMICOLON, Pattern: ";"},
            {Token: NEWLINE, Pattern: "\n"},
            {Token: LIKE, Pattern: "(like)|~"},
            {Token: LVALUE, Pattern: "[a-zA-Z\\~\\$\\_][a-zA-Z0-9\\/\\%_\\-]*"},
            {Token: QSTRING, Pattern: "(\"[^\"\\\\]*?(\\.[^\"\\\\]*?)*?\")|('[^'\\\\]*?(\\.[^'\\\\]*?)*?')"},
            {Token: QREGEX, Pattern: "%?[a-zA-Z0-9]+%?"},
            {Token: NUMBER, Pattern: "([+-]?([0-9]*\\.)?[0-9]+)"},
        })
    scanner.SetInput(s)
    return &SQLex{query: s, scanner: scanner}
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
        if query, ok := readline(fi); ok {
            SQParse(NewSQLex(query))
        } else {
            break
        }
    }
}
