package main

import (
	"io"
	"os"
)

func main() {
	b := []byte{0}
	for {
		_, err := os.Stdin.Read(b)
		if err == io.EOF || b[0] == 'q' {
			break
		}
	}
}
