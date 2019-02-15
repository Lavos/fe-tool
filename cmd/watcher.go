package cmd

import (
	"os"
	"bytes"
	"io"
	"fmt"
	"strings"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/fsnotify/fsnotify"
)

const (

)

var (

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
	var o *WatcherOutput

	var filenames []string

	tree := make(map[string]*WatcherOutput)

	for _, output := range wm.Outputs {
		o = &WatcherOutput{
			FileName: output.FileName,
			ManifestType: output.ManifestType,
			ManifestFile: output.ManifestFile,
			Source: output.Source,
			TemplateFile: output.TemplateFile,
			Prefix: output.Prefix,
			WatchGlobs: output.WatchGlobs,
		}

		switch output.ManifestType {

		case TypeJavascript, TypeSASS:
			// get manifest file from reference
			file, err = os.Open(output.ManifestFile)

			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}

			buf = new(bytes.Buffer)
			io.Copy(buf, file)
			m, err := ManifestFromBytes(buf.Bytes())

			fmt.Fprintf(os.Stdout, "Manifest: %#v\n", m)

			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}

			o.ParsedManifest = m

			for _, filename := range m.Files {
				loc = filepath.ToSlash(fmt.Sprintf("%s/%s", output.Source, filename))

				tree[loc] = o
				fmt.Printf("Watching: %s\n", loc)
				watcher.Add(loc)
			}

		case TypeHTML:
			// watch the template as well
			watcher.Add(o.TemplateFile)
			tree[o.TemplateFile] = o
			fmt.Printf("Watching: %s\n", o.TemplateFile)

			for _, glob := range o.WatchGlobs {
				filenames, err = filepath.Glob(glob)

				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}

				for _, filename := range filenames {
					tree[filename] = o
					fmt.Printf("Watching: %s\n", filename)
					watcher.Add(filename)
				}
			}
		}

		err = writeOutput(o)
		fmt.Fprintf(os.Stderr, "Tree Output Error: %s\n", err)
	}

	fmt.Printf("Tree: %#v\n", tree)

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

			// standardize filenames for map-lookup
			loc = filepath.ToSlash(event.Name)
			loc = strings.Join(filepath.SplitList(loc), "/")

			fmt.Printf("LOC: %s\n", loc)

			// lookup output from child
			o, ok = tree[loc]

			if !ok {
				fmt.Printf("Not found in tree: %s\n", loc)
				continue
			}

			err = writeOutput(o)
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

func writeOutput(o *WatcherOutput) error {
	fmt.Printf("test\n")
	// open file for writing
	file, err := os.Create(o.FileName)

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

	case TypeHTML:
		err = CompileFile(o.TemplateFile, env, file)
	}

	file.Close()

	return err
}

func init () {
	RootCmd.AddCommand(watcherCommand)
}
