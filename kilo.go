package main

import (
	"fmt"
	"syscall"
	"unicode"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// Ctrl+英字キーを押したときのコードを返す
func ctrlKey(r rune) rune {
	b := make([]byte, 4)
	if utf8.EncodeRune(b, r) != 1 { // r は1バイトのASCIIコードの前提
		panic("failed encode rune")
	}
	return rune(b[0] & 0x1F)
}

// ターミナルをRAWモードにする
func enableRawMode() *term.State {
	// MakeRaw を使用せずチュートリアルのコードに従うこととする
	//origTermios, err := term.MakeRaw(syscall.Stdin)
	//if err != nil {
	//	panic(err)
	//}

	// 変更前の属性値を取得しておく
	origTermios, err := term.GetState(syscall.Stdin)
	if err != nil {
		panic(err)
	}

	// ターミナルをRAWモードに設定する
	termios, err := unix.IoctlGetTermios(syscall.Stdin, unix.TCGETS)
	if err != nil {
		panic(err)
	}
	termios.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Cflag |= unix.CS8
	termios.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN | unix.ISIG
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
		if c == ctrlKey('q') {
			break
		}
	}
}
