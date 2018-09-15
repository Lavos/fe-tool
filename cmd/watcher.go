package cmd

import (
	"os"
	"bytes"
	"io"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/fsnotify/fsnotify"
)

const (

)

var (
	watcherCommand = &cobra.Command{
		Use:   "watcher",
		Short: "Reads a configuration manifest via STDIN, producing new output files when a referenced file is changed.",
		Example: "fe-tool watcher --output-dir ../dist/ < watcher.manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			buf := new(bytes.Buffer)
			io.Copy(buf, os.Stdin)

			wm, err := WatcherManifestFromBytes(buf.Bytes())

			if err != nil {
				return err
			}

			fmt.Printf("WatcherManifest: %#v\n", wm)

			// create watcher
			watcher, err := fsnotify.NewWatcher()

			if err != nil {
				return err
			}

			defer watcher.Close()

			// handle watcher events
			go watch(watcher, wm)

			fmt.Printf("Watching for changes...\n")

			done := make(chan bool)
			<-done
			return nil
		},
	}
)

func watch(watcher *fsnotify.Watcher, wm *WatcherManifest) {
	// build watch tree
	var file *os.File
	var buf *bytes.Buffer
	var err error
	var loc string

	tree := make(map[string]*WatcherOutput)

	for _, output := range wm.Outputs {
		// get manfest file from reference
		file, err = os.Open(output.ManifestFile)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		buf = new(bytes.Buffer)
		io.Copy(buf, file)
		m, err := ManifestFromBytes(buf.Bytes())

		fmt.Fprintf(os.Stdout, "Manifest: %#v\n", m)
		output.ParsedManifest = m

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		for _, filename := range m.Files {
			loc = fmt.Sprintf("%s/%s", output.Source, filename)

			tree[loc] = &output
			fmt.Printf("Watching: %s\n", loc)
			watcher.Add(loc)
		}
	}

	var o *WatcherOutput
	var ok bool
	var event fsnotify.Event

	for {
		select {
		case event, ok = <-watcher.Events:
			if !ok {
				return
			}

			fmt.Printf("EVENT: %s\n", event)

			if event.Op & fsnotify.Write != fsnotify.Write {
				continue
			}

			// lookup output from child
			o, ok = tree[event.Name]

			if !ok {
				fmt.Printf("Not found in tree: %s\n", event.Name)
				continue
			}

			// open file for writing
			file, err = os.Create(o.FileName)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file `%s` for writing: %s\n", o.FileName, err)
				os.Exit(1)
			}

			fmt.Printf("%#v\n", o)
			fmt.Printf("%#v\n", o.ParsedManifest)

			switch o.ManifestType {
			case TypeJavascript:
				err = mashManifest(o.ParsedManifest, o.Source, file)

			case TypeSASS:
				err = compileManifest(o.ParsedManifest, o.Source, file)
			}

			file.Close()

			fmt.Fprintf(os.Stderr, "%s\n", err)

		case err, ok = <-watcher.Errors:
			if !ok {
				return
			}

			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}
}

func init () {
	RootCmd.AddCommand(watcherCommand)
}
