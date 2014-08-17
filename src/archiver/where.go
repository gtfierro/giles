package main

import (
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"log"
	"strings"
)

/*
  appends stringified token to list of tokens,
  then empties token
*/
func addtoken(tokens *[]string, token *[]rune) {
	if len(*token) > 0 {
		(*tokens) = append((*tokens), string(*token))
		(*token) = []rune{}
	}
}

/*
  returns a slice of tokens, inserting whitespace where necessary
*/
func tokenize(q string) []string {
	if !strings.HasSuffix(q, ";") {
		q += ";"
	}
	var tokens []string
	var token []rune

	inquotes := false

	pos := 0
	for {
		if pos == len(q) {
			break
		}
		char := rune(q[pos])
		switch char {
		case '\'':
			inquotes = !inquotes
		case '!', '~', '=':
			if !inquotes {
				addtoken(&tokens, &token)
				tokens = append(tokens, string(char))
			}
		case ';', ' ':
			if !inquotes {
				addtoken(&tokens, &token)
			} else {
				token = append(token, char)
			}
		default:
			token = append(token, char)
		}
		pos++
	}

	return tokens
}

func makebson(tokens []string) (*bson.M, error) {
	var querytype string
	var where bson.M
	var iswhere bool = false
	var contents []string
	var token string
	var tag, predicate, value string
	var negate bool = false
	switch tokens[0] {
	case "select":
		querytype = "select"
	case "delete":
		querytype = "delete"
	default:
		return &where, errors.New("Query must be select or delete")
	}
	fmt.Println("query type:", querytype)
	/* the contents of the select/delete */
	pos := 1
Token:
	for {
		if pos == len(tokens) {
			// no where clause! do EVERYTHING
			where = nil
			iswhere = false
			break
		}
		token = tokens[pos]
		switch token {
		case "where": // end of contents
			iswhere = true
			break Token // break out of for loop
		default:
			// adds the token to the list of contents,
			// removing a trailing comma if there is one
			contents = append(contents, strings.TrimSuffix(token, ","))
		}
		pos++
	}
	// don't bother parsing the rest if there isn't any
	if !iswhere {
		goto Return
	}
	/* construct bson for where clause */
	// increment past 'where' token
	pos++
	for {
		negate = false // reinitialize
		if pos == len(tokens) {
			break
		}
		token = tokens[pos]
		if token == "and" || token == "or" {
			pos++
		}

		tag = tokens[pos]
		tag = strings.Replace(tag, "/", ".", -1)
		pos++

		predicate = tokens[pos]
		if predicate == "!" { //this is a composite predicate, so we grab the next token, too
			negate = true
			pos++
			predicate += tokens[pos]
		}
		pos++

		value = tokens[pos]

		term := bson.M{tag: value}
		if negate {
			term = bson.M{"$not": term}
		}

		where = term

		pos++
	}

Return:
	return &where, nil
}

func parse(q string) *bson.M {
	fmt.Println(q)
	tokens := tokenize(q)
	query, err := makebson(tokens)
	if err != nil {
		log.Panic(err)
	}
	return query
}
