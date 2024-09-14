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

// バージョン
const kiloVersion = "0.0.1"

// 矢印キー
const (
	arrowLeft = iota + 1000
	arrowRight
	arrowUp
	arrowDown
	homeKey
	endKey
	pageUp
	pageDown
)

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
func ctrlKey(r int) int {
	b := make([]byte, 4)
	if utf8.EncodeRune(b, rune(r)) != 1 { // r は1バイトのASCIIコードの前提
		panic("failed encode rune")
	}
	return int(b[0] & 0x1F)
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
func editorReadKey() int {
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

	// エスケープシーケンスを処理
	if b[0] == '\x1b' {
		// 後続コードを読み取る
		seq := make([]byte, 3)
		n0, _ := syscall.Read(syscall.Stdin, seq[0:1])
		if n0 != 1 {
			return '\x1b'
		}
		n1, _ := syscall.Read(syscall.Stdin, seq[1:2])
		if n1 != 1 {
			return '\x1b'
		}

		if seq[0] == '[' {
			if seq[1] >= '0' && seq[1] <= '9' {
				// 後続コードを読み取る
				n2, _ := syscall.Read(syscall.Stdin, seq[2:3])
				if n2 != 1 {
					return '\x1b'
				}
				if seq[2] == '~' {
					switch seq[1] {
					case '1', '7':
						return homeKey
					case '4', '8':
						return endKey
					case '5':
						return pageUp
					case '6':
						return pageDown
					}
				}
			} else {
				switch seq[1] {
				case 'A':
					return arrowUp
				case 'B':
					return arrowDown
				case 'C':
					return arrowRight
				case 'D':
					return arrowLeft
				case 'H':
					return homeKey
				case 'F':
					return endKey
				}
			}
		} else if seq[0] == 'O' {
			switch seq[1] {
			case 'H':
				return homeKey
			case 'F':
				return endKey
			}
		}
		return '\x1b'
	}
	return int(b[0])
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

// カーソル移動
func editorMoveCursor(key int) {
	switch key {
	case arrowLeft:
		if ec.cx != 0 {
			ec.cx--
		}
	case arrowRight:
		if ec.cx != ec.screenCols-1 {
			ec.cx++
		}
	case arrowUp:
		if ec.cy != 0 {
			ec.cy--
		}
	case arrowDown:
		if ec.cy != ec.screenRows-1 {
			ec.cy++
		}
	}
}

// キー入力を待ち、入力されたキーに対応する処理を行う
func editorProcessKeypress() bool {
	var quit bool

	// キー入力
	c := editorReadKey()

	switch c {
	case ctrlKey('q'):
		// Ctrl-Q: プログラム終了
		quit = true

	case homeKey:
		ec.cx = 0

	case endKey:
		ec.cx = ec.screenCols - 1

	case pageUp, pageDown:
		// ページ移動
		var c2 int
		if c == pageUp {
			c2 = arrowUp
		} else {
			c2 = arrowDown
		}
		for i := 0; i < ec.screenRows; i++ {
			editorMoveCursor(c2)
		}

	case arrowUp, arrowDown, arrowLeft, arrowRight:
		// カーソル移動
		editorMoveCursor(c)
	}
	return quit
}

/*** init ***/

// 初期化
func initEditor() {
	// カーソル位置初期化
	ec.cx = 0
	ec.cy = 0

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
