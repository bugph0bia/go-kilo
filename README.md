# go-kilo

Go言語でテキストエディタ kilo を作る

## 概要

C言語でテキストエディタを作るチュートリアルである [Build Your Own Text Editor](https://viewsourcecode.org/snaptoken/kilo/index.html) を参考に、Go言語でテキストエディタを作ってみる。  

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

## [1. Setup](https://viewsourcecode.org/snaptoken/kilo/01.setup.html)

### [1-1. How to install a C compiler…](https://viewsourcecode.org/snaptoken/kilo/01.setup.html#how-to-install-a-c-compiler)

#### チュートリアル

- C言語開発環境のセットアップ

#### 実践

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

### [1-2. The main() function](https://viewsourcecode.org/snaptoken/kilo/01.setup.html#the-main-function)

#### チュートリアル

- `main()` を用意。終了コードとして 0 を返すだけ。
- ターミナル上で終了コードを確認する場合は `echo $?`。

#### 実践

前の手順で kilo.go 及び main() を作成済み。  
Goでは明示的に return を書かなければ 0 を返すので、終了コードは省略。  

### [1-3. Compiling with make](https://viewsourcecode.org/snaptoken/kilo/01.setup.html#compiling-with-make)

#### チュートリアル

- make コマンドでビルドする環境を整える。Makefile を準備する。

#### 実践

Go用のMakefileを追加。  

## [2. Entering raw mode](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html)

#### チュートリアル

- 標準入力から1文字読み込む処理を追加する。
- ターミナルはデフォルトで CANONICAL モード（COOKED モード）で起動している。
    - 入力はユーザーが Enter を押したときだけプログラムに送られる。
    - このモードはテキストエディタのようなプログラムには適していないためRAWモードにする必要がある。
- プログラムを終了させるには `Ctrl-D` でファイル（標準入力）の終端を伝えるか、`Ctrl-C` で強制終了する。

#### 実践

標準入力から1文字取得するコードを追加。  
Ctrl+DやCtrl+Cで終了する点はチュートリアルと同じ。  

VSCodeのデバッグ中に標準入力をうまく扱えなかったため、launch.jsonに `"console": "integratedTerminal"` を追加して対処。  

## [2-1. Press `q` to quit?](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#press-q-to-quit)

#### チュートリアル

- `q`を入力することでプログラムが終了するようにする。

#### 実践

`q` が入力された場合も狩猟するようにコードを修正。  

## [2-2. Turn off echoing](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-echoing)

#### チュートリアル

- ターミナルの属性を変更する処理を追加する。
- `ECHO`をOFFにすることで入力した文字が画面に表示（エコー）されるのを抑制する。
    - `sudo` コマンドなどでパスワード入力時に使われる状態。
    - エコーOFFでプログラムが終了するとターミナルはエコーOFFのままとなる。`Ctrl-C`の後に`reset`コマンドを実行することで元の状態に戻るが、それでダメならターミナルを再起動すれば元に戻る。
- ターミナルの属性には、c_iflag (入力フラグ)、c_oflag (出力フラグ)、c_cflag (制御フラグ)、c_lflag (その他フラグ) がある。

#### 実践

Go言語でターミナルの属性を取得/設定する方法を探したところ、最終的に `term.MakeRaw()` を利用すればよいであろうことがわかった。  
そのため、`enableRawMode()` は自前で実装する必要がなくなった。  
`MakeRaw()` は厳密には、ここで行いたかった`ECHO`フラグのOFF以外にも諸々フラグの設定を行ってくれる関数となっている。  

調査の過程で参考にした情報：  

- https://stackoverflow.com/questions/69693105/golang-unix-tcgets-equivalent-on-mac
    - `unix.IoctlGetTermios()` と `unix.IoctlSetTermios()` を使用してターミナルの属性を取得するコード。
    - 渡すフラグは `unix.TCSAFLUSH` や `unix.TCGETA` `unix.TCSETA` ではなく `unix.TCGETS` `unix.TCSETS` でないとうまくいかない模様。`MakeRaw()`もそのように実装されている。Linux系のOSの違いによるものか？
- https://qiita.com/x-color/items/f2b6b0852c1a7484ffff
    - `import "golang.org/x/crypto/ssh/terminal"` で利用できる `terminal.ReadPassword()` の内部コード。
- https://github.com/mattn/go-tty
    - 最終的に利用することになった term モジュールと同じようなコードが実装されている。

CではビットOFFするときに `flag &= ~BITS` と書くが、Goでは `flag &^= BITS` と書く。  

## [2-3. Disable raw mode at exit](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#disable-raw-mode-at-exit)

#### チュートリアル

- プログラムの終了時にRAWモードを無効にする。

#### 実践

`MakeRaw()` の戻り値に変更前のターミナル属性が返ってくるので、これを保存しておいてプログラム終了時に実行する。  
Go言語では `atexit()` の代わりに `defer` 文を使用可能。  
前回と同様、`disableRawMode()` は実装する必要なし。  

## [2-4. Turn off canonical mode](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-canonical-mode)

#### チュートリアル

- `ICANON`をOFFにすることでCANONICALモードをOFFにする。入力を行単位ではなくバイト単位で読むことになる。
    - `I` から始まるが適用先は `c_lflag` であることに注意。

#### 実践

`MakeRaw()` の中で、`ICANON`フラグのOFFも行われているため、ここでは何もすることはない。  

## [2-5. Display keypresses](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#display-keypresses)

#### チュートリアル

- 入力されたキーを画面出力する処理を追加する。
    - 矢印キーや `Escape` `Page Up` などのエスケープシーケンスは27から始まる3から4バイトが出力される。
    - `BS` は127が出力される。
    - `Delete` は4バイトのエスケープシーケンス。
    - `Enter` は10が出力される。改行文字。"\n" とも呼ばれる。
    - `Ctrl-A`は1、`Ctrl-B`は2、…のように数値に対応する。ただし、ターミナルにとって特別な意味のあるキーは専用の動作をする。
        - `Ctrl-C`はプログラムの強制終了。
        - `Ctrl-S`は画面出力の停止。再開は`Ctrl-Q`。
        - `Ctrl-Z`はプログラムを一時停止してバックグラウンドに移す。`fg`コマンドでフォアグラウンドに戻せる。

#### 実践

`iscntrl()` に対応するGoの関数は `unicode.IsControl()` 。  
`printf()` は、ほぼそのまま使える `fmt.Printf()` がある。`%c` `%d` などの書式指定子もほぼそのまま使える。Goにはより汎用的に使える `%v` があるが、まずはC言語と同じものを使うこととする。   

1文字分を読み込んで出力する部分は、Rune型に変換するコードとする。  

`MakeRaw()` でこの先に行う予定の実装を先行して対応しているため、`Ctrl+Z` などの動きはチュートリアル通りに試すことはできない。  

## [2-6. Turn off `Ctrl-C` and `Ctrl-Z` signals](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-ctrl-c-and-ctrl-z-signals)

#### チュートリアル

- `SIGINT`をOFFにすることで`Ctrl-C`を無効化する。
- `SIGTSTP`をOFFにすることで`Ctrl-Z`を無効化する。macOSでは`Ctrl-Y`も無効化される。

#### 実践

`MakeRaw()` の中で、`ISIG`フラグのOFFも行われているため、ここでは何もすることはない。  

## [2-7. Disable `Ctrl-S` and `Ctrl-Q`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#disable-ctrl-s-and-ctrl-q)

#### チュートリアル

- `IXON`をOFFにすることで`Ctrl-S` `Ctrl-Q`を無効化する。

#### 実践

`MakeRaw()` の中で、`IXON`フラグのOFFも行われているため、ここでは何もすることはない。  

## [2-8. Disable `Ctrl-V`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#disable-ctrl-v)

#### チュートリアル

- `IEXTEN`をOFFにすることで`Ctrl-V`を無効化する。macOSでは`Ctrl-O`も修正される。

#### 実践

`MakeRaw()` の中で、`IEXTEN`フラグのOFFも行われているため、ここでは何もすることはない。  

## [2-9. Fix `Ctrl-M`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#fix-ctrl-m)

#### チュートリアル

- `Ctrl-A`から`Ctrl-Z`まで調べると1から26の数値に対応するのだが、`Ctrl-M`だけは13ではなく10になる。他に`Ctrl-J`と`Enter`も10となる。
    - ターミナルにはCR(13, '\r')をNL(10, '\n')に変換する機能が備わっている。
- `ICRNL`をOFFにすることでこの機能を無効化する。これで`Ctrl-M`と`Enter`は変換されず13となる。

#### 実践

`MakeRaw()` の中で、`ICRNL`フラグのOFFも行われているため、ここでは何もすることはない。  

## [2-10. Turn off all output processing](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-all-output-processing)

#### チュートリアル

- ターミナルは出力側でも同じような変換を行っており、NL(10, '\n')をCR+NL(13 10, '\r\n')に変換する。
    - ターミナル上で改行するためには、CRでカーソルを行の先頭に戻して、NLで一行下に移動（必要に応じてスクロール）させる必要があるため。これは、タイプライターやテレタイプの時代に生まれたもの。
- `OPOST`をOFFにすることでこの機能を無効化する。
    - 出力フラグでデフォルトでONなのはこの１つだけと思われる。
    - OFFにした後は、Print時に "\n" だけ指定してもカーソルが左端に戻らなくなるので、"\r\n" を指定する必要がある。

#### 実践

`MakeRaw()` の中で、`ICRNL`フラグのOFFも行われているため、ここでもフラグ制御は不要。  

文字を出力するときの末尾の改行コードを `\n` から `\r\n` に変更することで、これまでは出力するたびにカーソルが中途半端な位置にあったが、毎回ターミナルの左端に戻るようになることを確認できる。  
`\r` (CR: Carriage Return) の本来の役割を体感できる。  

## [2-11. Miscellaneous flags](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#miscellaneous-flags)

#### チュートリアル

- `BRKINT`をOFFにすることで、ブレーク条件が発生しても`Ctrl-C`が送られなくなる。
- `INPCK`をOFFにすることで、パリティチェックを無効化する。
    - 最近のターミナルには適用されない模様。
- `ISTRIP`をOFFにすることで、各入力バイトの最上位ビットを除去する動作を抑制する。
- `CS8`（複数ビットをもつビットマスク）をONにする。
    - 既にONになっているもの？
    - ONにする効果は？

#### 実践

`MakeRaw()` の中で、`BRKINT`, `ISTRIP`フラグのOFFと`CS8`ビットマスクのONも行われているため、ここでは何もすることはない。  
`INPCK`フラグのOFFについては行われていないが、チュートリアルの説明にもあるように最近のターミナルには適用されないフラグなので省略されているのではないかと推測する。  

ここまでで各種フラグ操作について見てきたが、`MakeRaw()` ではチュートリアルのコードで行っていないフラグ制御がまだ行われている。  
この違いについては一旦無視し、チュートリアルと動作の差異が出てきたら適宜確認することにして次に進む。  

## [2-12. A timeout for `read()`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#miscellaneous-flags)

#### チュートリアル

- 標準入力から文字を読み込む際のタイムアウトを設定する。
- `VMIN` を0にすることで、入力があるとすぐに読み込み処理から制御が戻るようになる。
- `VTIME` を1にすることで、1/10秒のタイムアウトを設定する。
- 上記を設定後にプログラムを実行すると、何も入力がないときに0が出力され続ける動作を確認できる。1/10秒より短い間隔で入力があるとその度に制御が返るため、1/10秒に1回しか読み取れない訳ではない。

#### 実践

`MakeRaw()` の中で `VMIN=1`, `VTIME=0` と設定されるが、これはチュートリアルと異なる設定である。  
通常、ターミナルをRAWモードにする場合は理にかなっているようだが、チュートリアルに合わせるために `VMIN=0`, `VTIME=1` と設定するコードを追加する。  
これによってRAWモード設定のコードが長くなるため、チュートリアルと同じ `enableRawMode()` という関数にまとめ、対称性のために `disableRawMode()` も作る。  

`VMIN=0`, `VTIME=1` に設定したことで、チュートリアルの説明の通りの下記の動作を確認することができるようになる。  

- 何もキーを押さないと `0` が出力され続ける（1/10秒周期）。
- キーを押すと対応する文字が出力される。1/10秒より早く押しても反応する。

VMIN と VTIME に関しては、下記を参照。  
http://www.unixwiz.net/techtips/termios-vmin-vtime.html

## [2-13. Error handling](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#error-handling)

#### チュートリアル

- 前項でターミナルを完全にRAWモードにすることができた。
- エラーハンドリングをコードに追加する。エラー発生時、エラーメッセージを表示して、終了コードを非0にしてプログラムを終了する。
- Cygwinではタイムアウト時にread()から`EAGAIN`が返されるため、これをエラーとして扱わないようにする。

#### 実践

チュートリアルの `die()` 関数は、Goでは `panic()` で代用できる。  
ここまでのコードでもすでに使用してきた。  
メッセージを出して終了コード1で中断するという意味では `log.Fatal()` の方が相応しいかもしれないが、`defer` を適切に処理する `panic()` を採用することにする。  
`panic()` だと終了コードは1になる保証はないようだが、非0なので問題ないと考える。  

`syscall.EAGAIN` をエラーにしないようにするために、標準入力から読み出す関数も `syscall.Read()` に変更した。  

## [2-14. Sections](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#sections)

#### チュートリアル

- ソースコードにセクションコメントを追加する。

#### 実践

チュートリアルではソースコードをセクションコメントで分割しているが、それには倣わないこととする。  

## `MakeRaw()` を使用せずチュートリアルに合わせる方針に変更

ここまでターミナルをRAWモードにする手段を `MakeRaw()` 呼び出しに頼ってきたが、結果的に 2-12 でチュートリアルの内容と乖離してしまった。  
一旦は `MakeRaw()` を使う方針で進めようと考えていたが、チュートリアルのコードをトレースすることも可能なので、方針変更して、チュートリアル通りに進めることとした。  
チュートリアルと `MakeRaw()` の差異を下記にまとめる。  

`MakeRaw()` でのみOFFにしている機能：  

| フラグ  | 機能   | 説明                                                                                                        |
| ------- | ------ | ----------------------------------------------------------------------------------------------------------- |
| c_iflag | IGNBRK | 入力中の BREAK 信号を無視する。                                                                             |
| c_iflag | PARMRK | IGNPARが設定されていない場合、パリティエラーあるいはフレームエラーの発生した文字の前に \377 \0 を付加する。 |
| c_iflag | INLCR  | 入力の NL (New Line: 改行文字) を CR (Carriage Return: 復帰文字) に 置き換える。                            |
| c_iflag | IGNCR  | 入力の CR を無視する。                                                                                      |
| c_oflag | -      | 差異無し                                                                                                    |
| c_cflag | CSIZE  | 文字サイズを設定する。値は CS5, CS6, CS7, CS8 である。                                                      |
| c_cflag | PARENB | 出力にパリティを付加し、入力のパリティチェックを行う。                                                      |
| c_lflag | ECHONL | ICANON も同時に設定された場合、ECHOが設定されていなくてもNL文字をエコーする。                               |

チュートリアルでのみOFFにしている機能：  

| フラグ  | 機能  | 説明                                                                                                                                         |
| ------- | ----- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| c_iflag | INPCK | 入力のパリティチェックを有効にする。<br/>チュートリアルによると現在のターミナルでは意味がないようで、`MakeRaw()` でOFFにしないのはそのため？ |

`MakeRaw()` とチュートリアルで設定値に差異がある変数：  

| 変数 | 機能  | 説明                                                                                                                                                                                         |
| ---- | ----- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| c_cc | VMIN  | 非カノニカル読み込み時の最小文字数 (MIN)。<br/> `MakeRaw()` では 1 を設定しており1文字分の入力があるまでブロックする。<br/>チュートリアルでは 0 を設定しており入力がなくてもブロックしない。 |
| c_cc | VTIME | 非カノニカル読み込み時のタイムアウト時間 (1/10秒単位)。<br/> `MakeRaw()` では 0 を設定しておりタイムアウトなし。<br/>チュートリアルでは 1 を設定しており1/10秒でタイムアウト。               |

参考：  

- https://ja.manpages.org/termios/3
- http://www.unixwiz.net/techtips/termios-vmin-vtime.html

## [3. Raw input and output](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html)

### [3-1. Press `Ctrl-Q` to quit](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#press-ctrl-q-to-quit)

#### チュートリアル

- `Ctrl-Q`でプログラムが終了するように変更する。
- ASCIIコードの特性：
    - 英字の上位3ビットを落とすとCtrl+英字キーを押したときのコードに対応する。
    - 英字の第5ビットのOFF/ONで大文字/小文字の変換が可能である。

#### 実践

Goにはマクロ関数はないので、`CTRL_KEY()`は通常の関数で実装する。  

C言語でASCIIコード1バイト分を表すのはchar型だが、Goではrune型としている。  
rune型はutf8のコードポイントを表すことになるが、ASCIIコードの範囲内では同じ値になるので支障はないはず。  

### [3-2. Refactor keyboard input](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#refactor-keyboard-input)

#### チュートリアル

- キー入力を待つ関数と、入力されたキーに応じた動作を定義する関数を用意する。
- 入力されたキーを画面表示することをやめる。
- その結果、main()関数を簡素化する。

#### 実践

ほぼチュートリアル通りにコードを実装。  

ただし、Ctrl-Q でプログラム終了するときに `os.Exit(0)` してしまうと、defer してある `disableRawMode()` が呼び出されない。  
そのため、`main()` を経由して終了するコードとした。  

また、CとGoの言語の違いがあるためセクションコメントを入れてこなかったが、この時点でかなり似たコードにできているため、セクションコメントを入れることとした。  
Goにはdefineがない、includesではなくimports、といった違いはある。  
dataセクションも設けることとし、パッケージレベルの変数（グローバル変数）の利用についてもチュートリアルに倣うこととした。  
