package page

import (
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/Prep50mobileApp/prep50-api/src/pkg/logger"
	"github.com/eknkc/amber"
	"github.com/kataras/iris/v12"
)

type Template struct {
	v map[string]*template.Template
}

var templates *Template = &Template{}

func init() {
	var templateDir = "./templates/views"
	var err error
	templates.v, err = amber.CompileDir(templateDir, amber.DefaultDirOptions, amber.DefaultOptions)
	if err != nil {
		log.Fatal(err)
	}
}

type Writer interface {
	io.Writer
	StatusCode(int)
}

func Render(w Writer, name string, data iris.Map) {
	p, ok := templates.v[name]
	if !ok {
		return
	}

	if err := p.Execute(w, data); !logger.HandleError(err) {
		w.StatusCode(http.StatusNotFound)
		w.Write([]byte("Not found"))
		return
	}
}
