package lib

import (
	"io"
	"os"
	"path"
)

type MashContext struct {
	sourceDirectory string
}

func NewMashContext(sourceDirectory string) *MashContext {
	return &MashContext{
		sourceDirectory: sourceDirectory,
	}
}

func (c *MashContext) MashFilesToWriter(m *Manifest, w io.Writer) error {
	files := make([]*os.File, len(m.Files))
	readers := make([]io.Reader, len(m.Files))

	defer func(){
		for _, file := range files {
			file.Close()
		}
	}()

	for i, name := range m.Files {
		file, err := os.Open(path.Join(c.sourceDirectory, name))

		if err != nil {
			return err
		}

		files[i] = file
		readers[i] = file
	}

	mr := io.MultiReader(readers...)

	_, err := io.Copy(w, mr)

	return err
}
