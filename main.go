package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	_ "image/gif"
	_ "image/jpeg"

	"github.com/mattn/go-tty"
	"github.com/nfnt/resize"
)

const chunkSize = 4096

type msg struct {
	payload string
	params  params
}

func (m msg) String() string {
	return fmt.Sprintf("\033_G%s;%s\033\\", m.params, m.payload)
}

type params map[string]string

func (p params) String() string {
	opts := []string{}
	for k, v := range p {
		opts = append(opts, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(opts, ",")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Use: %s /path/to/file\n", os.Args[0])
		os.Exit(1)
	}

	bs, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	if err := showImage(bs); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func showImage(bs []byte) error {
	defer fmt.Println("")
	startChunk := params{"f": "100", "a": "T"}

	buf := bytes.NewBuffer(bs)
	img, imgType, err := image.Decode(buf)
	if err != nil {
		return err
	}

	t, err := tty.Open()
	if err != nil {
		return err
	}

	_, _, w, h, err := t.SizePixel()
	if err != nil {
		return err
	}

	var resized bool
	if w < img.Bounds().Dx() || h < img.Bounds().Dy() {
		img = resizeToMax(w, h, img)
		resized = true
	}

	if imgType == "png" && !resized {
		buf.Reset()
		buf.Write(bs)
	} else {
		if err := png.Encode(buf, img); err != nil {
			return err
		}
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	if chunkSize >= len(encoded) {
		fmt.Print(msg{encoded, startChunk})
		return nil
	}

	startChunk["m"] = "1"
	fmt.Print(msg{encoded[:chunkSize], startChunk})

	for i := chunkSize; i < len(encoded); i += chunkSize {
		if i+chunkSize >= len(encoded) {
			fmt.Print(msg{encoded[i:], params{"m": "0"}})
			return nil
		}
		fmt.Print(msg{encoded[i : i+chunkSize], params{"m": "1"}})
	}

	return nil
}

func resizeToMax(w, h int, img image.Image) image.Image {
	x, y := img.Bounds().Dx(), img.Bounds().Dy()
	if w < x {
		y = y * (w / x)
		x = w
	}
	if h < y {
		x = x * (h / y)
		y = h
	}

	return resize.Resize(uint(x), uint(y), img, resize.Bilinear)
}
