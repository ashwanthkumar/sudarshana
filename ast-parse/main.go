package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	parseutil "gopkg.in/src-d/go-parse-utils.v1"
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

type SymbolTable struct {
	Symbols map[string]string
}

type ASTWalk struct {
	visited map[string]interface{}
	fileAst *ast.File
	fset    *token.FileSet
}

func NewASTWalk(fset *token.FileSet, fileAst *ast.File) *ASTWalk {
	return &ASTWalk{
		visited: make(map[string]interface{}),
		fileAst: fileAst,
		fset:    fset,
	}
}

func (a *ASTWalk) Walk(f func(ast.Node, *ASTWalk) bool) {
	ast.Inspect(a.fileAst, func(node ast.Node) bool {
		key := a.fset.Position(node.Pos()).String() + ":" + a.fset.Position(node.End()).String()
		_, present := a.visited[key]
		if present {
			return true
		} else {
			a.visited[key] = nil
			return f(node, a)
		}
	})
}

func (a *ASTWalk) Visit(node ast.Node) {
	key := a.fset.Position(node.Pos()).String() + ":" + a.fset.Position(node.End()).String()
	a.visited[key] = nil
}

func main() {
	sourceFile, err := os.Open("./src/src.go")
	src, err := ioutil.ReadAll(sourceFile)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	var importsMapping = make(map[string]string)
	// var symbolTables = make(map[string]*SymbolTable)
	// var activeFunc = "" // by default we're nowhere

	astWalk := NewASTWalk(fset, f)
	astWalk.Walk(func(node ast.Node, ctx *ASTWalk) bool {
		switch y := node.(type) {
		case *ast.ImportSpec:
			if strings.HasPrefix(y.Path.Value, "github.com") {
				importer := parseutil.NewImporter()
				pkgName, err := importer.Import(y.Path.Value)
				if err != nil {
					panic(err)
				}
				// fmt.Println(pkgName)
				importsMapping[y.Path.Value] = pkgName.Name()
			} else {
				// fmt.Println(y.Path.Value)
				importsMapping[y.Path.Value] = y.Path.Value
			}
		}
		return true
	})

}

//
// func parse(f *ast.File, importMapping map[string]string) {
// 	ast.Inspect(f, func(node ast.Node) bool {
// 		switch y := node.(type) {
// 		case *ast.ImportSpec:
// 			if strings.HasPrefix(y.Path.Value, "github.com") {
// 				importer := parseutil.NewImporter()
// 				pkgName, err := importer.Import(y.Path.Value)
// 				if err != nil {
// 					panic(err)
// 				}
// 				// fmt.Println(pkgName)
// 				importsMapping[y.Path.Value] = pkgName.Name()
// 			} else {
// 				// fmt.Println(y.Path.Value)
// 				importsMapping[y.Path.Value] = y.Path.Value
// 			}
//
// 		case *ast.FuncDecl:
// 			fmt.Printf("In func -- %s\n", y.Name.String())
// 			activeFunc = y.Name.String()
// 			symbolTables[y.Name.String()] = &SymbolTable{}
//
// 		case *ast.GenDecl:
// 			switch y.Tok {
// 			case token.VAR:
// 				// parse variables
// 				for _, spec := range y.Specs {
// 					valueSpec, ok := spec.(*ast.ValueSpec)
// 					if ok {
// 						fmt.Printf("Doing Import Spec -- %v\n", valueSpec)
// 					}
// 				}
// 			}
//
// 		case *ast.CallExpr:
// 			fmt.Printf("In func -- %s, and my symbolTable=%v\n", activeFunc, symbolTables[activeFunc])
// 			switch z := y.Fun.(type) {
// 			case *ast.SelectorExpr:
// 				x, ok := z.X.(*ast.Ident)
// 				if !ok {
// 					return false
// 				}
// 				pkg := x.Name
// 				sel := z.Sel
//
// 				name, lineNumber, offset := getNameLinePos(fset.Position(sel.NamePos))
// 				var kindStr string
// 				if x.Obj != nil {
// 					kindStr = sel.Obj.Kind.String()
// 				}
// 				s := Schema{
// 					Name:    sel.Name,
// 					Source:  name,
// 					Line:    lineNumber,
// 					Offset:  offset,
// 					Package: pkg,
// 					Kind:    kindStr,
// 				}
//
// 				j, _ := json.MarshalIndent(s, "", "\t")
// 				fmt.Println(string(j))
//
// 			}
// 		}
//
// 		return true
// 	})
//
// }

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
