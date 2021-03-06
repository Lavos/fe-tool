package lib

import (
	// "fmt"
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

func (b *HTMLBuildContext) ListModulesFromManifest(manifestLocation string) []string {
	use_path := path.Join(b.sourceDirectory, manifestLocation)

	manifest, err := ManifestFromFile(use_path)

	if err != nil {
		return []string{}
	}

	return manifest.Files
}

func (b *HTMLBuildContext) CompileFile(template_path string, data interface{}, w io.Writer) error {
	_, filename := path.Split(template_path)

	t := template.New("base")
	t.Delims("<%", "%>")
	t.Funcs(template.FuncMap{
		"InjectFragments": b.InjectFragments,
		"ListModulesFromManifest": b.ListModulesFromManifest,
	})

	_, err := t.ParseFiles(template_path)

	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, filename, data)
}
