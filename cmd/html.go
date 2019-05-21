package cmd

import (
	"io"
	"os"
	"strings"
	"bytes"
	"html/template"

	"github.com/Lavos/fe-tool/lib"
	"github.com/spf13/cobra"
)

var (
	prefix string
)

var (
	htmlCommand = &cobra.Command{
		Use: "html",
		Short: "Build HTML files from templates and environment variables",
		Long: `This command produces static HTML files from Go template files, replacing references to environment variables with their values.`,
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

			htmlBuildContext := lib.NewHTMLBuildContext(src)

			buf := new(bytes.Buffer)
			io.Copy(buf, os.Stdin)

			t := template.New("base")
			t.Delims("<%", "%>")
			t.Funcs(map[string]interface{}{
				"InjectFragments": htmlBuildContext.InjectFragments,
			})

			_, err := t.Parse(buf.String())

			if err != nil {
				return err
			}

			return t.Execute(os.Stdout, env)
		},
	}
)

func init() {
	htmlCommand.AddCommand(htmlOutputCommand)
	htmlCommand.PersistentFlags().StringVar(&src, "src", ".", `Location of fragments`)
	htmlCommand.PersistentFlags().StringVar(&prefix, "prefix", "", `Only environment variables with this prefix will be available to templates`)

	RootCmd.AddCommand(htmlCommand)
}
