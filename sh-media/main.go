package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

type Flags struct {
	Name        string
	Filename    string
	UseFilename string
	Replace     bool
}

func ParseFlags() *Flags {
	f := &Flags{}

	flag.StringVar(&f.UseFilename, "filename", "", "override the filename sent to the server")
	flag.BoolVar(&f.Replace, "replace", "", "replace other files with the same name")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: sh-media [flags] <name> <file>")
		fmt.Fprintln(os.Stderr, "")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "")
		statushub.PrintEnvUsage(os.Stderr)
	}

	flag.Parse()
	if len(flag.Args()) != 2 {
		flag.Usage()
		os.Exit(1)
	}
	f.Name = flag.Args()[0]
	f.Filename = flag.Args()[1]

	if f.UseFilename == "" {
		f.UseFilename = filepath.Base(f.Filename)
	}

	return f
}

func main() {
	flags := ParseFlags()

	client, err := statushub.AuthCLI()
	if err != nil {
		essentials.Die("Failed to create client:", err)
	}

	contents, err := ioutil.ReadFile(flags.Filename)
	if err != nil {
		essentials.Die("Failed to read file:", err)
	}

	mime := mime.TypeByExtension(filepath.Ext(flags.UseFilename))
	if mime == "" {
		mime = "application/octet-stream"
	}

	client.AddMedia(flags.Name, flags.UseFilename, mime, contents, flags.Replace)
}
