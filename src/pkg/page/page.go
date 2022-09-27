package page

import (
	"fmt"
	"io"

	"github.com/eknkc/amber"
)

var (
	templateDir = "./templates/email"
)

func Compile(wr io.Writer, s string, i interface{}) (err error) {
	temp, err := amber.CompileFile(fmt.Sprintf("%s/%s.amber", templateDir, s), amber.Options{PrettyPrint: true})
	if err != nil {
		return
	}
	return temp.Execute(wr, i)
}
