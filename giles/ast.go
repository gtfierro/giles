package main

import (
	"errors"
	"gopkg.in/mgo.v2/bson"
	"strconv"
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
		case '\'', '"':
			inquotes = !inquotes
		case ',':
			token = append(token, char)
			if !inquotes {
				addtoken(&tokens, &token)
			}
		case '!':
			if !inquotes {
				addtoken(&tokens, &token)
			}
			token = append(token, char)
		case '~', '=':
			if !inquotes {
				if len(token) > 0 && token[0] == '!' {
					token = append(token, char)
					addtoken(&tokens, &token)
				} else {
					addtoken(&tokens, &token)
					tokens = append(tokens, string(char))
				}
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

/**
 * Handles parsing the data range queries like "data in (start ref, end ref) [limit]"
**/
func parseDataTarget(tokens *[]string) Target_T {
	var dt = &DataTarget{Streamlimit: -1, Limit: 1}
	if len(*tokens) == 0 {
		return dt
	}
	// pos = 0 is the word 'data', pos = 1 is our dataquery type
	switch (*tokens)[1] {
	case "in":
		dt.Type = IN
	case "before":
		dt.Type = BEFORE
	case "after":
		dt.Type = AFTER
	default:
		log.Panic("Invalid data query", (*tokens)[1])
		return dt
	}
	pos := 2
	timetokens := []string{}
	for {
		if pos >= len(*tokens) {
			break
		}
		val := (*tokens)[pos]
		switch val {
		case "limit":
			limit, err := strconv.ParseUint((*tokens)[pos+1], 10, 64)
			if err != nil {
				log.Panic(err)
			}
			dt.Limit = uint32(limit)
			pos += 2
			continue
		case "streamlimit":
			limit, err := strconv.ParseInt((*tokens)[pos+1], 10, 64)
			if err != nil {
				log.Panic(err)
			}
			dt.Streamlimit = int(limit)
			pos += 2
			continue
		case "where": // terminating cases
			(*tokens) = (*tokens)[pos+1:]
			goto ReturnDataTarget
		default: // parse a time specification
			timetokens = append(timetokens, val)
			if strings.HasSuffix(val, ",") || strings.HasSuffix(val, ")") {
				time, err := handleTime(timetokens)
				if err != nil {
					log.Panic(err)
				}
				switch dt.Type {
				case IN:
					if dt.End.IsZero() {
						dt.End = time
					} else if dt.Start.IsZero() {
						dt.Start = time
					}
				case AFTER, BEFORE:
					dt.Ref = time
				}
				timetokens = []string{}
			}
		}
		pos++ //advance to next token
	}
	(*tokens) = []string{}
ReturnDataTarget:
	return dt
}

/*
 * Tags targets can optionally start with 'distinct', or can just be '*'
 * or can contain a list of tag paths.
 */
func parseTagsTarget(tokens *[]string) Target_T {
	var tt = &TagsTarget{Distinct: false, Contents: []string{}}
	if len(*tokens) == 0 {
		return tt
	}
	pos := 0
	if (*tokens)[pos] == "distinct" {
		tt.Distinct = true
		pos++
	}
	for idx, val := range (*tokens)[pos:] {
		if val == "where" {
			/* if we hit this, then we have reached the end of the target
			 * definition. We alter "tokens" so that it starts with the "where"
			 * and return our Target_T
			**/
			(*tokens) = (*tokens)[idx+1:]
			return tt
		}
		// adds the token to the list of contents,
		// removing a trailing comma if there is one
		tmp := strings.TrimSuffix(val, ",")
		tmp = strings.Replace(tmp, "/", ".", -1)
		tt.Contents = append(tt.Contents, tmp)
	}
	(*tokens) = []string{}
	return tt
}

func parseSetTarget(tokens *[]string) Target_T {
	var st = &SetTarget{Updates: bson.M{}}
	if len(*tokens) == 0 {
		return st
	}
	pos := 0
	for {
		token := (*tokens)[pos]
		if token == "where" {
			(*tokens) = (*tokens)[pos+1:]
			return st
		}
		// key is token
		// check that (*tokens)[pos+1] is an equals sign
		if (*tokens)[pos+1] != "=" {
			log.Panic("Invalid syntax for setting tag")
		}
		value := (*tokens)[pos+2]
		token = cleantagstring(token)
		st.Updates[token] = value
		pos += 3
	}
	(*tokens) = []string{}
	return st
}

func parseWhere(tokens *[]string) *Node {
	var stack = [](Node){}
	pos := 0
	for {
		if pos == len(*tokens) {
			break
		}
		switch (*tokens)[pos] {
		case "and":
			left := stack[len(stack)-1]            // last item off stack
			stack = stack[:len(stack)-1]           // pop it off
			right, num := getNodeAt(pos+1, tokens) // next node
			node := Node{Type: AND_NODE, Left: left, Right: right}
			stack = append(stack, node)
			pos += 1 + num
			continue
		case "or":
			left := stack[len(stack)-1]            // last item off stack
			stack = stack[:len(stack)-1]           // pop it off
			right, num := getNodeAt(pos+1, tokens) // next node
			node := Node{Type: OR_NODE, Left: left, Right: right}
			stack = append(stack, node)
			pos += 1 + num
			continue
		default:
			node, num := getNodeAt(pos, tokens)
			stack = append(stack, node)
			pos += num
			continue
		}
		pos++
	}
	if len(stack) > 0 {
		return &stack[0]
	}
	return &Node{Type: DEF_NODE}
}

func getNodeAt(index int, tokens *[]string) (Node, int) {
	var node = Node{}
	var numtokens = 0
	if (*tokens)[index] == "has" {
		node.Left = (*tokens)[index+1]
		node.Type = getNodeType((*tokens)[index])
		node.Right = ""
		numtokens = 2
	} else {
		node.Left = (*tokens)[index]
		node.Type = getNodeType((*tokens)[index+1])
		node.Right = (*tokens)[index+2]
		numtokens = 3
	}
	node.Left = strings.Replace(node.Left.(string), "/", ".", -1)
	//node.Right = strings.Replace(node.Right.(string), "/", ".", -1)
	return node, numtokens
}

func makeAST(tokens []string) (*AST, error) {
	var ast = &AST{}

	/* QueryType */
	switch tokens[0] {
	case "select":
		ast.QueryType = SELECT_TYPE
	case "delete":
		ast.QueryType = DELETE_TYPE
	case "set":
		ast.QueryType = SET_TYPE
	default:
		return ast, errors.New("Query must be select or delete or set")
	}

	/* TargetType */
	// here, we postpone error checking to the parse____Target methods
	tmp_type := tokens[1]
	tokens = tokens[1:]
	switch tmp_type {
	case "data":
		ast.TargetType = DATA_TARGET
		ast.Target = parseDataTarget(&tokens)
	default:
		if ast.QueryType == SELECT_TYPE {
			ast.TargetType = TAGS_TARGET
			ast.Target = parseTagsTarget(&tokens)
		} else if ast.QueryType == SET_TYPE {
			ast.TargetType = SET_TARGET
			ast.Target = parseSetTarget(&tokens)
		}
	}

	/* Where */
	ast.Where = parseWhere(&tokens)

	return ast, nil
}

func parse(q string) *AST {
	tokens := tokenize(q)
	ast, _ := makeAST(tokens)
	return ast
}
