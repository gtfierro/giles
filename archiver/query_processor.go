package archiver

import (
	"fmt"
	"strings"
)

type QueryProcessor struct {
	a      *Archiver
	graphs map[string]*Node
	done   chan struct{}
}

func NewQueryProcessor(a *Archiver) (qp *QueryProcessor) {
	qp = &QueryProcessor{
		a:      a,
		graphs: make(map[string]*Node),
		done:   make(chan struct{}),
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
	fmt.Printf("select keys %v\n", l.query.Contents)
	return l
}

func (qp *QueryProcessor) GetNodeFromOp(op *OpNode, query *query) *Node {
	var (
		operator OperationType
		found    bool
		node     *Node
	)
	if operator, found = OpLookup[op.Operator]; !found {
		return nil
	}

	// Populate extra information in nodes that need it
	switch operator {
	case WINDOW:
		node = NodeLookup[operator](qp.done, op.Arguments, query.data)
	default:
		node = NodeLookup[operator](qp.done, op.Arguments)
	}

	return node
}

// Checks that the ouput of node @out is compatible with the input of node @in.
// First checks that the structures match. If structures match, then it checks
// the data type. If the datatypes do not match, then we return false
// This does not actually resolve the types of the nodes, but rather just
// checks that they are potentially compatible. The actual type resolution
// is performed when the nodes are evaluated
func (qp *QueryProcessor) CheckOutToIn(out, in *Node) bool {
	// check structures exist
	outStructure, outFound := out.Tags["out:structure"]
	inStructure, inFound := in.Tags["in:structure"]
	if !inFound || !outFound {
		log.Error("Both out and in must supply a structure")
		return false
	}

	// check structure matches
	if (outStructure.(StructureType) & inStructure.(StructureType)) == 0 {
		log.Error("Out structure does not match in: %v %v\n", outStructure, inStructure)
		return false
	}

	// check datatypes exist
	outDatatype, outFound := out.Tags["out:datatype"]
	inDatatype, inFound := in.Tags["in:datatype"]
	if !outFound || !inFound {
		log.Error("Both out and in must supply a data type")
		return false
	}

	// check datatype matches
	if (outDatatype.(DataType) & inDatatype.(DataType)) == 0 {
		log.Error("Out datatype does not match in: %v %v\n", outDatatype, inDatatype)
		return false
	}

	return true
}

func nodeHasOutput(n *Node, structure StructureType, datatype DataType) bool {
	return n.HasOutput(uint(structure), uint(datatype))
}

func nodeHasInput(n *Node, structure StructureType, datatype DataType) bool {
	return n.HasInput(uint(structure), uint(datatype))
}
