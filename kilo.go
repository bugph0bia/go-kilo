package main

/*** imports ***/

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

/*** const ***/

const kiloVersion = "0.0.1"

/*** data ***/

// エディタ状態
type editorConfig struct {
	// カーソル位置
	cx, cy int
	// スクリーンサイズ
	screenRows, screenCols int
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

// カーソル位置を取得
func getCursorPosition() (int, int, error) {
	// カーソル位置を問い合わせ
	_, err := syscall.Write(syscall.Stdin, []byte("\x1b[6n"))
	if err != nil {
		return 0, 0, err
	}

	fmt.Print("\r\n")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	s := scanner.Text()
	if s[0] != '\x1b' || s[1] != '[' {
		return 0, 0, errors.New("failed get cursor position")
	}
	s = s[2 : len(s)-1]
	var rows, cols int
	_, err = fmt.Sscanf(s, "%d;%d", &rows, &cols)
	if err != nil {
		return 0, 0, err
	}

	return rows, cols, nil
}

// ウィンドウサイズを取得
func getWindowsSize() (int, int, error) {
	ws, err := unix.IoctlGetWinsize(syscall.Stdin, unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 {
		// カーソルをスクリーン右下端に移動
		_, err := syscall.Write(syscall.Stdin, []byte("\x1b[999C\x1b[999B"))
		if err != nil {
			return 0, 0, err
		}
		return getCursorPosition()
	}
	return int(ws.Row), int(ws.Col), nil
}

/*** output ***/

// 行を描画
func editorDrawRows(ab *string) {
	for y := 0; y < ec.screenRows; y++ {
		// スクリーンの上から1/3の位置にエディタ名とバージョンを表示
		if y == ec.screenRows/3 {
			welcome := fmt.Sprintf("kilo editor -- version %s", kiloVersion)
			welcomeLen := min(len(welcome), ec.screenCols)
			padding := (ec.screenCols - welcomeLen) / 2
			if padding > 0 {
				*ab += "~"
				padding--
			}
			*ab += strings.Repeat(" ", padding)
			*ab += welcome[:welcomeLen]
		} else {
			*ab += "~"
		}

		*ab += "\x1b[K"
		if y < ec.screenRows-1 {
			*ab += "\r\n"
		}
	}
}

// リフレッシュ
func editorRefreshScreen() {
	// 出力用文字列バッファ
	var ab string

	ab += "\x1b[?25l" // カーソルを非表示
	ab += "\x1b[H"    // カーソル位置を左上へ

	// 行を描画
	editorDrawRows(&ab)

	ab += fmt.Sprintf("\x1b[%d;%dH", ec.cy+1, ec.cx+1) // カーソル位置を設定
	ab += "\x1b[?25h"                                  // カーソルを表示

	syscall.Write(syscall.Stdin, []byte(ab))
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

// 初期化
func initEditor() {
	// カーソル位置初期化
	ec.cx = 10
	ec.cy = 10

	// ウィンドウサイズ取得
	rows, cols, err := getWindowsSize()
	if err != nil {
		panic(err)
	}
	ec.screenRows, ec.screenCols = rows, cols
}

func main() {
	// プログラム終了時にスクリーンを消去
	defer func() {
		syscall.Write(syscall.Stdin, []byte("\x1b[2J"))
		syscall.Write(syscall.Stdin, []byte("\x1b[H"))
	}()

	// ターミナルをRAWモードにする
	enableRawMode()
	defer disableRawMode()

	// 初期化
	initEditor()

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
