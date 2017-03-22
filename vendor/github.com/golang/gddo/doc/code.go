// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package doc

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/printer"
	"go/scanner"
	"go/token"
	"math"
	"strconv"
)

const (
	notPredeclared = iota
	predeclaredType
	predeclaredConstant
	predeclaredFunction
)

// predeclared represents the set of all predeclared identifiers.
var predeclared = map[string]int{
	"bool":       predeclaredType,
	"byte":       predeclaredType,
	"complex128": predeclaredType,
	"complex64":  predeclaredType,
	"error":      predeclaredType,
	"float32":    predeclaredType,
	"float64":    predeclaredType,
	"int16":      predeclaredType,
	"int32":      predeclaredType,
	"int64":      predeclaredType,
	"int8":       predeclaredType,
	"int":        predeclaredType,
	"rune":       predeclaredType,
	"string":     predeclaredType,
	"uint16":     predeclaredType,
	"uint32":     predeclaredType,
	"uint64":     predeclaredType,
	"uint8":      predeclaredType,
	"uint":       predeclaredType,
	"uintptr":    predeclaredType,

	"true":  predeclaredConstant,
	"false": predeclaredConstant,
	"iota":  predeclaredConstant,
	"nil":   predeclaredConstant,

	"append":  predeclaredFunction,
	"cap":     predeclaredFunction,
	"close":   predeclaredFunction,
	"complex": predeclaredFunction,
	"copy":    predeclaredFunction,
	"delete":  predeclaredFunction,
	"imag":    predeclaredFunction,
	"len":     predeclaredFunction,
	"make":    predeclaredFunction,
	"new":     predeclaredFunction,
	"panic":   predeclaredFunction,
	"print":   predeclaredFunction,
	"println": predeclaredFunction,
	"real":    predeclaredFunction,
	"recover": predeclaredFunction,
}

type AnnotationKind int16

const (
	// Link to export in package specified by Paths[PathIndex] with fragment
	// Text[strings.LastIndex(Text[Pos:End], ".")+1:End].
	LinkAnnotation AnnotationKind = iota

	// Anchor with name specified by Text[Pos:End] or typeName + "." +
	// Text[Pos:End] for type declarations.
	AnchorAnnotation

	// Comment.
	CommentAnnotation

	// Link to package specified by Paths[PathIndex].
	PackageLinkAnnotation

	// Link to builtin entity with name Text[Pos:End].
	BuiltinAnnotation
)

type Annotation struct {
	Pos, End  int32
	Kind      AnnotationKind
	PathIndex int16
}

type Code struct {
	Text        string
	Annotations []Annotation
	Paths       []string
}

// declVisitor modifies a declaration AST for printing and collects annotations.
type declVisitor struct {
	annotations []Annotation
	paths       []string
	pathIndex   map[string]int
	comments    []*ast.CommentGroup
}

func (v *declVisitor) add(kind AnnotationKind, importPath string) {
	pathIndex := -1
	if importPath != "" {
		var ok bool
		pathIndex, ok = v.pathIndex[importPath]
		if !ok {
			pathIndex = len(v.paths)
			v.paths = append(v.paths, importPath)
			v.pathIndex[importPath] = pathIndex
		}
	}
	v.annotations = append(v.annotations, Annotation{Kind: kind, PathIndex: int16(pathIndex)})
}

func (v *declVisitor) ignoreName() {
	v.add(-1, "")
}

func (v *declVisitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.TypeSpec:
		v.ignoreName()
		switch n := n.Type.(type) {
		case *ast.InterfaceType:
			for _, f := range n.Methods.List {
				for _ = range f.Names {
					v.add(AnchorAnnotation, "")
				}
				ast.Walk(v, f.Type)
			}
		case *ast.StructType:
			for _, f := range n.Fields.List {
				for _ = range f.Names {
					v.add(AnchorAnnotation, "")
				}
				ast.Walk(v, f.Type)
			}
		default:
			ast.Walk(v, n)
		}
	case *ast.FuncDecl:
		if n.Recv != nil {
			ast.Walk(v, n.Recv)
		}
		v.ignoreName()
		ast.Walk(v, n.Type)
	case *ast.Field:
		for _ = range n.Names {
			v.ignoreName()
		}
		ast.Walk(v, n.Type)
	case *ast.ValueSpec:
		for _ = range n.Names {
			v.add(AnchorAnnotation, "")
		}
		if n.Type != nil {
			ast.Walk(v, n.Type)
		}
		for _, x := range n.Values {
			ast.Walk(v, x)
		}
	case *ast.Ident:
		switch {
		case n.Obj == nil && predeclared[n.Name] != notPredeclared:
			v.add(BuiltinAnnotation, "")
		case n.Obj != nil && ast.IsExported(n.Name):
			v.add(LinkAnnotation, "")
		default:
			v.ignoreName()
		}
	case *ast.SelectorExpr:
		if x, _ := n.X.(*ast.Ident); x != nil {
			if obj := x.Obj; obj != nil && obj.Kind == ast.Pkg {
				if spec, _ := obj.Decl.(*ast.ImportSpec); spec != nil {
					if path, err := strconv.Unquote(spec.Path.Value); err == nil {
						v.add(PackageLinkAnnotation, path)
						if path == "C" {
							v.ignoreName()
						} else {
							v.add(LinkAnnotation, path)
						}
						return nil
					}
				}
			}
		}
		ast.Walk(v, n.X)
		v.ignoreName()
	case *ast.BasicLit:
		if n.Kind == token.STRING && len(n.Value) > 128 {
			v.comments = append(v.comments,
				&ast.CommentGroup{List: []*ast.Comment{{
					Slash: n.Pos(),
					Text:  fmt.Sprintf("/* %d byte string literal not displayed */", len(n.Value)),
				}}})
			n.Value = `""`
		} else {
			return v
		}
	case *ast.CompositeLit:
		if len(n.Elts) > 100 {
			if n.Type != nil {
				ast.Walk(v, n.Type)
			}
			v.comments = append(v.comments,
				&ast.CommentGroup{List: []*ast.Comment{{
					Slash: n.Lbrace,
					Text:  fmt.Sprintf("/* %d elements not displayed */", len(n.Elts)),
				}}})
			n.Elts = n.Elts[:0]
		} else {
			return v
		}
	default:
		return v
	}
	return nil
}

func (b *builder) printDecl(decl ast.Decl) (d Code) {
	v := &declVisitor{pathIndex: make(map[string]int)}
	ast.Walk(v, decl)
	b.buf = b.buf[:0]
	err := (&printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}).Fprint(
		sliceWriter{&b.buf},
		b.fset,
		&printer.CommentedNode{Node: decl, Comments: v.comments})
	if err != nil {
		return Code{Text: err.Error()}
	}

	var annotations []Annotation
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(b.buf))
	s.Init(file, b.buf, nil, scanner.ScanComments)
	prevTok := token.ILLEGAL
loop:
	for {
		pos, tok, lit := s.Scan()
		switch tok {
		case token.EOF:
			break loop
		case token.COMMENT:
			p := file.Offset(pos)
			e := p + len(lit)
			if prevTok == token.COMMENT {
				annotations[len(annotations)-1].End = int32(e)
			} else {
				annotations = append(annotations, Annotation{Kind: CommentAnnotation, Pos: int32(p), End: int32(e)})
			}
		case token.IDENT:
			if len(v.annotations) == 0 {
				// Oops!
				break loop
			}
			annotation := v.annotations[0]
			v.annotations = v.annotations[1:]
			if annotation.Kind == -1 {
				continue
			}
			p := file.Offset(pos)
			e := p + len(lit)
			annotation.Pos = int32(p)
			annotation.End = int32(e)
			annotations = append(annotations, annotation)
		}
		prevTok = tok
	}
	return Code{Text: string(b.buf), Annotations: annotations, Paths: v.paths}
}

func (b *builder) position(n ast.Node) Pos {
	var position Pos
	pos := b.fset.Position(n.Pos())
	src := b.srcs[pos.Filename]
	if src != nil {
		position.File = int16(src.index)
		position.Line = int32(pos.Line)
		end := b.fset.Position(n.End())
		if src == b.srcs[end.Filename] {
			n := end.Line - pos.Line
			if n >= 0 && n <= math.MaxUint16 {
				position.N = uint16(n)
			}
		}
	}
	return position
}

func (b *builder) printExample(e *doc.Example) (code Code, output string) {
	output = e.Output

	b.buf = b.buf[:0]
	var n interface{}
	if _, ok := e.Code.(*ast.File); ok {
		n = e.Play
	} else {
		n = &printer.CommentedNode{Node: e.Code, Comments: e.Comments}
	}
	err := (&printer.Config{Mode: printer.UseSpaces, Tabwidth: 4}).Fprint(sliceWriter{&b.buf}, b.fset, n)
	if err != nil {
		return Code{Text: err.Error()}, output
	}

	// additional formatting if this is a function body
	if i := len(b.buf); i >= 2 && b.buf[0] == '{' && b.buf[i-1] == '}' {
		// remove surrounding braces
		b.buf = b.buf[1 : i-1]
		// unindent
		b.buf = bytes.Replace(b.buf, []byte("\n    "), []byte("\n"), -1)
		// remove output comment
		if j := exampleOutputRx.FindIndex(b.buf); j != nil {
			b.buf = bytes.TrimSpace(b.buf[:j[0]])
		}
	} else {
		// drop output, as the output comment will appear in the code
		output = ""
	}

	var annotations []Annotation
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(b.buf))
	s.Init(file, b.buf, nil, scanner.ScanComments)
	prevTok := token.ILLEGAL
scanLoop:
	for {
		pos, tok, lit := s.Scan()
		switch tok {
		case token.EOF:
			break scanLoop
		case token.COMMENT:
			p := file.Offset(pos)
			e := p + len(lit)
			if prevTok == token.COMMENT {
				annotations[len(annotations)-1].End = int32(e)
			} else {
				annotations = append(annotations, Annotation{Kind: CommentAnnotation, Pos: int32(p), End: int32(e)})
			}
		}
		prevTok = tok
	}

	return Code{Text: string(b.buf), Annotations: annotations}, output
}
