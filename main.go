package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
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

	showPNG(bs)
}

func showPNG(bs []byte) {
	defer fmt.Println("")

	startChunk := params{"f": "100", "a": "T"}
	encoded := base64.StdEncoding.EncodeToString(bs)

	if chunkSize >= len(encoded) {
		fmt.Print(msg{encoded, startChunk})
		return
	}

	startChunk["m"] = "1"
	fmt.Print(msg{encoded[:chunkSize], startChunk})

	for i := chunkSize; i < len(encoded); i += chunkSize {
		if i+chunkSize >= len(encoded) {
			fmt.Print(msg{encoded[i:], params{"m": "0"}})
			return
		}
		fmt.Print(msg{encoded[i : i+chunkSize], params{"m": "1"}})
	}
}
