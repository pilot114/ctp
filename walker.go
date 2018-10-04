package main

import (
	"github.com/z7zmey/php-parser/parser"
	"github.com/z7zmey/php-parser/walker"
	"github.com/z7zmey/php-parser/node"
	"fmt"
	"reflect"
	"io"
)

type Walker struct {
	Writer     io.Writer
	Indent     string
	Positions  parser.Positions
}

func (d Walker) EnterNode(w walker.Walkable) bool {
	n := w.(node.Node)

	// см. типы в z7zmey/php-parser/node/stmt
	if reflect.TypeOf(n).String() == "*stmt.ClassMethod" {

		val := reflect.Indirect(reflect.ValueOf(n))
		fmt.Println(val)

		//methodName := reflect.ValueOf(val.Field(2))

		fmt.Println(val.Field(2))

		//if d.Positions != nil {
		//	if p := d.Positions[n]; p != nil {
		//		fmt.Fprintln(d.Writer, "Position:", *p)
		//	}
		//}
	}

	return true
}

func (d Walker) GetChildrenVisitor(key string) walker.Visitor {
	//fmt.Fprintf(d.Writer, "%v%q:\n", d.Indent+"  ", key)
	return Walker{d.Writer, d.Indent + "    ", d.Positions}
}

func (d Walker) LeaveNode(n walker.Walkable) {
}

