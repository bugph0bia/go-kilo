package main

/*** imports ***/

import (
	"syscall"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

/*** data ***/

// エディタ状態
type editorConfig struct {
	// ターミナルの初期モード
	origTermios *term.State
}

var ec editorConfig

/*** terminal ***/

// Ctrl+英字キーを押したときのコードを返す
func ctrlKey(r rune) rune {
	b := make([]byte, 4)
	if utf8.EncodeRune(b, r) != 1 { // r は1バイトのASCIIコードの前提
		panic("failed encode rune")
	}
	return rune(b[0] & 0x1F)
}

// ターミナルをRAWモードから復帰する
func disableRawMode() {
	term.Restore(syscall.Stdin, ec.origTermios)
}

// ターミナルをRAWモードにする
func enableRawMode() {
	// MakeRaw を使用せずチュートリアルのコードに従うこととする
	//t, err := term.MakeRaw(syscall.Stdin)
	//if err != nil {
	//	panic(err)
	//}

	// 変更前の属性値を取得しておく
	t, err := term.GetState(syscall.Stdin)
	if err != nil {
		panic(err)
	}
	ec.origTermios = t

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
}

// キー入力を待ち、入力結果を返す
func editorReadKey() rune {
	b := []byte{0}
	for {
		nread, err := syscall.Read(syscall.Stdin, b)
		if err != nil && err != syscall.EAGAIN {
			panic(err)
		}
		if nread == 1 {
			break
		}
	}
	return rune(b[0])
}

/*** output ***/

// 行を描画
func editorDrawRows() {
	for y := 0; y < 24; y++ {
		syscall.Write(syscall.Stdin, []byte("~\r\n"))
	}
}

// リフレッシュ
func editorRefreshScreen() {
	// スクリーンを消去
	syscall.Write(syscall.Stdin, []byte("\x1b[2J"))
	syscall.Write(syscall.Stdin, []byte("\x1b[H"))

	// 行を描画
	editorDrawRows()

	syscall.Write(syscall.Stdin, []byte("\x1b[H"))
}

/*** input ***/

// キー入力を待ち、入力されたキーに対応する処理を行う
func editorProcessKeypress() bool {
	var quit bool

	// キー入力
	c := editorReadKey()

	switch c {
	// Ctrl-Q: プログラム終了
	case ctrlKey('q'):
		quit = true
	}
	return quit
}

/*** init ***/

func main() {
	// プログラム終了時にスクリーンを消去
	defer func() {
		syscall.Write(syscall.Stdin, []byte("\x1b[2J"))
		syscall.Write(syscall.Stdin, []byte("\x1b[H"))
	}()

	// ターミナルをRAWモードにする
	enableRawMode()
	defer disableRawMode()

	// メインループ
	for {
		// リフレッシュ
		editorRefreshScreen()
		// 入力されたキーに対応する処理を行う
		if quit := editorProcessKeypress(); quit {
			break
		}
	}
}
