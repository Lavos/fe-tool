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
)

func mashManifest(m *Manifest, src_location string, w io.Writer) error {
	for _, name := range m.Files {
		file, err := os.Open(src_location + "/" + name)

		if err != nil {
			return err
		}

		io.Copy(w, file)
		file.Close()
	}

	return nil
}

func jsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/javascript")

	upath := req.URL.Path

	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}

	parts := strings.Split(upath, ".")

	if len(parts) == 0 {
		l.Printf("No manifest file found for: %s", upath)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	upath = path.Clean(root + parts[0] + ".manifest")

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

	err = mashManifest(m, src, w)

	if err != nil {
		l.Printf("Mash Error: %s", err)
	}
}


var (
	jsCommand = &cobra.Command{
		Use: "js",
		Short: "Mash JavaScript files",
	}

	jsOutputCommand = &cobra.Command{
		Use:   "output",
		Short: "Parses STDIN as manifest, returning mashed files to STDOUT",
		Example: "fe-tool js --src javascript-location/ < javascript-location/prod.manifest > main.js",
		RunE: func(cmd *cobra.Command, args []string) error {
			buf := new(bytes.Buffer)
			io.Copy(buf, os.Stdin)

			m, err := ManifestFromBytes(buf.Bytes())

			if err != nil {
				return err
			}

			return mashManifest(m, src, os.Stdout)
		},
	}

	jsServerCommand = &cobra.Command{
		Use:   "server",
		Short: "Host a HTTP server, returning mashed JavaScript files on request",
		Example: "fe-tool js --src javascript-location/ --root javascript-location/ --port 9000",
		RunE: func(cmd *cobra.Command, args []string) error {
			mime.AddExtensionType(".js", "text/javascript")

			l.Printf("Listening at %d, hosting manifests from: %s", port, root)

			s := &http.Server{
				Addr:           fmt.Sprintf(":%d", port),
				Handler:        http.HandlerFunc(jsHandler),
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
)

func init() {
	jsCommand.AddCommand(jsOutputCommand)
	jsCommand.AddCommand(jsServerCommand)

	jsServerCommand.Flags().Int64Var(&port, "port", 8000, "Port to listen for HTTP server")
	jsServerCommand.Flags().StringVar(&root, "root", ".", "Host manifest files from this directory")

	jsCommand.PersistentFlags().StringVar(&src, "src", "./src", "Location of referenced JavaScript files")
	RootCmd.AddCommand(jsCommand)
}
