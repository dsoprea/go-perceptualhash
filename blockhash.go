package blockhash

import (
    "fmt"
    "sort"
    "strconv"
    "image"
    "image/color"
    "math"
    "math/big"

    "github.com/dsoprea/go-logging"
)

type Blockhash struct {
    image image.Image
    hashbits int
    toColor *color.Model
    hasAlpha bool
    digest string
}

func NewBlockhash(image image.Image, hashbits int) *Blockhash {
    return &Blockhash{
        image: image,
        hashbits: hashbits,
    }
}

type opaqueableModel interface {
    Opaque() bool
}


func (bh *Blockhash) totalValue(x, y int) (value uint32) {
    defer func() {
        if state := recover(); state != nil {
            log.Panic(state.(error))
        }
    }()

    p := bh.image.At(x, y)

    // Has the notion of opaqueness, which implies that is supports an alpha
    // channel.
    _, isOpaqueable := p.(opaqueableModel)

    // The RGBA() will return the alpha-multiplied values but the fields will
    // still be in their premultiplied state.
    if bh.image.ColorModel() != color.NRGBAModel {
        p = color.NRGBAModel.Convert(p)
    }

    c2 := p.(color.NRGBA)

    if isOpaqueable == true && c2.A == 0 {
        return 765
    }

    return uint32(c2.R) + uint32(c2.G) + uint32(c2.B)
}

func (bh *Blockhash) median(data []float64) float64 {
    defer func() {
        if state := recover(); state != nil {
            log.Panic(state.(error))
        }
    }()

    sort.Float64s(data)

    len_ := len(data)
    if len(data) % 2 == 0 {
        return data[len_ / 2]
    } else {
        return (data[len_ / 2] + data[len_ / 2 + 1]) / 2.0
    }
}

func (bh *Blockhash) bitsToHexhash(bitString []int) string {
    defer func() {
        if state := recover(); state != nil {
            log.Panic(state.(error))
        }
    }()

    s := make([]byte, len(bitString))

    for i, d := range bitString {
        if d == 0 {
            s[i] = '0'
        } else if d == 1 {
            s[i] = '1'
        } else {
            log.Panicf("invalid bit value (%d) at offset (%d)", d, i)
        }
    }

    b := new(big.Int)
    b.SetString(string(s), 2)

    width := int(math.Floor(float64(bh.hashbits) / float64(4)))
    encoded := fmt.Sprintf("%0" + strconv.Itoa(width) + "x", b)

    return encoded
}

func (bh *Blockhash) translateBlocksToBits(blocksInline []float64, pixelsPerBlock float64) (results []int) {
    defer func() {
        if state := recover(); state != nil {
            log.Panic(state.(error))
        }
    }()

    results = make([]int, len(blocksInline))
    halfBlockValue := pixelsPerBlock * 256.0 * 3.0 / 2.0

    bandsize := int(math.Floor(float64(len(blocksInline)) / 4.0))
    for i := 0; i < 4; i++ {
        m := bh.median(blocksInline[i * bandsize : (i + 1) * bandsize])
        for j := i * bandsize; j < (i + 1) * bandsize; j++ {
            v := blocksInline[j]

// TODO(dustin): Use epsilon.
            if v > m || (math.Abs(v - m) < 1 && m > halfBlockValue) {
                results[j] = 1
            } else {
                results[j] = 0
            }
        }
    }

    return results
}

func (bh *Blockhash) process() (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    if bh.digest != "" {
        return nil
    }

    r := bh.image.Bounds()
    width := r.Max.X
    height := r.Max.Y

    isEvenX := (width % bh.hashbits) == 0
    isEvenY := (height % bh.hashbits) == 0

    blocks := make([][]float64, bh.hashbits)

    for i := 0; i < bh.hashbits; i++ {
        blocks[i] = make([]float64, bh.hashbits)
    }

    blockWidth := float64(width) / float64(bh.hashbits)
    blockHeight := float64(height) / float64(bh.hashbits)

    for y := 0; y < height; y++ {
        var weightTop, weightBottom, weightLeft, weightRight float64
        var blockTop, blockBottom, blockLeft, blockRight int

        if isEvenY {
            blockTop = int(math.Floor(float64(y) / blockHeight))
            blockBottom = blockTop

            weightTop = 1.0
            weightBottom = 0.0
        } else {
            yMod := math.Mod((float64(y) + 1.0), blockHeight)
            yInt, yFrac := math.Modf(yMod)

            weightTop = (1.0 - yFrac)
            weightBottom = yFrac

            // y_int will be 0 on bottom/right borders and on block boundaries
            if yInt > 0.0 || (y + 1) == height {
                blockTop = int(math.Floor(float64(y) / blockHeight))
                blockBottom = blockTop
            } else {
                blockTop = int(math.Floor(float64(y) / blockHeight))
                blockBottom = int(math.Ceil(float64(y) / blockHeight))
            }
        }

        for x := 0; x < width; x++ {
            value := bh.totalValue(x, y)

            if isEvenX {
                blockRight = int(math.Floor(float64(x) / blockWidth))
                blockLeft = blockRight

                weightLeft = 1.0
                weightRight = 0.0
            } else {
                xMod := math.Mod((float64(x) + 1.0), blockWidth)
                xInt, xFrac := math.Modf(xMod)

                weightLeft = (1.0 - xFrac)
                weightRight = (xFrac)

                if xInt > 0.0 || (x + 1) == width {
                    blockRight = int(math.Floor(float64(x) / blockWidth))
                    blockLeft = blockRight
                } else {
                    blockLeft = int(math.Floor(float64(x) / blockWidth))
                    blockRight = int(math.Ceil(float64(x) / blockWidth))
                }
            }

            blocks[blockTop][blockLeft] += float64(value) * weightTop * weightLeft
            blocks[blockTop][blockRight] += float64(value) * weightTop * weightRight
            blocks[blockBottom][blockLeft] += float64(value) * weightBottom * weightLeft
            blocks[blockBottom][blockRight] += float64(value) * weightBottom * weightRight
        }
    }

// TODO(dustin): !! Debug here.
    blocksInline := make([]float64, bh.hashbits * bh.hashbits)
    for y := 0; y < bh.hashbits; y++ {
        for x := 0; x < bh.hashbits; x++ {
            blocksInline[y * bh.hashbits + x] = blocks[y][x]
        }
    }

    result := bh.translateBlocksToBits(blocksInline, blockWidth * blockHeight)
    bh.digest = bh.bitsToHexhash(result)

    return nil
}

func (bh *Blockhash) Hash() string {
    defer func() {
        if state := recover(); state != nil {
            err := log.Wrap(state.(error))
            log.PanicIf(err)
        }
    }()

    err := bh.process()
    log.PanicIf(err)

    return bh.digest
}
