package cmd

import (
	"os"
	"bytes"
	"io"
	"fmt"
	"strings"
	"net/http"
	"path"

	"goji.io"
	"goji.io/pat"

	"github.com/Lavos/fe-tool/lib"
	qmd "github.com/Lavos/qmd/lib"

	"github.com/spf13/cobra"
)

const (
	TypeJavascript = "javascript"
	TypeSASS = "sass"
	TypeHTML = "html"
	TypeStatic = "static"
)

var (
	singleCommand = &cobra.Command{
		Use:   "single",
		Short: "Reads a configuration manifest via STDIN, hosting many types of servers at once, configured via routes.",
		Example: "fe-tool single < single.manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			buf := new(bytes.Buffer)
			io.Copy(buf, os.Stdin)

			m, err := lib.SingleManifestFromBytes(buf.Bytes())

			if err != nil {
				return err
			}

			// create mux
			mux := goji.NewMux()

			// register middleware
			mux.Use(Logger)

			// add routes from Manifest
			for _, route := range m.Routes {
				fmt.Printf("%#v\n", route)

				switch route.Type {
				case TypeJavascript:
					mux.Handle(pat.Get(route.RequestPath), JavaScriptHandler(route.Manifest, route.Source))

				case TypeSASS:
					mux.Handle(pat.Get(route.RequestPath), SASSHandler(route.Manifest, route.Source))

				case TypeHTML:
					mux.Handle(pat.Get(route.RequestPath), HTMLHandler(route.Source, route.Template, route.Prefix))

				case TypeStatic:
					mux.Handle(pat.Get(route.RequestPath), StaticHandler(route.Source))
				}
			}

			hs := &http.Server{
				Addr:           fmt.Sprintf(":%d", port),
				Handler:        mux,
			}

			fmt.Printf("%#v\n", hs)

			go func(){
				l.Fatal(hs.ListenAndServe())
			}()

			WaitForSignal()
			return nil
		},
	}
)

func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("REQUEST: %s\n", r.URL.String())

		// continue the chain
		handler.ServeHTTP(w, r)
	})
}

func getManifest(manifest_location string) (*lib.Manifest, error) {
	// attempt to open manifest file
	manifest_file, err := os.Open(manifest_location)

	if err != nil {
		return nil, fmt.Errorf("No manifest file found for: %s", manifest_location)
	}

	buf := new(bytes.Buffer)
	io.Copy(buf, manifest_file)

	m, err := lib.ManifestFromBytes(buf.Bytes())

	if err != nil {
		return nil, fmt.Errorf("Could not get manifest: %s", err)
	}

	return m, nil
}

func JavaScriptHandler(manifest_location, source_location string) http.HandlerFunc {
	masher := lib.NewMashContext(source_location)

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		m, err := getManifest(manifest_location)

		if err != nil {
			l.Printf("Get Manifest Error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/javascript")
		err = masher.MashFilesToWriter(m, w)

		if err != nil {
			l.Printf("Mash Error: %s", err)
		}
	})
}

func SASSHandler(manifest_location, source_location string) http.HandlerFunc {
	masher := lib.NewMashContext(source_location)

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		m, err := getManifest(manifest_location)

		if err != nil {
			l.Printf("Get Manifest Error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// create Qmd for docker container
		q := qmd.NewQmd("docker", "run", "--rm", "-i", "-a", "stdin", "-a", "stdout", "codycraven/sassc", "-s")
		stdin, err := q.Cmd.StdinPipe()

		if err != nil {
			l.Printf("Could not get Stdin of Qmd: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		q.Cmd.Stdout = w

		q.Start()

		w.Header().Set("Content-Type", "text/css")
		err = masher.MashFilesToWriter(m, stdin)
		stdin.Close()

		if err != nil {
			l.Printf("Mash Error: %s", err)
		}

		err = q.Wait()

		if err != nil {
			l.Printf("Exit Code: %s", err)
		}
	})
}

func HTMLHandler(source, template_location, prefix string) http.HandlerFunc {
	env := make(map[string]string)

	for _, setting := range os.Environ() {
		pair := strings.SplitN(setting, "=", 2)

		if strings.HasPrefix(pair[0], prefix) {
			env[pair[0]] = pair[1]
		}
	}

	htmlBuildContext := lib.NewHTMLBuildContext(source)

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		err := htmlBuildContext.CompileFile(path.Join(source, template_location), env, w)

		if err != nil {
			l.Printf("Parse Error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func StaticHandler(root string) http.HandlerFunc {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With")

		filename := path.Base(r.URL.Path)
		ext := path.Ext(filename)

		file, err := os.Open(fmt.Sprintf("%s/%s", root, filename))

		if err != nil {
			l.Printf("Could not open file `%s`: %s", filename, err)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		switch ext {
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")

		case ".css":
			w.Header().Set("Content-Type", "text/css")

		case ".js":
			w.Header().Set("Content-Type", "application/javascript")

		case ".mp4":
			w.Header().Set("Content-Type", "video/mp4")

		case ".webm":
			w.Header().Set("Content-Type", "video/webm")
		}

		io.Copy(w, file)
		file.Close()
	})
}

func init () {
	singleCommand.Flags().Int64Var(&port, "port", 8000, `Port to listen for HTTP server`)
	RootCmd.AddCommand(singleCommand)
}
