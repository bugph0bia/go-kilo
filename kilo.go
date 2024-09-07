package main

import (
	"io"
	"os"
)

func main() {
	b := []byte{0}
	for {
		if _, err := os.Stdin.Read(b); err == io.EOF {
			break
		}
	}
}
