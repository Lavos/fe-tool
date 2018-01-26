package cmd

import (
	"os"
	"log"
	"github.com/spf13/cobra"
)

var (
	l = log.New(os.Stderr, "[fe-tool] ", log.LstdFlags)

	port int64
	root string
	src string
)

var RootCmd = &cobra.Command{
	Use: "fe-tool",
}
