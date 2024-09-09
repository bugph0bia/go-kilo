package main

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"unicode"

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
		if err == io.EOF {
			break
		}
		c := rune(b[0])
		// q で終了
		if c == 'q' {
			break
		}
		if unicode.IsControl(c) {
			fmt.Printf("%d\n", c)
		} else {
			fmt.Printf("%d ('%c')\n", c, c)
		}
	}
}
