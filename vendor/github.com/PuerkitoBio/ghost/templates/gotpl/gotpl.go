package gotpl

import (
	"html/template"

	"github.com/PuerkitoBio/ghost/templates"
)

// The template compiler for native Go templates.
type GoTemplateCompiler struct{}

// Implementation of the TemplateCompiler interface.
func (this *GoTemplateCompiler) Compile(f string) (templates.Templater, error) {
	return template.ParseFiles(f)
}

func init() {
	templates.Register(".tmpl", new(GoTemplateCompiler))
}
