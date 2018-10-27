package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"

	parseutil "gopkg.in/src-d/go-parse-utils.v1"
)

type Declaration struct {
	Label        string        `json:"label"`
	Type         string        `json:"type"`
	ReceiverType string        `json:"receiverType,omitempty"`
	Start        token.Pos     `json:"start"`
	End          token.Pos     `json:"end"`
	LineStart    int           `json:"lineStart"`
	LineEnd      int           `json:"lineEnd"`
	Children     []Declaration `json:"children,omitempty"`
}

func getReceiverType(fset *token.FileSet, decl *ast.FuncDecl) (string, error) {
	if decl.Recv == nil {
		return "", nil
	}

	buf := &bytes.Buffer{}
	if err := format.Node(buf, fset, decl.Recv.List[0].Type); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func parse(inputPackage string) {
	importer := parseutil.NewImporter()
	pkg, err := importer.Import(inputPackage)
	if err == nil {
		// fmt.Printf("pkg=%s\n", pkg.Path())
		// fmt.Printf("Name=%s\n", pkg.Name())
		dir, err := os.Open(pkg.Path())
		if err != nil {
			log.Fatalf("%q", err)
		}
		files, err := dir.Readdir(100)
		if err != nil {
			log.Fatalf("%q", err)
		}
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") && !strings.HasSuffix(file.Name(), "_test.go") {
				// path := pkg.Path() + "/" + file.Name()
				// fmt.Printf("visited file or dir: %q\n", path)
				parsefile(pkg.Name(), pkg.Path(), file.Name())
			}
		}
	}
}

func parsefile(packageName string, directory string, filename string) {
	inputFile := directory + "/" + filename
	sourceFile, err := os.Open(inputFile)
	src, err := ioutil.ReadAll(sourceFile)

	fset := token.NewFileSet()
	fileAst, err := parser.ParseFile(fset, inputFile, src, 0)
	if err != nil {
		panic(err)
	}

	visitor := NewASTVisitor(fset, inputFile)
	ast.Walk(visitor, fileAst)
	expressions := visitor.NewExprs

	meta := Meta{"github.com"}
	source := SourceFile{
		Meta:    meta,
		Path:    inputFile,
		Package: packageName,
		File:    filename,
	}

	source.Exprs = expressions

	// fmt.Printf("Length=%d\n", len(expressions))
	// for _, expr := range expressions {
	// fmt.Printf("%s --- %s -- %s --- %s --- %d\n", expr.Type, expr.VariableType, expr.Scope, expr.Name, fset.Position(expr.Pos).Offset)
	outAsJSON, err := json.Marshal(source)
	if err == nil {
		fmt.Printf("%s\n", string(outAsJSON))
	}
	// }
}

type ASTVisitor struct {
	InputFile      string
	NewExprs       []Expr
	fset           *token.FileSet
	visited        map[string]interface{}
	currentFunc    string
	fullPathToFile string
}

func NewASTVisitor(fset *token.FileSet, fullPathToFile string) *ASTVisitor {
	return &ASTVisitor{
		fset:           fset,
		visited:        make(map[string]interface{}),
		fullPathToFile: fullPathToFile,
	}
}

func (a *ASTVisitor) Key(pos token.Pos) string {
	return fmt.Sprintf("%d", pos)
}

func (a *ASTVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		key := a.Key(node.Pos())
		// fmt.Printf("I'm at offset=%d, key=%s\n", a.fset.Position(node.Pos()).Offset, key)
		_, seenAlready := a.visited[key]
		if !seenAlready {
			switch expr := node.(type) {
			case *ast.FuncDecl:
				reciver, err := getReceiverType(a.fset, expr)
				if err == nil {
					if reciver != "" {
						a.currentFunc = reciver + "#" + expr.Name.String()
					} else {
						a.currentFunc = expr.Name.String()
					}
				}
			case *ast.GenDecl:
				a.currentFunc = ""
			}

			exp := parseNode2(node, a.currentFunc, a.fullPathToFile)
			if nil != exp {
				a.NewExprs = append(a.NewExprs, exp)
				for _, pos := range exp.AllPos() {
					keyForExpr := a.Key(pos)
					a.visited[keyForExpr] = nil
				}
				// fmt.Printf("%v\n", exp)
			}
		}
	}
	return a
}

func toIdentText(in ast.Node) (string, bool) {
	identExpr, ok := in.(*ast.Ident)
	if ok {
		return identExpr.String(), true
	} else {
		return "", false
	}
}

func asSelectorExpr(in ast.Node) (*ast.SelectorExpr, bool) {
	selectorExpr, ok := in.(*ast.SelectorExpr)
	if ok {
		return selectorExpr, true
	} else {
		return nil, false
	}
}

func asKeyValueExpr(in ast.Node) (*ast.KeyValueExpr, bool) {
	kvExpr, ok := in.(*ast.KeyValueExpr)
	if ok {
		return kvExpr, true
	} else {
		return nil, false
	}
}
func asBasicLit(in ast.Node) (*ast.BasicLit, bool) {
	basicExpr, ok := in.(*ast.BasicLit)
	if ok {
		return basicExpr, true
	} else {
		return nil, false
	}
}

var importMappingCache map[string]string

func init() {
	importMappingCache = make(map[string]string)
}

func parseNode2(node ast.Node, scope string, fullPathToFile string) Expr {
	// expressions := []Expr{}
	// fmt.Printf("%s -- %v\n", reflect.TypeOf(node), node)
	switch expr := node.(type) {
	case *ast.AssignStmt:
		leftExprs := make([]Expr, 0)
		for _, lhs := range expr.Lhs {
			// var lExpression Expr
			switch l := lhs.(type) {
			case *ast.Ident:
				// ignore the _ names, we don't care mostly
				if l.String() != "_" {
					lExpression := Variable{
						Name:   l.String(),
						Type:   "variable",
						Offset: l.Pos(),
						CScope: scope,
					}

					switch r := expr.Rhs[0].(type) {
					case *ast.CompositeLit:
						switch rType := r.Type.(type) {
						case *ast.Ident:
							lExpression.Reference = rType.String()
						}
					}

					leftExprs = append(leftExprs, lExpression)
				}
			}
		}

		var rhs Expr
		switch r := expr.Rhs[0].(type) {
		case *ast.CompositeLit:
			switch rType := r.Type.(type) {
			case *ast.Ident:
				typeOfVariable := rType
				createStruct := ConstructStruct{
					Type:   "constructstruct",
					Offset: expr.Pos(),
					CScope: scope,
					Struct: typeOfVariable.String(),
				}
				createStruct.KeyValueArgs = make(map[string]string)

				for _, elt := range r.Elts {
					eltAsKV, ok := asKeyValueExpr(elt)
					if ok {
						key, _ := toIdentText(eltAsKV.Key)
						value, _ := toIdentText(eltAsKV.Value)
						createStruct.KeyValueArgs[key] = value
					} else {
						eltAsBasic, ok := asBasicLit(elt)
						if ok {
							createStruct.Args = append(createStruct.Args, eltAsBasic.Value)
						}
					}
				}
				rhs = createStruct
			}
		}

		// attempt to parse the values
		if nil == rhs {
			rhs = parseNode2(expr.Rhs[0], scope, fullPathToFile)
		}
		assignment := Assignment{
			Lefts:  leftExprs,
			CScope: scope,
			Type:   "assignment",
			Offset: expr.Pos(),
		}
		if nil != rhs {
			assignment.Right = rhs
		}

		return assignment

	case *ast.CallExpr:
		f := Func{
			Type:   "function",
			CScope: scope,
			Offset: expr.Pos(),
		}
		funSelector, ok := asSelectorExpr(expr.Fun)
		if ok {
			xAsIdent, ok := funSelector.X.(*ast.Ident)
			if ok {
				f.Reference = xAsIdent.String()
				cacheValue, alreadyInCache := importMappingCache[xAsIdent.String()]
				if alreadyInCache {
					f.Reference = cacheValue
				} else {
					// TODO: Invoke Guru and see if we can get the canonical name of this reference
					query := fmt.Sprintf("%s:#%d", fullPathToFile, funSelector.Pos())
					packageReference := guru_describe(query)
					// fmt.Printf("query=%s\noutput=%q\n", query, packageReference)
					if nil != packageReference && nil != packageReference.Package {
						if "" != packageReference.Package.Path {
							resolvedPath := resolveGuruPath(packageReference.Package.Path)
							importMappingCache[xAsIdent.String()] = resolvedPath
							f.Reference = resolvedPath
						}
					}
				}
			}
			f.Name = funSelector.Sel.String()
		} else {
			funSelector, _ := expr.Fun.(*ast.Ident)
			f.Name = funSelector.String()
		}
		f.Args = make([]Expr, 0)

		for _, arg := range expr.Args {
			switch argExpr := arg.(type) {
			case *ast.Ident:
				v := Variable{
					Type:   "variable",
					CScope: scope,
					Name:   argExpr.String(),
					Offset: argExpr.Pos(),
				}
				f.Args = append(f.Args, v)
			default:
				otherExpr := parseNode2(arg, scope, fullPathToFile)
				if nil != otherExpr {
					f.Args = append(f.Args, otherExpr)
				}
			}
		}
		return f
	case *ast.BasicLit:
		if "" != scope {
			v := Value{
				Type:   "constant",
				CScope: scope,
				TypeOf: expr.Kind.String(),
				Value:  expr.Value,
				Offset: expr.Pos(),
			}
			return v
		}
	}

	return nil
}

func resolveGuruPath(guruPath string) string {
	if strings.Contains(guruPath, "/vendor/") {
		// we've a vendored path, filter things before /vendor/ to get the import path
		return strings.Split(guruPath, "/vendor/")[1]
	}

	return guruPath
}

func outline(fset *token.FileSet, fileAst *ast.File) []Declaration {
	declarations := []Declaration{}

	for _, decl := range fileAst.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			receiverType, err := getReceiverType(fset, decl)
			if err != nil {
				// reportError(fmt.Errorf("Failed to parse receiver type: %v", err))
			}
			declarations = append(declarations, Declaration{
				decl.Name.String(),
				"function",
				receiverType,
				decl.Pos(),
				decl.End(),
				fset.Position(decl.Pos()).Line,
				fset.Position(decl.End()).Line,
				[]Declaration{},
			})
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				// case *ast.ImportSpec:
				// 	declarations = append(declarations, Declaration{
				// 		spec.Path.Value,
				// 		"import",
				// 		"",
				// 		spec.Pos(),
				// 		spec.End(),
				// 		[]Declaration{},
				// 	})
				case *ast.TypeSpec:
					//TODO: Members if it's a struct or interface type?
					declarations = append(declarations, Declaration{
						spec.Name.String(),
						"type",
						"",
						spec.Pos(),
						spec.End(),
						fset.Position(spec.Pos()).Line,
						fset.Position(spec.End()).Line,
						[]Declaration{},
					})
				case *ast.ValueSpec:
					for _, id := range spec.Names {
						declarations = append(declarations, Declaration{
							id.Name,
							"variable",
							"",
							id.Pos(),
							id.End(),
							fset.Position(id.Pos()).Line,
							fset.Position(id.End()).Line,
							[]Declaration{},
						})
					}
				default:
					// reportError(fmt.Errorf("Unknown token type: %s", decl.Tok))
				}
			}
		default:
			// reportError(fmt.Errorf("Unknown declaration @", decl.Pos()))
		}
	}

	return declarations
}
