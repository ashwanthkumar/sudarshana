package main

import (
	"go/token"
)

func GetAllPositions(exprs []Expr) []token.Pos {
	positions := make([]token.Pos, 0)
	for _, expr := range exprs {
		// fmt.Printf("%v\n", expr)
		positions = append(positions, expr.AllPos()...)
	}
	return positions
}

// Meta represents the metadata for the SourceFile
type Meta struct {
	Source string `json:"source"`
}

// SourceFile represents the parsed AST for the given file
type SourceFile struct {
	Meta    Meta   `json:"meta"`
	Path    string `json:"path"`
	Package string `json:"package"`
	File    string `json:"file"`
	Exprs   []Expr `json:"lines"`
}

// Base type of all Expressions
type Expr interface {
	Scope() string
	Pos() token.Pos
	AllPos() []token.Pos
}

// Func represents a function call
type Func struct {
	Name string `json:"name"`
	// Reference has a non-empty value, if this instance is invoked from a package or a struct
	Reference string    `json:"reference,omitempty"`
	Args      []Expr    `json:"arguments"`
	Offset    token.Pos `json:"offset"`
	CScope    string    `json:"scope"`
	Type      string    `json:"type"`
}

func (v Func) Pos() token.Pos {
	return v.Offset
}

func (v Func) Scope() string {
	return v.CScope
}

func (v Func) AllPos() []token.Pos {
	positions := make([]token.Pos, 0)
	positions = append(positions, v.Pos())
	positions = append(positions, GetAllPositions(v.Args)...)
	return positions
}

// Variable represents a variable access in an expression
type Variable struct {
	Name string `json:"name"`
	// Reference has a non-empty value, if this instance is invoked from a package or a struct
	Reference string    `json:"reference,omitempty"`
	Offset    token.Pos `json:"offset"`
	CScope    string    `json:"scope"`
	Type      string    `json:"type"`
}

func (v Variable) Pos() token.Pos {
	return v.Offset
}

func (v Variable) Scope() string {
	return v.CScope
}

func (v Variable) AllPos() []token.Pos {
	positions := make([]token.Pos, 0)
	positions = append(positions, v.Pos())
	return positions
}

// Value represents a constanct of type string, int, double etc.
type Value struct {
	TypeOf string    `json:"typeOf"`
	Value  string    `json:"value"`
	Offset token.Pos `json:"offset"`
	CScope string    `json:"scope"`
	Type   string    `json:"type"`
}

func (v Value) Pos() token.Pos {
	return v.Offset
}

func (v Value) Scope() string {
	return v.CScope
}

func (v Value) AllPos() []token.Pos {
	positions := make([]token.Pos, 0)
	positions = append(positions, v.Pos())
	return positions
}

// Assignment represents an assignment expression
type Assignment struct {
	Lefts  []Expr    `json:"lhs"`
	Right  Expr      `json:"rhs"`
	Offset token.Pos `json:"offset"`
	CScope string    `json:"scope"`
	Type   string    `json:"type"`
}

func (v Assignment) Pos() token.Pos {
	return v.Offset
}

func (v Assignment) Scope() string {
	return v.CScope
}

func (v Assignment) AllPos() []token.Pos {
	positions := make([]token.Pos, 0)
	positions = append(positions, GetAllPositions(v.Lefts)...)
	if nil != v.Right {
		positions = append(positions, v.Right.AllPos()...)
	}
	return positions
}

// PropertyAccessInStruct represents accessing a property from a struct
type PropertyAccessInStruct struct {
	Struct   string    `json:"struct"`
	Property string    `json:"property"`
	Offset   token.Pos `json:"offset"`
	CScope   string    `json:"scope"`
	Type     string    `json:"type"`
}

func (v PropertyAccessInStruct) Pos() token.Pos {
	return v.Offset
}
func (v PropertyAccessInStruct) Scope() string {
	return v.CScope
}

type ConstructStruct struct {
	Struct       string            `json:"struct"`
	Args         []string          `json:"arguments,omitempty"`
	KeyValueArgs map[string]string `json:"kvargs,omitempty"`
	CScope       string            `json:"scope"`
	Type         string            `json:"type"`
	Offset       token.Pos         `json:"offset"`
}

func (v ConstructStruct) Pos() token.Pos {
	return v.Offset
}

func (v ConstructStruct) Scope() string {
	return v.CScope
}

func (v ConstructStruct) AllPos() []token.Pos {
	positions := make([]token.Pos, 0)
	positions = append(positions, v.Pos())
	return positions
}
