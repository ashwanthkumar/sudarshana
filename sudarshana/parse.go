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
	"os"
)

type Expression struct {
	Name         string
	Pos          token.Pos
	Type         string
	VariableType string
	Scope        string
}

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

func parse(inputFile string) {
	sourceFile, err := os.Open(inputFile)
	src, err := ioutil.ReadAll(sourceFile)

	fset := token.NewFileSet()
	fileAst, err := parser.ParseFile(fset, inputFile, src, 0)
	if err != nil {
		panic(err)
	}

	visitor := NewASTVisitor(fset)
	ast.Walk(visitor, fileAst)
	expressions := visitor.NewExprs

	// fmt.Printf("Length=%d\n", len(expressions))
	for _, expr := range expressions {
		// fmt.Printf("%s --- %s -- %s --- %s --- %d\n", expr.Type, expr.VariableType, expr.Scope, expr.Name, fset.Position(expr.Pos).Offset)
		outAsJson, err := json.Marshal(expr)
		if err == nil {
			fmt.Printf("%s\n", string(outAsJson))
		}
	}
}

type ASTVisitor struct {
	Expressions []Expression
	NewExprs    []Expr
	fset        *token.FileSet
	visited     map[string]interface{}
	currentFunc string
}

func NewASTVisitor(fset *token.FileSet) *ASTVisitor {
	return &ASTVisitor{
		Expressions: []Expression{},
		fset:        fset,
		visited:     make(map[string]interface{}),
	}
}

func (a *ASTVisitor) Key(node ast.Node) string {
	return fmt.Sprintf("%d", node.Pos())
}

func (a *ASTVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		key := a.Key(node)
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

			expressions := parseNode(node, a.currentFunc)
			for _, expr := range expressions {
				// fmt.Printf("Found %s --- %d\n", expr.Name, a.fset.Position(expr.Pos).Offset)
				keyForExpr := fmt.Sprintf("%d", expr.Pos)
				a.visited[keyForExpr] = nil
			}

			a.Expressions = append(a.Expressions, expressions...)

			exp := parseNode2(node, a.currentFunc)
			if nil != exp {
				a.NewExprs = append(a.NewExprs, exp)
				for _, pos := range exp.AllPos() {
					keyForExpr := fmt.Sprintf("%d", pos)
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

func parseNode2(node ast.Node, scope string) Expr {
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
						typeOfVariable := r.Type.(*ast.Ident)
						lExpression.Reference = typeOfVariable.String()
					}

					leftExprs = append(leftExprs, lExpression)
				}
			}
		}

		var rhs Expr
		switch r := expr.Rhs[0].(type) {
		case *ast.CompositeLit:
			typeOfVariable := r.Type.(*ast.Ident)
			createStruct := ConstructStruct{
				Type:   "constructstruct",
				Offset: expr.Pos(),
				CScope: scope,
				Struct: typeOfVariable.String(),
			}

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

		// attempt to parse the values
		if nil == rhs {
			rhs = parseNode2(expr.Rhs[0], scope)
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
				otherExpr := parseNode2(arg, scope)
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

func parseNode(node ast.Node, scope string) []Expression {
	// fmt.Printf("%s -- %v\n", reflect.TypeOf(node), node)
	expressions := []Expression{}
	switch expr := node.(type) {
	case *ast.CallExpr:
		funcAstExpressions := parseNode(expr.Fun, scope)
		updatedFuncAsts := make([]Expression, len(funcAstExpressions))
		for idx, f := range funcAstExpressions {
			f.Type = "function"
			updatedFuncAsts[idx] = f
		}
		for _, argExpr := range expr.Args {
			// fmt.Printf("%v\n", argExpr)
			e := parseNode(argExpr, scope)
			expressions = append(expressions, e...)
		}
		expressions = append(expressions, updatedFuncAsts...)
	case *ast.SelectorExpr:
		xAsIdent, ok := expr.X.(*ast.Ident)
		if ok {
			e := Expression{
				Name:  xAsIdent.String() + "#" + expr.Sel.String(),
				Pos:   xAsIdent.Pos(),
				Type:  "reference",
				Scope: scope,
			}
			expressions = append(expressions, e)
		}
	case *ast.AssignStmt:
		for _, lhs := range expr.Lhs {
			var lExpression Expression
			switch l := lhs.(type) {
			case *ast.Ident:
				// ignore the _ names, we don't care mostly
				if l.String() != "_" {
					lExpression = Expression{
						Name:  l.String(),
						Type:  "variable",
						Pos:   l.Pos(),
						Scope: scope,
					}

					switch r := expr.Rhs[0].(type) {
					case *ast.CompositeLit:
						typeOfVariable := r.Type.(*ast.Ident)
						lExpression.VariableType = typeOfVariable.String()
					}
					expressions = append(expressions, lExpression)
				}
			}
		}
		// for _, rhs := range expr.Rhs {
		// 	e := parseNode(rhs)
		// 	expressions = append(expressions, e...)
		// }
	default:
		// do nothing
	}

	return expressions
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
