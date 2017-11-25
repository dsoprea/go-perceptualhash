package blockhash

import (
    "image"
    "image/color"
    "math"
    "sort"

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
    bh := &Blockhash{
        image: image,
        hashbits: hashbits,
    }

    err := bh.configure()
    log.PanicIf(err)

    return bh
}

func (bh *Blockhash) configure() (err error) {
    defer func() {
        if state := recover(); state != nil {
            err = log.Wrap(state.(error))
        }
    }()

    cm := bh.image.ColorModel()

    if cm != color.RGBAModel && cm != color.NRGBAModel {
        switch cm {

        // Convert models supporting alpha down to RGBA.
        case color.RGBA64Model:
        case color.AlphaModel:
        case color.Alpha16Model:
        case color.NYCbCrAModel:
            bh.toColor = &color.RGBAModel
            bh.hasAlpha = true

        // Convert models not supporting alpha down to RGB.
        case color.NRGBA64Model:
        case color.GrayModel:
        case color.Gray16Model:
        case color.CMYKModel:
        case color.YCbCrModel:
            bh.toColor = &color.NRGBAModel
            bh.hasAlpha = false

        default:
            log.Panicf("unsupported color model")
        }
    }

    return nil
}

func (bh *Blockhash) totalValue(x, y int) (value float64) {
    c := bh.image.At(x, y)

    if bh.toColor != nil {
        c = (*bh.toColor).Convert(c)
    }

    if bh.hasAlpha {
        r, g, b, a := c.RGBA()
        if a == 0 {
            return 765
        } else {
            return float64(r + g + b)
        }
    } else {
        r, g, b, _ := c.RGBA()
        return float64(r + g + b)
    }
}

func (bh *Blockhash) median(data []float64) float64 {
    sort.Float64s(data)

    len_ := len(data)
    if len(data) % 2 == 0 {
        return data[len_ / 2]
    } else {
        return (data[len_ / 2] + data[len_ / 2 + 1]) / 2.0
    }
}

func (bh *Blockhash) translateBlocksToBits(blocksInline []float64, pixelsPerBlock float64) (results []int) {
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

            blocks[blockTop][blockLeft] += value * weightTop * weightLeft
            blocks[blockTop][blockRight] += value * weightTop * weightRight
            blocks[blockBottom][blockLeft] += value * weightBottom * weightLeft
            blocks[blockBottom][blockRight] += value * weightBottom * weightRight
        }
    }

    blocksInline := make([]float64, bh.hashbits * bh.hashbits)
    for y := 0; y < bh.hashbits; y++ {
        for x := 0; x < bh.hashbits; x++ {
            blocksInline[y * bh.hashbits + x] = blocks[y][x]
        }
    }

    result := bh.translateBlocksToBits(blocksInline, blockWidth * blockHeight)
    result = result

    // return bits_to_hexhash(result)


// TODO(dustin): !! Finish.


// TODO(dustin): Finish.
    bh.digest = ""

    return nil
}

// func (bh *Blockhash) medianf(data float64) float64
// {
//     s := sort.Float64Slice(data)
//     s.Sort()

//     len_ := len(s)
//     if len(s) % 2 {
//         return s[len_ / 2]
//     } else {
//         return (s[len_ / 2] + s[len_ / 2 + 1]) / 2
//     }
// }

// func (bh *Blockhash) translate_blocks_to_bitsf(blocks [][]float64, nblocks int, pixelsPerBlock int) (result []uint32)
// {
//     result = make([]uint32, bh.hashbits * bh.hashbits)

//     halfBlockValue := pixelsPerBlock * 256 * 3 / 2;
//     bandsize := nblocks / 4;

//     for i := 0; i < 4; i++ {
//         m = bh.medianf(&blocks[i * bandsize], bandsize)
//         for j := i * bandsize; j < (i + 1) * bandsize; j++ {
//             v = blocks[j]
//             result[j] = v > m || (math.Abs(v - m) < 1 && m > halfBlockValue)
//         }
//     }

//     return result
// }

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
