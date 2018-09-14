package cmd

import (
	"io"
	"os"
	"fmt"
	"strings"
	"bytes"
	"path/filepath"
	"html/template"
	"net/http"
	"mime"
	"time"
	"path"
	"flag"

	"github.com/spf13/cobra"
)

var (
	fm = template.FuncMap{
		"InjectFragments": InjectFragments,
	}

	prefix string
	env = make(map[string]string)
)

func InjectFragments(pattern string) (template.HTML, error) {
	var html template.HTML
	filenames, err := filepath.Glob(pattern)

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

func CompileFile(template_path string, data interface{}, w io.Writer) error {
	_, filename := path.Split(template_path)

	t := template.New("base")
	t.Delims("<%", "%>")
	t.Funcs(fm)

	_, err := t.ParseFiles(template_path)

	if err != nil {
		return err
	}

	return t.ExecuteTemplate(w, filename, data)
}

func htmlHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	upath := req.URL.Path

	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}

	if strings.HasSuffix(upath, "/") {
		upath = upath + "index.html"
	}

	template_path := fmt.Sprintf("%s%s.template", root, path.Clean(upath))
	err := CompileFile(template_path, env, w)

	if err != nil {
		l.Printf("Compile Error: %s", err)
	}
}

var (
	htmlCommand = &cobra.Command{
		Use: "html",
		Short: "Build HTML files from templates and environment variables",
		Long: `This command produces static HTML files from Go template files, replacing references to environment variables with their values.`,
	}

	htmlServerCommand = &cobra.Command{
		Use:   "server",
		Short: "Host a HTTP server for returning processed templates on request",
		Example: "fe-tool html server --port 9000 --prefix WEBSITE --root template-location/",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, setting := range os.Environ() {
				pair := strings.SplitN(setting, "=", 2)

				if strings.HasPrefix(pair[0], prefix) {
					env[pair[0]] = pair[1]
				}
			}

			mime.AddExtensionType(".html", "text/html")

			l.Printf("Listening at %d, hosting %s", port, root)

			s := &http.Server{
				Addr:           fmt.Sprintf(":%d", port),
				Handler:        http.HandlerFunc(htmlHandler),
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

	htmlOutputCommand = &cobra.Command{
		Use:   "output",
		Short: "Parses STDIN as template against environment variables, outputing processed template to STDOUT",
		Example: "fe-tool html output --prefix WEBSITE < template-location/index.template > deploy/index.html",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, setting := range os.Environ() {
				pair := strings.SplitN(setting, "=", 2)

				if strings.HasPrefix(pair[0], prefix) {
					env[pair[0]] = pair[1]
				}
			}

			flag.Parse()

			buf := new(bytes.Buffer)
			io.Copy(buf, os.Stdin)

			t := template.New("base")
			t.Delims("<%", "%>")
			t.Funcs(fm)

			_, err := t.Parse(buf.String())

			if err != nil {
				return err
			}

			return t.Execute(os.Stdout, env)
		},
	}
)

func init() {
	htmlServerCommand.Flags().Int64Var(&port, "port", 8000, `Port to listen for HTTP server`)
	htmlServerCommand.Flags().StringVar(&root, "root", ".", `Host template files from this directory`)

	htmlCommand.AddCommand(htmlServerCommand)
	htmlCommand.AddCommand(htmlOutputCommand)
	htmlCommand.PersistentFlags().StringVar(&prefix, "prefix", "", `Only environment variables with this prefix will be available to templates`)

	RootCmd.AddCommand(htmlCommand)
}
