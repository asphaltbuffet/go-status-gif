// Package cmd contains all CLI commands used by the application.
package cmd

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type schedule struct {
	color    uint8
	duration int
}

var colors = map[string]color.Color{
	"white":     color.White, // this is used for background
	"black":     color.Black, // this is used for border
	"red":       color.RGBA{255, 0, 0, 255},
	"redoff":    color.RGBA{127, 0, 0, 255},
	"green":     color.RGBA{0, 255, 0, 255},
	"greenoff":  color.RGBA{0, 127, 0, 255},
	"blue":      color.RGBA{0, 0, 255, 255},
	"blueoff":   color.RGBA{0, 0, 127, 255},
	"yellow":    color.RGBA{0, 255, 255, 255},
	"yellowoff": color.RGBA{0, 127, 127, 255},
}

var (
	outfile       = "status.gif"
	patternInput  = "red 50 redoff 50 red 50 redoff 50 red 50 redoff 500"
	size          = 500
	border        = 25
	pattern       = []schedule{}
	colorTracking = map[string]int{"white": 0, "black": 1}
)

var rootCmd = &cobra.Command{
	Use: "go-status-gif",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Create(outfile)
		if err != nil {
			panic(err)
		}
		defer func() {
			closeErr := f.Close()
			if closeErr != nil {
				log.Error("failed to close output file: ", closeErr)
			}
		}()

		p := strings.Split(patternInput, " ")

		for i := 0; i < len(p); i += 2 {
			d, _ := strconv.Atoi(p[i+1])

			_, exists := colorTracking[p[i]]
			if !exists {
				colorTracking[p[i]] = len(colorTracking)
			}

			// debug print statement
			fmt.Printf("%v\n", colorTracking)

			pattern = append(pattern, schedule{
				color:    uint8(colorTracking[p[i]]),
				duration: d,
			})
		}

		// cmd.Println(pattern)

		createBasicGif(f, size, border)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outfile, "outfile", "o", outfile, "output file")
	rootCmd.PersistentFlags().StringVarP(&patternInput, "pattern", "p", patternInput, "color pattern '<color> <1/100th seconds> ...'")
	rootCmd.PersistentFlags().IntVarP(&size, "size", "s", size, "image size in pixels (size x size)")
	rootCmd.PersistentFlags().IntVarP(&border, "border", "b", border, "image border size in pixels")
}

func createBasicGif(out io.Writer, size, margin int) {
	radius := (size - margin*2) / 2

	// translate colors used to indexed palette
	palette := []color.Color{}
	for i := 0; i < len(colorTracking); i++ {
		for c := range colorTracking {
			if i == colorTracking[c] {
				palette = append(palette, colors[c])
			}
		}
	}

	rect := image.Rect(0, 0, size-1, size-1)

	images := []*image.Paletted{}
	delay := []int{}

	for i := 0; i < len(pattern); i++ {
		images = append(images, image.NewPaletted(rect, palette))
		delay = append(delay, pattern[i].duration)
	}

	// debug print statement
	fmt.Printf("%v\n", colorTracking)

	for x := margin; x < size-margin; x++ {
		for y := margin; y < size-margin; y++ {
			d := int(distanceFromCenter(x, y, size/2, size/2))
			switch {
			case d > radius: // leave background color
			case d > radius-10: // set border to black
				for j := 0; j < len(pattern); j++ {
					images[j].SetColorIndex(x, y, 1)
				}
			default:
				for j := 0; j < len(pattern); j++ {
					images[j].SetColorIndex(x, y, pattern[j].color)
				}
			}
		}
	}

	anim := gif.GIF{
		LoopCount: 0,
		Delay:     delay,
		// Image:     []*image.Paletted{img, img2}}
		Image: images,
	}

	_ = gif.EncodeAll(out, &anim)
}

func distanceFromCenter(x, y, centerX, centerY int) float64 {
	return math.Sqrt(float64((x-centerX)*(x-centerX) + (y-centerY)*(y-centerY)))
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
