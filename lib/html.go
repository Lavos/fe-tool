package lib

import (
	"io"
	"os"
	"bytes"
	"path"
	"path/filepath"
	"html/template"
)

type HTMLBuildContext struct {
	sourceDirectory string
}

func NewHTMLBuildContext(sourceDirectory string) *HTMLBuildContext {
	return &HTMLBuildContext{
		sourceDirectory: sourceDirectory,
	}
}

func (b *HTMLBuildContext) InjectFragments(pattern string) (template.HTML, error) {
	use_path := path.Join(b.sourceDirectory, pattern)

	var html template.HTML
	filenames, err := filepath.Glob(use_path)

	if err != nil {
		return html, err
	}

	buf := new(bytes.Buffer)
	var file *os.File

	for _, fn := range filenames {
		file, err = os.Open(fn)

		if err != nil {
			return html, err
		}

		buf.ReadFrom(file)
		file.Close()
	}

	return template.HTML(buf.String()), nil
}

func (b *HTMLBuildContext) CompileFile(template_path string, data interface{}, w io.Writer) error {
	use_path := path.Join(b.sourceDirectory, template_path)
	_, filename := path.Split(template_path)

	t := template.New("base")
	t.Delims("<%", "%>")
	t.Funcs(template.FuncMap{
		"InjectFragments": b.InjectFragments,
	})

	_, err := t.ParseFiles(use_path)

	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, filename, data)
}
