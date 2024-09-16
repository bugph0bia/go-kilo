package main

/*** imports ***/

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

/*** const ***/

// バージョン
const kiloVersion = "0.0.1"

// タブストップ数
const kiloTabStop = 8

// 未保存の終了時のキー押下要求回数
const kiloQuitTimes = 3

// 特殊キー
const (
	backspace = 127
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
	// レンダリング
	render string
}

// エディタステータス
type editorConfig struct {
	// カーソル位置
	cx, cy int // テキスト上の位置
	rx     int // 画面上のレンダリング位置
	// オフセット
	rowOff, colOff int
	// スクリーンサイズ
	screenRows, screenCols int
	// 行バッファ
	row []eRow
	// ダーティフラグ
	dirty int
	// ファイル名
	fileName string
	// ステータスメッセージ
	statusMsg     string
	statusMsgTime time.Time
	// ターミナルの初期モード
	origTermios *term.State
}

var ec editorConfig

/*** terminal ***/

// Ctrl+英字キーを押したときのコードを返す
func ctrlKey(r rune) rune {
	b := make([]byte, 4)
	if utf8.EncodeRune(b, rune(r)) != 1 { // r は1バイトのASCIIコードの前提
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
	return rune(b[0])
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

// 行内の位置 cx から rx を算出する
func editorRowCxToRx(row eRow, cx int) int {
	rx := 0
	for j := 0; j < cx; j++ {
		// タブ文字に相当するスペースの数を計算
		if row.chars[j] == '\t' {
			rx += (kiloTabStop - 1) - (rx % kiloTabStop)
		}
		rx++
	}
	return rx
}

// 更新済みの行データを返す
func editorUpdateRow(row *eRow) {
	// タブ文字をスペースに変換（8タブ）
	tabs := strings.Count(row.chars, "\t")
	render := make([]byte, 0, len(row.chars)+tabs*(kiloTabStop-1)) // 予め必要なバイト数をキャパシティに確保しておく
	for j := 0; j < len(row.chars); j++ {
		if row.chars[j] == '\t' {
			render = append(render, ' ')
			for len(render)%kiloTabStop != 0 {
				render = append(render, ' ')
			}
		} else {
			render = append(render, row.chars[j])
		}
	}
	row.render = string(render)
}

// 行データ追加
func editorAppendRow(s string) {
	row := eRow{chars: s}
	editorUpdateRow(&row)
	ec.row = append(ec.row, row)

	ec.dirty++
}

// 行データを削除
func editorDelRow(at int) {
	// 対象が範囲外なら終了
	if at < 0 || at >= len(ec.row) {
		return
	}
	ec.row = ec.row[:at+(copy(ec.row[at:], ec.row[at+1:]))] // スライスからインデックス at の要素を削除して前詰め

	ec.dirty++
}

// 行データに文字を挿入
func editorRowInsertChar(row *eRow, at int, c rune) {
	// カーソル位置が行の範囲外なら末尾へ追加
	if at < 0 || at > len(row.chars) {
		at = len(row.chars)
	}
	// 文字を挿入
	row.chars = row.chars[:at] + string(c) + row.chars[at:]
	editorUpdateRow(row)

	ec.dirty++
}

// 行データに文字列を連結
func editorRowAppendString(row *eRow, s string) {
	// 行データの末尾に文字列を連結
	row.chars += s
	editorUpdateRow(row)

	ec.dirty++
}

// 行データから文字を削除
func editorRowDelChar(row *eRow, at int) {
	// カーソル位置が行の範囲外なら終了
	if at < 0 || at >= len(row.chars) {
		return
	}
	// 文字を削除
	row.chars = row.chars[:at] + row.chars[at+1:]
	editorUpdateRow(row)

	ec.dirty++
}

/*** editor operations ***/

// 文字を挿入
func editorInsertChar(c rune) {
	// カーソル行が最終行の次であれば新しい行データを挿入
	if ec.cy == len(ec.row) {
		editorAppendRow("")
	}
	// 行データに文字を挿入
	editorRowInsertChar(&ec.row[ec.cy], ec.cx, c)
	ec.cx++
}

// 文字を削除
func editorDelChar() {
	// カーソル行が最終行の次であれば終了
	if ec.cy == len(ec.row) {
		return
	}
	// カーソルがファイルの先頭であれば終了
	if ec.cx == 0 && ec.cy == 0 {
		return
	}

	if ec.cx > 0 {
		// カーソルが行の途中であれば、行内で1文字削除する
		editorRowDelChar(&ec.row[ec.cy], ec.cx-1)
		ec.cx--
	} else {
		// カーソルが行頭であれば、現在行を前の行の末尾に連結する
		editorRowAppendString(&ec.row[ec.cy-1], ec.row[ec.cy].chars)
		editorDelRow(ec.cy)
		ec.cy--
	}
}

/*** file i/o ***/

// 行バッファの内容を文字列に変換
func editorRowsToString() string {
	lines := make([]string, 0)
	for _, r := range ec.row {
		lines = append(lines, r.chars)
	}
	return strings.Join(lines, "\n") + "\n"
}

// ファイル読み込み
func editorOpen(fileName string) {
	ec.fileName = fileName

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

	ec.dirty = 0
}

// ファイル保存
func editorSave() {
	if ec.fileName == "" {
		return
	}

	// 行バッファの内容を取得
	s := editorRowsToString()

	// ファイルオープン
	f, err := os.OpenFile(ec.fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		// エラーメッセージ表示
		editorSetStatusMessage("Can't save! I/O error: %s", err.Error())
		return
	}
	defer f.Close()

	// ファイルの中身を指定の長さで切り捨て
	err = f.Truncate(int64(len(s)))
	if err != nil {
		// エラーメッセージ表示
		editorSetStatusMessage("Can't save! I/O error: %s", err.Error())
		return
	}

	// ファイルへ書き込み
	_, err = fmt.Fprint(f, s)
	if err != nil {
		// エラーメッセージ表示
		editorSetStatusMessage("Can't save! I/O error: %s", err.Error())
		return
	}

	// 保存完了メッセージ表示
	editorSetStatusMessage("%d bytes written to disk", len(s))

	ec.dirty = 0
}

/*** output ***/

// スクロール処理
func editorScroll() {
	ec.rx = 0
	// cx から rx を算出
	if ec.cy < len(ec.row) {
		ec.rx = editorRowCxToRx(ec.row[ec.cy], ec.cx)
	}

	// 上方向
	if ec.cy < ec.rowOff {
		ec.rowOff = ec.cy
	}
	// 下方向
	if ec.cy >= ec.rowOff+ec.screenRows {
		ec.rowOff = ec.cy - ec.screenRows + 1
	}
	// 左方向
	if ec.rx < ec.colOff {
		ec.colOff = ec.rx
	}
	// 右方向
	if ec.rx >= ec.colOff+ec.screenCols {
		ec.colOff = ec.rx - ec.screenCols + 1
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
			rowLen := max(len(ec.row[fileRow].render)-ec.colOff, 0)
			rowLen = min(rowLen, ec.screenCols)
			if rowLen > 0 {
				*ab += ec.row[fileRow].render[ec.colOff : ec.colOff+rowLen]
			}
		}

		// カーソル位置を復帰して改行
		*ab += "\x1b[K"
		*ab += "\r\n"
	}
}

// ステータスバーを描画
func editorDrawStatusBar(ab *string) {
	// 色反転
	*ab += "\x1b[7m"

	// ファイル名
	fileName := ec.fileName
	if fileName == "" {
		fileName = "[No Name]"
	}
	// ファイルの編集状態
	dirtyMsg := ""
	if ec.dirty > 0 {
		dirtyMsg = "(modified)"
	}
	// 左側テキスト
	status := fmt.Sprintf("%.20s - %d lines %s", fileName, len(ec.row), dirtyMsg)
	stLen := min(len(status), ec.screenCols)
	*ab += status[:stLen]

	// 現在行/全行数
	rstatus := fmt.Sprintf("%d/%d", ec.cy+1, len(ec.row))
	rstLen := len(rstatus)
	// 右側テキストは表示する余裕がある場合にのみ表示する
	between := ec.screenCols - stLen - rstLen
	if between >= 0 {
		*ab += strings.Repeat(" ", between)
		*ab += rstatus
	} else {
		// 残りをスペースで埋める
		*ab += strings.Repeat(" ", ec.screenCols-stLen)
	}

	// 色反転解除
	*ab += "\x1b[m"
	*ab += "\r\n"
}

// メッセージバーを描画
func editorDrawMessageBar(ab *string) {
	// カーソル位置を復帰
	*ab += "\x1b[K"
	// メッセージを描画
	msg := ec.statusMsg
	msgLen := min(len(msg), ec.screenCols)
	if msgLen > 0 && (time.Since(ec.statusMsgTime) < 5*time.Second) { // メッセージがセットされてから5秒以内に限る
		*ab += msg[:msgLen]
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

	// テキスト行を描画
	editorDrawRows(&ab)
	// ステータスバーを描画
	editorDrawStatusBar(&ab)
	// メッセージバーを描画
	editorDrawMessageBar(&ab)

	// カーソルを指定位置に移動して表示
	ab += fmt.Sprintf("\x1b[%d;%dH", (ec.cy-ec.rowOff)+1, (ec.rx-ec.colOff)+1)
	ab += "\x1b[?25h"

	// テキストバッファの内容を出力
	syscall.Write(syscall.Stdout, []byte(ab))
}

// ステータスメッセージを表示
func editorSetStatusMessage(format string, a ...any) {
	ec.statusMsg = fmt.Sprintf(format, a...)
	ec.statusMsgTime = time.Now()
}

/*** input ***/

// カーソル移動
func editorMoveCursor(key rune) {
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
	rowLen := 0
	if ec.cy < len(ec.row) {
		row = ec.row[ec.cy].chars
		rowLen = len(row)
	}
	ec.cx = min(ec.cx, rowLen)
}

// 未保存で終了するためのキー押下残り回数
var quitTimes int = kiloQuitTimes

// キー入力を待ち、入力されたキーに対応する処理を行う
func editorProcessKeypress() (quit bool) {
	// キー入力
	c := editorReadKey()

	switch c {
	// Enter
	case '\r':
		// TODO:

	// Ctrl-Q
	case ctrlKey('q'):
		// 未保存の終了におけるキー押下回数が規定に満たない場合
		if ec.dirty > 0 && quitTimes > 0 {
			// メッセージを表示して残り回数をデクリメント
			editorSetStatusMessage("WARNING!!! File has unsaved changes. Press Ctrl-Q %d more times to quit.", quitTimes)
			quitTimes--
			return
		}
		// プログラム終了
		quit = true

	// Ctrl-S
	case ctrlKey('s'):
		// ファイル保存
		editorSave()

	// Home
	case homeKey:
		// カーソルを行頭へ移動
		ec.cx = 0

	// End
	case endKey:
		// カーソルを行末へ移動
		if ec.cy < len(ec.row) {
			ec.cx = len(ec.row[ec.cy].chars)
		}

	// BS, Ctrl-H, Del
	case backspace, ctrlKey('h'), delKey:
		// Delの場合はカーソルを1つ右に移動しておく
		if c == delKey {
			editorMoveCursor(arrowRight)
		}
		// カーソルの左側の文字を削除
		editorDelChar()

	// PageUp, PageDown
	case pageUp, pageDown:
		var arrow rune
		if c == pageUp {
			// 内部的に上矢印キーを発行
			arrow = arrowUp
			// カーソルを画面上端に移動
			ec.cy = ec.rowOff
		} else if c == pageDown {
			// 内部的に下矢印キーを発行
			arrow = arrowDown
			// カーソルを画面下端に移動
			ec.cy = ec.rowOff + ec.screenRows - 1
			ec.cy = min(ec.cy, len(ec.row)) // ファイル終端より先には移動しない
		}
		// 1画面分の行数だけ上下カーソル移動を発行することで1ページ分のスクロールを行う
		for i := 0; i < ec.screenRows; i++ {
			editorMoveCursor(arrow)
		}

	// 矢印キー
	case arrowUp, arrowDown, arrowLeft, arrowRight:
		// カーソルを上下左右に移動
		editorMoveCursor(c)

	// Ctrl-L, ESC
	case ctrlKey('l'), '\x1b':
		// 何もしない

	// その他のキー
	default:
		// カーソル位置に文字を挿入
		editorInsertChar(c)
	}

	// 未保存で終了するためのキー押下残り回数をリセット
	quitTimes = kiloQuitTimes

	return
}

/*** init ***/

// 初期化
func initEditor() {
	// カーソル位置初期化
	ec.cx = 0
	ec.cy = 0
	ec.rx = 0
	ec.rowOff = 0
	ec.colOff = 0
	ec.dirty = 0

	// ウィンドウサイズ取得
	rows, cols, err := getWindowsSize()
	if err != nil {
		panic(err)
	}
	ec.screenRows, ec.screenCols = rows, cols

	// ステータスバー、ステータスメッセージを確保
	ec.screenRows -= 2
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

	// ステータスメッセージを表示
	editorSetStatusMessage("HELP: Ctrl-S = save | Ctrl-Q = quit")

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
