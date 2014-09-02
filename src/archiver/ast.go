package main

import (
	"errors"
	"fmt"
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

func parseDataTarget(tokens *[]string) Target_T {
	return nil
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

func parseWhere(tokens *[]string) *Node {
	var stack = [](Node){}
	pos := 0
	for {
		if pos == len(*tokens) {
			break
		}
		switch (*tokens)[pos] {
		case "has":
			fmt.Println("has")
			pos += 2
			continue
		case "and":
			left := stack[len(stack)-1]       // last item off stack
			stack = stack[:len(stack)-1]      // pop it off
			right := getNodeAt(pos+1, tokens) // next node
			node := Node{Type: AND_NODE, Left: left, Right: right}
			stack = append(stack, node)
			pos += 4
			continue
		case "or":
			left := stack[len(stack)-1]       // last item off stack
			stack = stack[:len(stack)-1]      // pop it off
			right := getNodeAt(pos+1, tokens) // next node
			node := Node{Type: OR_NODE, Left: left, Right: right}
			stack = append(stack, node)
			pos += 4
			continue
		default:
			node := getNodeAt(pos, tokens)
			stack = append(stack, node)
			pos += 3
			continue
		}
		pos++
	}
	for _, n := range stack {
		fmt.Println("clause:", n.ToBson())
	}
	if len(stack) > 0 {
		return &stack[0]
	}
	return &Node{Type: DEF_NODE}
}

func getNodeAt(index int, tokens *[]string) Node {
	var node = Node{}
	node.Left = (*tokens)[index]
	node.Type = getNodeType((*tokens)[index+1])
	node.Right = (*tokens)[index+2]

	node.Left = strings.Replace(node.Left.(string), "/", ".", -1)
	node.Right = strings.Replace(node.Right.(string), "/", ".", -1)
	return node
}

func makeAST(tokens []string) (*AST, error) {
	var ast = &AST{}

	/* QueryType */
	switch tokens[0] {
	case "select":
		ast.QueryType = SELECT_TYPE
	case "delete":
		ast.QueryType = DELETE_TYPE
	default:
		return ast, errors.New("Query must be select or delete")
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
		ast.TargetType = TAGS_TARGET
		ast.Target = parseTagsTarget(&tokens)
	}

	/* Where */
	ast.Where = parseWhere(&tokens)

	return ast, nil
}

func parse(q string) *AST {
	fmt.Println(q)
	tokens := tokenize(q)
	ast, _ := makeAST(tokens)
	ast.Repr()
	return ast
}

//func main() {
//	var q string
//	q = "select *"
//	parse(q)
//
//	q = "select * where Metadata/System = 'Lighting'"
//	parse(q)
//
//	q = "select * where Metadata/System='Building Lighting'"
//	parse(q)
//
//	q = "select Metadata/System, uuid where Metadata/site='012345' and Metadata/System = 'HVAC'"
//	parse(q)
//
//	q = "select uuid where Metadata/site != '123'"
//	parse(q)
//
//	q = "select uuid where Metadata/site!='123'"
//	parse(q)
//}
