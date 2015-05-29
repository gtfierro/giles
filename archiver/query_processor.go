package archiver

import (
	"fmt"
	"github.com/gtfierro/giles/internal/tree"
	"strings"
)

type QueryProcessor struct {
	a      *Archiver
	graphs map[string]tree.Tree
}

func NewQueryProcessor(a *Archiver) (qp *QueryProcessor) {
	qp = &QueryProcessor{
		a:      a,
		graphs: make(map[string]tree.Tree),
	}
	return
}

func (qp *QueryProcessor) Parse(querystring string) *SQLex {
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
	fmt.Printf("operator list %v\n", l.query.operators)
	return l
}

func (qp *QueryProcessor) GetNodeFromOp(op *OpNode, inputType NodeType) tree.Node {
	var (
		operator OperationType
		found    bool
	)
	if operator, found = OpLookup[op.Operator]; !found {
		return nil
	}
	fmt.Printf("options %v\n", NodeLookup[operator])
	fmt.Printf("found? %v\n", NodeLookup[operator][inputType])
	//return
	return NodeLookup[operator][inputType](op.Arguments)
}
