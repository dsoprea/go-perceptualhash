package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "image/jpeg"
	_ "image/png"

	"github.com/dsoprea/go-logging"
	"github.com/jessevdk/go-flags"

	"github.com/dsoprea/go-perceptualhash"
)

var (
	mLog = log.NewLogger("main")
)

type options struct {
	Hashbits  int      `long:"bits" short:"b" default:"16" description:"Hash bit length (N^2)"`
	Filepaths []string `long:"filepath" short:"f" required:"true" description:"Image file-path (provide at least once)"`
	Digest    bool     `long:"digest" short:"d" description:"Just print digest (no filenames)"`
}

func main() {
	defer func() {
		if state := recover(); state != nil {
			log.Panic(state)
		}
	}()

	o := new(options)
	if _, err := flags.Parse(o); err != nil {
		os.Exit(1)
	}

	len_ := 0
	for _, filepath := range o.Filepaths {
		len_ = int(math.Max(float64(len_), float64(len(filepath))))
	}

	for _, filepath := range o.Filepaths {
		f, err := os.Open(filepath)
		log.PanicIf(err)

		defer f.Close()

		image, _, err := image.Decode(f)
		log.PanicIf(err)

		bh := blockhash.NewBlockhash(image, o.Hashbits)
		hexdigest := bh.Hexdigest()

		if o.Digest {
			fmt.Println(hexdigest)
		} else {
			fmt.Printf("%s%s %s\n", filepath, strings.Repeat(" ", len_-len(filepath)), hexdigest)
		}
	}
}
