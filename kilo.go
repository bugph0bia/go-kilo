package main

import (
	"io"
	"os"
	"syscall"

	"golang.org/x/term"
)

func main() {
	// ターミナルをRAWモードにする
	term.MakeRaw(syscall.Stdin)

	b := []byte{0}
	for {
		// 標準入力から1文字ずつ読み込む
		_, err := os.Stdin.Read(b)
		// Ctrl+D, q で終了
		if err == io.EOF || b[0] == 'q' {
			break
		}
	}
}
