package cmd

import (
	"os"
	"bytes"
	"io"

	"github.com/Lavos/fe-tool/lib"
	qmd "github.com/Lavos/qmd/lib"

	"github.com/spf13/cobra"
)

var (
	sassCommand = &cobra.Command{
		Use:   "sass",
		Short: "Reads a configuration manifest via STDIN, outputing mashed and compiled CSS from source SASS",
		Example: "fe-tool sass < sass.manifest > compiled.css",
		RunE: func(cmd *cobra.Command, args []string) error {
			buf := new(bytes.Buffer)
			io.Copy(buf, os.Stdin)

			manifest, err := lib.ManifestFromBytes(buf.Bytes())

			if err != nil {
				return err
			}

			// get mashed files
			masher := lib.NewMashContext(src)

			// create Qmd for docker container
			q := qmd.NewQmd("docker", "run", "-i", "--rm", "-a", "stdin", "-a", "stdout", "-a", "stderr", "codycraven/sassc", "-s")
			stdin, err := q.Cmd.StdinPipe()

			if err != nil {
				return err
			}

			q.Cmd.Stdout = os.Stdout
			q.Cmd.Stderr = os.Stderr

			err = q.Start()

			if err != nil {
				return err
			}

			err = masher.MashFilesToWriter(manifest, stdin)
			stdin.Close()

			if err != nil {
				return err
			}

			return q.Wait()
		},
	}
)

func init () {
	sassCommand.PersistentFlags().StringVar(&src, "src", ".", `Location of SASS files`)
	RootCmd.AddCommand(sassCommand)
}
