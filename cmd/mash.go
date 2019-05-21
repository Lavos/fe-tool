package cmd

import (
	"os"
	"bytes"
	"io"

	"github.com/Lavos/fe-tool/lib"

	"github.com/spf13/cobra"
)

var (
	mashCommand = &cobra.Command{
		Use: "mash",
		Short: "Parses STDIN as manifest, returning mashed files to STDOUT",
		Example: "fe-tool mash --src javascript-location/ < javascript-location/prod.manifest > main.js",
		RunE: func(cmd *cobra.Command, args []string) error {
			buf := new(bytes.Buffer)
			io.Copy(buf, os.Stdin)

			m, err := lib.ManifestFromBytes(buf.Bytes())

			if err != nil {
				return err
			}

			masher := lib.NewMashContext(src)
			return masher.MashFilesToWriter(m, os.Stdout)
		},
	}
)

func init() {
	mashCommand.PersistentFlags().StringVar(&src, "src", ".", "Location of referenced manifest files")
	RootCmd.AddCommand(mashCommand)
}
