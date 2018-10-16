package walkers

import (
	"github.com/z7zmey/php-parser/parser"
	"github.com/z7zmey/php-parser/walker"
	"github.com/z7zmey/php-parser/node"
	"io"
	"reflect"
	"fmt"
	"strings"
)

type Signature struct {
	Namespace string
	MethodName string
}

type FindInfo struct {
	FileName string // имя файла
	Line string     // строка, в которой нашли совпадение
	LineNumber int  // номер строки
	Node node.Node  // нода совпадения (после парсинга)
	Signature Signature
}

type NodeFinder struct {
	Writer     io.Writer
	Indent     string
	Positions  parser.Positions
	FindSignature string
	FindInfo *FindInfo
}

// извлекает ноду MethodCall по сигнатуре
func (d NodeFinder) EnterNode(w walker.Walkable) bool {
	n := w.(node.Node)

	if reflect.TypeOf(n).String() == "*expr.MethodCall" {
		if d.Positions != nil {
			if p := d.Positions[n]; p != nil {
				if p.StartLine <= d.FindInfo.LineNumber && p.EndLine >= d.FindInfo.LineNumber {
					// надо проверить, что вызов действительно содержит сигнатуру
					// для этого нужно получить имя метода
					if d.reflectGet(n, "Method") == d.FindSignature {
						d.FindInfo.Node = n
					}
				}
			}
		}
	}

	return true
}

// получаем строковое значение поля из ноды по имени поля
// TODO: возможно, есть более простой способ
func (d NodeFinder) reflectGet(node node.Node, field string) string {
	e := reflect.ValueOf(node).Elem()

	for i := 0; i < e.NumField(); i++ {
		varName := e.Type().Field(i).Name
		if varName == field {
			tmp := fmt.Sprintf("%v", e.Field(i).Interface())
			return strings.Trim(tmp, "{}&")
		}
	}
	return ""
}

func (d NodeFinder) GetChildrenVisitor(key string) walker.Visitor {
	return NodeFinder{d.Writer, d.Indent + "    ", d.Positions, d.FindSignature, d.FindInfo}
}

func (d NodeFinder) LeaveNode(n walker.Walkable) {}