package main

import (
	"fmt"
	"syscall"
	"unicode"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// ターミナルをRAWモードにする
func enableRawMode() *term.State {
	// RAWモード
	origTermios, err := term.MakeRaw(syscall.Stdin)
	if err != nil {
		panic(err)
	}

	// MakeRaw()では、VMIN=1, VTIME=0 に設定されるためここで修正
	termios, err := unix.IoctlGetTermios(syscall.Stdin, unix.TCGETS)
	if err != nil {
		panic(err)
	}
	termios.Cc[unix.VMIN] = 0
	termios.Cc[unix.VTIME] = 1
	if err := unix.IoctlSetTermios(syscall.Stdin, unix.TCSETS, termios); err != nil {
		panic(err)
	}

	return origTermios
}

// ターミナルをRAWモードから復帰する
func disableRawMode(origTermios *term.State) {
	term.Restore(syscall.Stdin, origTermios)
}

func main() {
	// ターミナルをRAWモードにする
	origTermios := enableRawMode()
	// プログラム終了時にターミナルのモードを復帰する
	defer disableRawMode(origTermios)

	for {
		// 標準入力から1バイトずつ読み込む
		b := []byte{0}
		_, err := syscall.Read(syscall.Stdin, b)
		if err != nil && err != syscall.EAGAIN { // Cygwin 対応のために EAGAIN はエラーにしない
			panic(err)
		}
		c := rune(b[0])
		// 読み込んだ文字を画面表示
		if unicode.IsControl(c) {
			fmt.Printf("%d\r\n", c)
		} else {
			fmt.Printf("%d ('%c')\r\n", c, c)
		}
		// q で終了
		if c == 'q' {
			break
		}
	}
}
