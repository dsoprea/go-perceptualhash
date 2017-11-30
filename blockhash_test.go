package blockhash

import (
    "testing"
    "os"
    // "fmt"
    "image"
    "path"
    "path/filepath"

    _ "image/jpeg"

    "github.com/dsoprea/go-logging"
)

const (
    testImageJpeg1Big = "20170618_155330.jpg"
    testImagePng1Big = "20170618_155330.png"
    testImageBmp1Big = "20170618_155330.bmp"
    testImageJpeg1Small = "20170618_155330-small.jpg"
    testImagePng1Small = "20170618_155330-small.png"
    testImageBmp1Small = "20170618_155330-small.bmp"
    testImageJpeg2Big = "amazing-mountain-valley-wallpaper-29910-30628-hd-wallpapers.jpg"
    testImageJpeg2Small = "amazing-mountain-valley-wallpaper-29910-30628-hd-wallpapers-small.jpg"
)

var (
    assetsPath = ""
)

func init() {
    var err error
    assetsPath, err = filepath.Abs("test_assets")
    log.PanicIf(err)
}

func getTestBh(filename string) (f *os.File, bh *Blockhash) {
    filepath := path.Join(assetsPath, filename)

    f, err := os.Open(filepath)
    log.PanicIf(err)

    image, _, err := image.Decode(f)
    log.PanicIf(err)

    bh = NewBlockhash(image, 16)

    return f, bh
}

func TestBitsToHexhash_11010010(t *testing.T) {
    f, bh := getTestBh(testImageJpeg1Big)
    defer f.Close()

    digest := bh.bitsToHexhash([]int { 1, 1, 0, 1, 0, 0, 1, 0 })

    if digest != "00d2" {
        t.Errorf("11010010 did not produce the right digest")
    }
}

func TestBitsToHexhash_1(t *testing.T) {
    f, bh := getTestBh(testImageJpeg1Big)
    defer f.Close()

    digest := bh.bitsToHexhash([]int { 1 })

    if digest != "0001" {
        t.Errorf("digest not correct")
    }
}

func TestBitsToHexhash_101(t *testing.T) {
    f, bh := getTestBh(testImageJpeg1Big)
    defer f.Close()

    digest := bh.bitsToHexhash([]int { 1, 0, 1 })

    if digest != "0005" {
        t.Errorf("digest not correct")
    }
}

func TestBitsToHexhash_10101(t *testing.T) {
    f, bh := getTestBh(testImageJpeg1Big)
    defer f.Close()

    digest := bh.bitsToHexhash([]int { 1, 0, 1, 0, 1 })

    if digest != "0015" {
        t.Errorf("digest not correct")
    }
}

// TestBitsToHexhash_10111111111111111 One bit larger than the official bit
// width.
func TestBitsToHexhash_10111111111111111(t *testing.T) {
    f, bh := getTestBh(testImageJpeg1Big)
    defer f.Close()

    digest := bh.bitsToHexhash([]int { 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1 })

    if digest != "17fff" {
        t.Errorf("digest not correct")
    }
}

// TestBitsToHexhash_1111111111111111 Exactly the right size..
func TestBitsToHexhash_1101111111011111(t *testing.T) {
    f, bh := getTestBh(testImageJpeg1Big)
    defer f.Close()

    digest := bh.bitsToHexhash([]int { 1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1 })

    if digest != "dfdf" {
        t.Errorf("digest not correct")
    }
}
