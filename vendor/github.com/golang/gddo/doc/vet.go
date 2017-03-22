// Copyright 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file or at
// https://developers.google.com/open-source/licenses/bsd.

package doc

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/golang/gddo/gosrc"
)

// This list of deprecated exports is used to find code that has not been
// updated for Go 1.
var deprecatedExports = map[string][]string{
	`"bytes"`:         {"Add"},
	`"crypto/aes"`:    {"Cipher"},
	`"crypto/hmac"`:   {"NewSHA1", "NewSHA256"},
	`"crypto/rand"`:   {"Seed"},
	`"encoding/json"`: {"MarshalForHTML"},
	`"encoding/xml"`:  {"Marshaler", "NewParser", "Parser"},
	`"html"`:          {"NewTokenizer", "Parse"},
	`"image"`:         {"Color", "NRGBAColor", "RGBAColor"},
	`"io"`:            {"Copyn"},
	`"log"`:           {"Exitf"},
	`"math"`:          {"Fabs", "Fmax", "Fmod"},
	`"os"`:            {"Envs", "Error", "Getenverror", "NewError", "Time", "UnixSignal", "Wait"},
	`"reflect"`:       {"MapValue", "Typeof"},
	`"runtime"`:       {"UpdateMemStats"},
	`"strconv"`:       {"Atob", "Atof32", "Atof64", "AtofN", "Atoi64", "Atoui", "Atoui64", "Btoui64", "Ftoa64", "Itoa64", "Uitoa", "Uitoa64"},
	`"time"`:          {"LocalTime", "Nanoseconds", "NanosecondsToLocalTime", "Seconds", "SecondsToLocalTime", "SecondsToUTC"},
	`"unicode/utf8"`:  {"NewString"},
}

type vetVisitor struct {
	errors map[string]token.Pos
}

func (v *vetVisitor) Visit(n ast.Node) ast.Visitor {
	if sel, ok := n.(*ast.SelectorExpr); ok {
		if x, _ := sel.X.(*ast.Ident); x != nil {
			if obj := x.Obj; obj != nil && obj.Kind == ast.Pkg {
				if spec, _ := obj.Decl.(*ast.ImportSpec); spec != nil {
					for _, name := range deprecatedExports[spec.Path.Value] {
						if name == sel.Sel.Name {
							v.errors[fmt.Sprintf("%s.%s not found", spec.Path.Value, sel.Sel.Name)] = n.Pos()
							return nil
						}
					}
				}
			}
		}
	}
	return v
}

func (b *builder) vetPackage(pkg *Package, apkg *ast.Package) {
	errors := make(map[string]token.Pos)
	for _, file := range apkg.Files {
		for _, is := range file.Imports {
			importPath, _ := strconv.Unquote(is.Path.Value)
			if !gosrc.IsValidPath(importPath) &&
				!strings.HasPrefix(importPath, "exp/") &&
				!strings.HasPrefix(importPath, "appengine") {
				errors[fmt.Sprintf("Unrecognized import path %q", importPath)] = is.Pos()
			}
		}
		v := vetVisitor{errors: errors}
		ast.Walk(&v, file)
	}
	for message, pos := range errors {
		pkg.Errors = append(pkg.Errors,
			fmt.Sprintf("%s (%s)", message, b.fset.Position(pos)))
	}
}
