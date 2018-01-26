package cmd

import (
	"io"
	"os"
	"bytes"
	"fmt"
	"net/http"
	"mime"
	"time"
	"path"
	"strings"

	"github.com/spf13/cobra"
	libsass "github.com/wellington/go-libsass"
)

func compileManifest(m *Manifest, w io.Writer) error {
	files := make([]*os.File, len(m.Files))
	readers := make([]io.Reader, len(m.Files))

	defer func(){
		for _, file := range files {
			file.Close()
		}
	}()

	for i, name := range m.Files {
		file, err := os.Open(src + "/" + name)

		if err != nil {
			return err
		}

		files[i] = file
		readers[i] = file
	}

	mr := io.MultiReader(readers...)

	comp, err := libsass.New(w, mr)

	if err != nil {
		return err
	}

	return comp.Run()
}

func sassHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/css")

	upath := req.URL.Path

	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}

	upath = path.Clean(root + upath + ".manifest")

	// attempt to open manifest file
	manifest_file, err := os.Open(upath)

	if err != nil {
		l.Printf("No manifest file found for: %s", upath)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, manifest_file)

	m, err := ManifestFromBytes(buf.Bytes())

	if err != nil {
		l.Print("Could not get manifest: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = compileManifest(m, w)

	if err != nil {
		l.Printf("Mash Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

var (
	sassCommand = &cobra.Command{
		Use: "sass",
		Short: "Process SASS files to output mashed CSS",
	}

	sassServerCommand = &cobra.Command{
		Use:   "server",
		Short: "Host HTTP server returning processed and mashed files on request",
		RunE: func(cmd *cobra.Command, args []string) error {
			mime.AddExtensionType(".css", "text/css")

			l.Printf("Listening at %d, hosting manifest files from: %s", port, root)

			s := &http.Server{
				Addr:           fmt.Sprintf(":%d", port),
				Handler:        http.HandlerFunc(sassHandler),
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				MaxHeaderBytes: 1 << 20,
			}

			go func(){
				l.Fatal(s.ListenAndServe())
			}()

			WaitForSignal()
			return nil
		},
	}

	sassOutputCommand = &cobra.Command{
		Use:   "output",
		Short: "Parses STDIN as Manifest, returning proceessed and mashed files to STDOUT",
		RunE: func(cmd *cobra.Command, args []string) error {
			buf := new(bytes.Buffer)
			io.Copy(buf, os.Stdin)

			m, err := ManifestFromBytes(buf.Bytes())

			if err != nil {
				return err
			}

			return compileManifest(m, os.Stdout)
		},
	}
)

func init() {
	sassServerCommand.Flags().Int64Var(&port, "port", 8000, "Port to listen for HTTP server")
	sassServerCommand.Flags().StringVar(&root, "root", ".", "Host manifest files from this directory")

	sassCommand.PersistentFlags().StringVar(&src, "src", "./src", "Location of SASS files")
	sassCommand.AddCommand(sassServerCommand)
	sassCommand.AddCommand(sassOutputCommand)

	RootCmd.AddCommand(sassCommand)
}
