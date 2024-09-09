# go-kilo

Go言語でテキストエディタ kilo を作る

## 概要

C言語でテキストエディタを作るチュートリアルの [Build Your Own Text Editor](https://viewsourcecode.org/snaptoken/kilo/index.html) を参考に、Go言語でテキストエディタを作ってみる。  

## 目的

- Go言語の習熟
- C言語の復習
- テキストエディタの作りを学ぶ

（C言語は久しぶり、Go言語は勉強中）  

## 方針

- 極力チュートリアルに沿って進める。
    - C言語からGo言語に単純に置き換えられないところが出てきたら都度考える。
    - 最低でもチュートリアルのステップごとにコミットする。
- この REDAME に開発記録をメモしながら進める。

## 1. [Setup](https://viewsourcecode.org/snaptoken/kilo/01.setup.html)

### 1-1. [How to install a C compiler…](https://viewsourcecode.org/snaptoken/kilo/01.setup.html#how-to-install-a-c-compiler)

C言語ではなく、Go言語の環境をセットアップする。  

チュートリアルはWindowsに対応していないようなので、Linuxで進める。  
WSL2 でインストール済みのUbuntu20.04があったのでこれを利用することにする。  

UbuntuにはGoのランタイムはインストール済みであったが、バージョンが古いので入れ直し。  
Goの最新バージョンは v1.23.1 であった。  

```sh
# 現在のバージョンを確認
$ go version
go version go1.19 linux/amd64

# 公式サイトから最新のランタイムをダウンロード
$ wget https://dl.google.com/go/go1.23.1.linux-amd64.tar.gz
# ランタイムをアップグレード
$ sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.1.linux-amd64.tar.gz

$ go version
go version go1.23.1 linux/amd64
```

今回の開発用のモジュールのディレクトリを作成する。モジュール名は go-kilo 。  
※モジュールは Windows 側（/mnt/c/... 配下）には置かないこと。Linux側にソースファイルが置かれていないとdlvがファイルを認識できず、デバッグが行えない。  

モジュールを初期化する。  

```sh
# モジュールフォルダへ移動
$ cd .../go-kilo
# Goモジュールとして初期化
$ go mod init go-kilo
```

エディタはWindows側にインストールしたVSCodeを使うこととする。  

1. VSCode に拡張機能 Remote Developement を入れる。
2. WSL2でモジュールフォルダに移動して `code .` を実行。  
3. Windows側でVSCodeが起動して、ステータスバーの左端に「WSL: Ubuntu-20.04」と表示されればOK。
4. さらにGoの拡張機能をインストールして、表示される指示に従ってGo関連のツール（dlvなど）をインストールして準備完了。
5. Hello Worlodのソースを書いて、VSCodeからデバッグができること（ブレークポイントに止まるなど）を確認。

### 1-2. [The main() function](https://viewsourcecode.org/snaptoken/kilo/01.setup.html#the-main-function)

前の手順で kilo.go 及び main() を作成済み。  
Goでは明示的に return を書かなければ 0 を返すので、終了コードは省略。  

### 1-3. [Compiling with make](https://viewsourcecode.org/snaptoken/kilo/01.setup.html#compiling-with-make)

Go用のMakefileを追加。  

## 2. [Entering raw mode](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html)

標準入力から1文字取得するコードを追加。  
Ctrl+DやCtrl+Cで終了する点はチュートリアルと同じ。  

VSCodeのデバッグ中に標準入力をうまく扱えなかったため、launch.jsonに `"console": "integratedTerminal"` を追加して対処。  

## 2-1. [Press `q` to quit?](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#press-q-to-quit)

`q` が入力された場合も狩猟するようにコードを修正。  

## 2-2. [Turn off echoing](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-echoing)

Go言語でターミナルの属性を取得/設定する方法を探したところ、最終的に `term.MakeRaw()` を利用すればよいであろうことがわかった。  
そのため、`enableRawMode()` は自前で実装する必要がなくなった。  
`MakeRaw()` は厳密には、ここで行いたかった`ECHO`フラグのOFF以外にも諸々フラグの設定を行ってくれる関数となっている。  

調査の過程で参考にした情報：  

- https://stackoverflow.com/questions/69693105/golang-unix-tcgets-equivalent-on-mac
    - `unix.IoctlGetTermios()` と `unix.IoctlSetTermios()` を使用してターミナルの属性を取得するコード。
    - 渡すフラグは `unix.TCGETA` `unix.TCSETA` ではなく `unix.TCGETS` `unix.TCSETS` でないとうまくいかない？Linux系のOSの違いによるもの？
- https://qiita.com/x-color/items/f2b6b0852c1a7484ffff
    - `import "golang.org/x/crypto/ssh/terminal"` で利用できる `terminal.ReadPassword()` の内部コード。
- https://github.com/mattn/go-tty
    - 最終的に利用することになった term モジュールと同じようなコードが実装されている。

## 2-3. [Disable raw mode at exit](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#disable-raw-mode-at-exit)

`MakeRaw()` の戻り値に変更前のターミナル属性が返ってくるので、これを保存しておいてプログラム終了時に実行する。  
Go言語では `atexit()` の代わりに `defer` 文を使用可能。  
前回と同様、`disableRawMode()` は実装する必要なし。  

## 2-4. [Turn off canonical mode](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-canonical-mode)

`MakeRaw()` の中で、`ICANON`フラグのOFFも行われているため、ここでは何もすることはない。  

## 2-5. [Display keypresses](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#display-keypresses)

`iscntrl()` に対応するGoの関数は `unicode.IsControl()` 。  
`printf()` は、ほぼそのまま使える `fmt.Printf()` がある。`%c` `%d` などの書式指定子もほぼそのまま使える。Goにはより汎用的に使える `%v` があるが、まずはC言語と同じものを使うこととする。   

1文字分を読み込んで出力する部分は、Rune型に変換するコードとする。  

`MakeRaw()` でこの先に行う予定の実装を先行して対応しているため、`Ctrl+Z` などの動きはチュートリアル通りに試すことはできない。  

## 2-6. [Turn off `Ctrl-C` and `Ctrl-Z` signals](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-ctrl-c-and-ctrl-z-signals)

`MakeRaw()` の中で、`ISIG`フラグのOFFも行われているため、ここでは何もすることはない。  

## 2-7. [Disable `Ctrl-S` and `Ctrl-Q`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#disable-ctrl-s-and-ctrl-q)

`MakeRaw()` の中で、`IXON`フラグのOFFも行われているため、ここでは何もすることはない。  

## 2-8. [Disable `Ctrl-V`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#disable-ctrl-v)

`MakeRaw()` の中で、`IEXTEN`フラグのOFFも行われているため、ここでは何もすることはない。  

## 2-9. [Fix `Ctrl-M`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#fix-ctrl-m)

`MakeRaw()` の中で、`ICRNL`フラグのOFFも行われているため、ここでは何もすることはない。  

## 2-10. [Turn off all output processing](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-all-output-processing)

`MakeRaw()` の中で、`ICRNL`フラグのOFFも行われているため、ここでもフラグ制御は不要。  

文字を出力するときの末尾の改行コードを `\n` から `\r\n` に変更することで、これまでは出力するたびにカーソルが中途半端な位置にあったが、毎回ターミナルの左端に戻るようになることを確認できる。  
`\r` (CR: Carriage Return) の本来の役割を体感できる。  

## 2-11. [Miscellaneous flags](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#miscellaneous-flags)

`MakeRaw()` の中で、`BRKINT`, `ISTRIP`フラグのOFFと`CS8`ビットマスクのONも行われているため、ここでは何もすることはない。  
`INPCK`フラグのOFFについては行われていないが、チュートリアルの説明にもあるように最近のターミナルには適用されないフラグなので省略されているのではないかと推測する。  

ここまでで各種フラグ操作について見てきたが、`MakeRaw()` ではチュートリアルのコードで行っていないフラグ制御がまだ行われている。  
この違いについては一旦無視し、チュートリアルと動作の差異が出てきたら適宜確認することにして次に進む。  

## 2-12. [A timeout for `read()`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#miscellaneous-flags)

`MakeRaw()` の中で `VMIN=1`, `VTIME=0` と設定されるが、これはチュートリアルと異なる設定である。  
通常、ターミナルをRAWモードにする場合は理にかなっているようだが、チュートリアルに合わせるために `VMIN=0`, `VTIME=1` と設定するコードを追加する。  
これによってRAWモード設定のコードが長くなるため、チュートリアルと同じ `enableRawMode()` という関数にまとめ、対称性のために `disableRawMode()` も作る。  

`VMIN=0`, `VTIME=1` に設定したことで、チュートリアルの説明の通りの下記の動作を確認することができるようになる。  

- 何もキーを押さないと `0` が出力され続ける（1/10秒周期）。
- キーを押すと対応する文字が出力される。1/10秒より早く押しても反応する。

VMIN と VTIME に関しては、下記を参照。  
http://www.unixwiz.net/techtips/termios-vmin-vtime.html

## 2-13. [Error handling](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#error-handling)

チュートリアルの `die()` 関数は、Goでは `panic()` で代用できる。  
ここまでのコードでもすでに使用してきた。  
メッセージを出して終了コード1で中断するという意味では `log.Fatal()` の方が相応しいかもしれないが、`defer` を適切に処理する `panic()` を採用することにする。  
`panic()` だと終了コードは1になる保証はないようだが、非0なので問題ないと考える。  

`syscall.EAGAIN` をエラーにしないようにするために、標準入力から読み出す関数も `syscall.Read()` に変更した。  

## 2-14. [Sections](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#sections)

チュートリアルではソースコードをセクションコメントで分割しているが、それには倣わないこととする。  
