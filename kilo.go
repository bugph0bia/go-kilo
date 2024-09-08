package main

import (
	"io"
	"os"
	"syscall"

	"golang.org/x/term"
)

func main() {
	// ターミナルをRAWモードにする
	origTermios, err := term.MakeRaw(syscall.Stdin)
	if err != nil {
		panic(err)
	}
	// プログラム終了時にターミナルのモードを元に戻す
	defer term.Restore(syscall.Stdin, origTermios)

	b := []byte{0}
	for {
		// 標準入力から1バイトずつ読み込む
		_, err := os.Stdin.Read(b)
		// Ctrl+D, q で終了
		if err == io.EOF || b[0] == 'q' {
			break
		}
	}
}
