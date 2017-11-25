package main

import (
    "os"
    "fmt"
    "image"

    _ "image/jpeg"

    "github.com/dsoprea/go-logging"
    "github.com/jessevdk/go-flags"

    "github.com/dsoprea/go-perceptualhash"
)

var (
    mLog = log.NewLogger("main")
)

type options struct {
    Hashbits int `long:"bits" short:"b" default:"16" description:"Hash bit length"`
    Filepaths []string `long:"filepath" short:"f" required:"true" description:"Image file-path (provide at least once)"`
}

func main() {
    defer func() {
        if state := recover(); state != nil {
            mLog.Errorf(nil, state.(error), "There was an error.")
        }
    }()

    o := new(options)
    if _, err := flags.Parse(o); err != nil {
        os.Exit(1)
    }

    for _, filepath := range o.Filepaths {
        f, err := os.Open(filepath)
        log.PanicIf(err)

        defer f.Close()

        image, _, err := image.Decode(f)
        log.PanicIf(err)

        bh := blockhash.NewBlockhash(image, o.Hashbits)
        digest := bh.Hash()


// TODO(dustin): Debugging.
        fmt.Printf("Digest: %s\n", digest)
    }
}
