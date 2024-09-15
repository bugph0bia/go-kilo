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

// 特殊キー
const (
	arrowLeft = iota + 1000
	arrowRight
	arrowUp
	arrowDown
	delKey
	homeKey
	endKey
	pageUp
	pageDown
)

/*** data ***/

// エディタ行バッファ
type eRow struct {
	// テキスト
	chars string
}

// エディタステータス
type editorConfig struct {
	// カーソル位置
	cx, cy int
	// オフセット
	rowOff, colOff int
	// スクリーンサイズ
	screenRows, screenCols int
	// ターミナルの初期モード
	origTermios *term.State
	// 行バッファ
	row []eRow
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
					case '3':
						return delKey
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
	// カーソル位置問い合わせクエリ
	_, err := syscall.Write(syscall.Stdin, []byte("\x1b[6n"))
	if err != nil {
		return 0, 0, err
	}

	fmt.Print("\r\n")

	// カーソル位置の結果を取得
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
	// システムコールからウィンドウサイズを取得
	ws, err := unix.IoctlGetWinsize(syscall.Stdin, unix.TIOCGWINSZ)
	if err != nil || ws.Col == 0 {
		// ウィンドウサイズ取得の予備的な手段を用意
		// カーソルをスクリーン右下端に移動して位置を取得する方法
		_, err := syscall.Write(syscall.Stdin, []byte("\x1b[999C\x1b[999B"))
		if err != nil {
			return 0, 0, err
		}
		return getCursorPosition()
	}
	return int(ws.Row), int(ws.Col), nil
}

/*** row operations ***/

// 行データ追加
func editorAppendRow(s string) {
	ec.row = append(ec.row, eRow{chars: s})
}

/*** file i/o ***/

// ファイル読み込み
func editorOpen(fileName string) {
	// ファイルオープン
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 全行読み込み
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		editorAppendRow(scanner.Text())
	}
}

/*** output ***/

// スクロール処理
func editorScroll() {
	// 上方向
	if ec.cy < ec.rowOff {
		ec.rowOff = ec.cy
	}
	// 下方向
	if ec.cy >= ec.rowOff+ec.screenRows {
		ec.rowOff = ec.cy - ec.screenRows + 1
	}
	// 左方向
	if ec.cx < ec.colOff {
		ec.colOff = ec.cx
	}
	// 右方向
	if ec.cx >= ec.colOff+ec.screenCols {
		ec.colOff = ec.cx - ec.screenCols + 1
	}
}

// 行を描画
func editorDrawRows(ab *string) {
	for y := 0; y < ec.screenRows; y++ {
		fileRow := y + ec.rowOff
		// ブランク行の表示
		if fileRow >= len(ec.row) {
			// 表示するテキストデータが無い（ブランクで起動している）状態であれば、
			// スクリーンの上から1/3の位置にエディタ名とバージョンを表示する
			if len(ec.row) == 0 && y == ec.screenRows/3 {
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
				// ブランク行は ~ で埋める
				*ab += "~"
			}
		} else {
			// 行バッファの内容を出力
			rowLen := max(len(ec.row[fileRow].chars)-ec.colOff, 0)
			rowLen = min(rowLen, ec.screenCols)
			if rowLen > 0 {
				*ab += ec.row[fileRow].chars[ec.colOff : ec.colOff+rowLen]
			}
		}

		// カーソル位置を復帰して改行
		*ab += "\x1b[K"
		if y < ec.screenRows-1 {
			*ab += "\r\n"
		}
	}
}

// リフレッシュ
func editorRefreshScreen() {
	// スクロール処理
	editorScroll()

	// 出力用文字列バッファ
	var ab string

	// カーソルを非表示にして左上へ
	ab += "\x1b[?25l"
	ab += "\x1b[H"

	// 行を描画
	editorDrawRows(&ab)

	// カーソルを指定位置に移動して表示
	ab += fmt.Sprintf("\x1b[%d;%dH", (ec.cy-ec.rowOff)+1, (ec.cx-ec.colOff)+1)
	ab += "\x1b[?25h"

	// テキストバッファの内容を出力
	syscall.Write(syscall.Stdout, []byte(ab))
}

/*** input ***/

// カーソル移動
func editorMoveCursor(key int) {
	// 現在行
	var row string
	if ec.cy < len(ec.row) {
		row = ec.row[ec.cy].chars
	}

	switch key {
	case arrowLeft:
		if ec.cx != 0 {
			ec.cx--
		} else if ec.cy > 0 {
			// 行頭からは前の行の末尾へ移動
			ec.cy--
			ec.cx = len(ec.row[ec.cy].chars)
		}
	case arrowRight:
		if ec.cx < len(row) {
			ec.cx++
		} else if ec.cy < len(ec.row) {
			ec.cy++
			ec.cx = 0
		}
	case arrowUp:
		if ec.cy != 0 {
			ec.cy--
		}
	case arrowDown:
		if ec.cy < len(ec.row) {
			ec.cy++
		}
	}

	// 新しい行の末尾にカーソルをスナップ
	if ec.cy < len(ec.row) {
		row = ec.row[ec.cy].chars
	}
	ec.cx = min(ec.cx, len(row))
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
		// カーソルを左端へ移動
		ec.cx = 0

	case endKey:
		// カーソルを右端へ移動
		ec.cx = ec.screenCols - 1

	case pageUp, pageDown:
		// カーソルを上端／下端へ移動
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
		// カーソルを上下左右に移動
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
	ec.rowOff = 0
	ec.colOff = 0

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
	// ファイル読み込み
	if len(os.Args) >= 2 {
		editorOpen(os.Args[1])
	}

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
