# fe-tool

A small development / deployment program for processing SASS, JS, and static HTML templates.

## Installation

A standard `go get` will download and build the binary automatically. Be warned, by default, `libsass` is built from source. This can take a while.

```bash
$ go get github.com/Lavos/fe-tool
```

## Building

```
$ cd ~/go/src/github.com/Lavos/fe-tool
$ go build -o fe-tool
```

## Usage

`fe-tool` is broken up into commands and subcommands, each with an individual use.

A full list of subcommands and other help can be read from the binary itself:

```bash
$ fe-tool
Usage:
  fe-tool [command]

Available Commands:
  help        Help about any command
  html        Build HTML files from templates and environment variables
  js          Mash JavaScript files
  sass        Process SASS files to output mashed CSS
  single      Reads a configuration manifest via STDIN, hosting many types of servers at once, configured via routes.

Flags:
  -h, --help   help for fe-tool

Use "fe-tool [command] --help" for more information about a command.
```
