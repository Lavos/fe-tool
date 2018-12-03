package cmd

import (
	"os"
	"log"
	"strings"
	"github.com/spf13/cobra"
)

var (
	l = log.New(os.Stderr, "[fe-tool] ", log.LstdFlags)

	port int64
	root string
	src string

	env = make(map[string]string)
)

var RootCmd = &cobra.Command{
	Use: "fe-tool",
}

func init() {
	for _, setting := range os.Environ() {
		pair := strings.SplitN(setting, "=", 2)

		if strings.HasPrefix(pair[0], prefix) {
			env[pair[0]] = pair[1]
		}
	}
}
