package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"gopkg.in/src-d/go-parse-utils.v1"
)

type Schema struct {
	Source  string `json:"source"`
	Line    int64  `json:"line"`
	Offset  int64  `json:"offset"`
	Kind    string `json:"kind"`
	Package string `json:"package"`
	Name    string `json:"name"`
	// Meta    Meta   `json:"meta"`
}

type Meta struct {
	Source string `json:"source"`
	Repo   string `json:"repo"`
	Stars  int64  `json:"stars"`
	Forks  int64  `json:"forks"`
}

func main() {
	sourceFile, err := os.Open("./src/src.go")
	src, err := ioutil.ReadAll(sourceFile)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	ast.Inspect(f, func(node ast.Node) bool {
		switch y := node.(type) {
		case *ast.ImportSpec:
			fmt.Println(y.Path.Value)
			if strings.HasPrefix(y.Path.Value, "github.com") {
				importer := parseutil.NewImporter()
				pkgName, err := importer.Import(y.Path.Value)
				if err != nil {
					panic(err)
				}
				fmt.Println(pkgName)
			} else {
				fmt.Println(y.Path.Value)
			}

		case *ast.CallExpr:
			switch z := y.Fun.(type) {
			case *ast.SelectorExpr:
				x, ok := z.X.(*ast.Ident)
				if !ok {
					return false
				}
				pkg := x.Name
				sel := z.Sel

				name, lineNumber, offset := getNameLinePos(fset.Position(sel.NamePos))
				var kindStr string
				if x.Obj != nil {
					kindStr = sel.Obj.Kind.String()
				}
				s := Schema{
					Name:    sel.Name,
					Source:  name,
					Line:    lineNumber,
					Offset:  offset,
					Package: pkg,
					Kind:    kindStr,
				}

				j, _ := json.MarshalIndent(s, "", "\t")
				fmt.Println(string(j))

			}
		}

		return true
	})
}

func getNameLinePos(pos token.Position) (string, int64, int64) {

	NameWithPosSlice := strings.Split(pos.String(), ":")
	name := NameWithPosSlice[0]
	lineNumber, err := strconv.ParseInt(NameWithPosSlice[1], 10, 64)
	if err != nil {
		fmt.Errorf(err.Error())
	}
	offset, err := strconv.ParseInt(NameWithPosSlice[2], 10, 64)
	if err != nil {
		fmt.Errorf(err.Error())
	}

	return name, lineNumber, offset
}
