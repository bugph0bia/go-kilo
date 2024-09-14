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
Goでは `atexit()` の代わりに `defer` 文を使用可能。  
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

### [3-3. Clear the screen](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#clear-the-screen)

#### チュートリアル

- エスケープシーケンスは常にESC文字(27, 0x1B)で始まり、'['が続く。
    - エスケープシーケンスを用いることで、テキストに色を付けたり、カーソルを移動させたり、画面の一部を消去したりといった、様々なテキスト整形作業を指示できる。
- ターミナルにエスケープシーケンス "\x1b[2J" を書き込む。
    - スクリーンを消去する J コマンドを使用する。
    - コマンドは引数を取る。この場合の引数は"2"であり、画面全体を消去する指示となる。
    - "\x1b[1J" はカーソル位置までを消去する。
    - "\x1b[0J" はカーソル位置から画面端までを消去する。"0"はデフォルト引数であり、"\x1b[J"でも同じ意味となる。
- このテキストエディタは、最近のターミナルで広くサポートされるV100エスケープシーケンスを主に使用する。
    - [V100 User Guide](http://vt100.net/docs/vt100-ug/chapter3.html)
    - ターミナルのサポート範囲を最大にする場合は、ncursesライブラリを使うと良い。terminfoデータベースを使って、そのターミナルの機能と使うべきエスケープシーケンスを調べることができる。

#### 実践

チュートリアル通りのコードを追加。  

### [3-4. Reposition the cursor](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#reposition-the-cursor)

#### チュートリアル

- H コマンドを使ってカーソルを配置する。
    - 実際には配置する位置を表す行番号と列番号の2つの引数を取るが、デフォルトではどちらも1になるため、"\x1b[H"で左上にカーソルを置ける。行番号、列番号、ともに1始まりであり、0始まりではない。

#### 実践

チュートリアル通りのコードを追加。  

### [3-5. Clear the screen on exit](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#clear-the-screen-on-exit)

#### チュートリアル

- プログラム終了時にスクリーンをクリアする。
- エラー発生時にスクリーンクリア→エラーメッセージ表示の順に処理されるようにする。

#### 実践

チュートリアルでは、エラー用の関数 `die()` の中にスクリーン消去処理を書いているが、本ツールでは `panic()` を利用している。  
また、`atexit()` の中にスクリーン消去処理を書くとエラーメッセージまで消えてしまうとあるが、Goの`panic()`はdeferを適切に処理してからエラーメッセージを表示してくれる。  
このことから、本ツールではdeferを用いて前項で作成した `editorRefreshScreen()` を呼び出しすことでスクリーン消去することとした。  


### [3-6. Tildes](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#tildes)

#### チュートリアル

- Vim のように画面左端にチルダ(~)を並べる。
    - ターミナルのサイズはまだわからないが、一旦24行分とする。

#### 実践

チュートリアルどおりにコードを実装。  
プログラム終了時のスクリーン消去処理は、`editorRefreshScreen()`を使えなくなってしまったので個別の関数呼び出しとする。  

### [3-7. Global state](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#global-state)

#### チュートリアル

- エディタの状態を保持するグローバル変数を準備する。

#### 実践

チュートリアルではエディタの状態を保持するグローバル変数は `E` としているが、Goでは大文字の変数名は外部公開されてしまうため避けたい。そのため本ツールでは `ec` という変数名にする。  

### [3-8. Window size, the easy way](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#window-size-the-easy-way)


#### チュートリアル

- ウィンドウサイズを取得する。
    - ほとんどのシステムでは、システムコール `ioctl()` を `TIOCGWINSZ` リクエストで呼び出すことで、ターミナルのサイズを取得することができる。
    - `TIOCGWINSZ` は、**T**erminal **IOC**tl (**I**nput/**O**utput **C**on**t**ro**l**) **G**et **WIN**dow **S**i**Z**e の略。

#### 実践

Goでは、システムコール `ioctl()` は `unix.IoctlXXX()` 系のラッパー関数から利用できる。  
値を取得するための関数は `unix.IoctlGetXXX()` として各種用意されており、`XXX` 部分は戻り値の型ごとに用意されている模様。  
ウィンドウサイズを取得する関数は `unix.IoctrlGetWinsize()` であり、`Winsize` 型の戻り値を返す。第2引数には `unix.TIOCGWINSZ` を渡す必要があり、これはCの場合と同様である。`Winsize` 型の戻り値を取得するケースは `unix.TIOCGWINSZ` リクエストしかないようなので第2引数を省略できても良いように思えるが、一貫性のために引数を渡すことになっているのだと推測される。  

取得したサイズはエディタの状態を保持する構造体のメンバとするが、チュートリアルの `screenrows` `screencols` は気に入らなかったので、`screenRows` `screenCols` とした。  

### [3-9. Window size, the hard way](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#window-size-the-hard-way)

#### チュートリアル

- `ioctl()` では全ての環境でサイズを取得できるわけではないため、サイズを取得するための予備的な方法として、カーソルをスクリーンの右下に移動してその位置を取得する処理を用意する。
    - カーソルを右下に移動する単純な方法は用意されていないため、エスケープシーケンスを使用する。C コマンドに引数999を渡し、カーソルを右方向へ限界まで移動する。同様に、B コマンドに引数999を渡し、カーソルを下方向へ限界まで移動する。
    - H コマンドで "\x1b[999;999H" のようにカーソル移動しない理由は、H コマンドでカーソルを画面外に移動したときの挙動がドキュメントに明記されていないため。
- 次に、カーソル位置を取得する。
    - N コマンドを利用するとターミナルにステータス情報を問い合わせることができる。6という引数を与えると、標準入力からカーソル位置を取得できる。[参考](http://vt100.net/docs/vt100-ug/chapter3.html#DSR)
    - 標準入力に返されるのはエスケープシーケンスである。"\x1b[24;80R" など。[参考](http://vt100.net/docs/vt100-ug/chapter3.html#CPR)
- 取得したカーソル位置の結果を含む文字列から、行と列を取り出す。

#### 実践

カーソルを画面右下に移動するコードを追加して、その動作を確認する。動作確認のために、強制的に今回追加したコードを利用する条件を埋めておく。  
次に、カーソル位置を取得した結果を画面に出力して、その結果を確認する。  
さらに、カーソル位置を取得した結果を画面に文字列として出力する。このとき、先頭のESCと末尾の'R'は除去する。  
カーソル位置を含む文字列から行と列の数値を取り出して、ウィンドウサイズとして利用する。  
コードが動くことを確認できたので、通常の `unix.IoctrlGetWinsize()` で取得したサイズを使用するコードに戻す。  

### [3-10. The last line](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#the-last-line)

#### チュートリアル

- 最終行のチルダの後の改行コードを出力しないようにする。

#### 実践

チュートリアル通りのコードを追加。  

### [3-11. Append buffer](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#append-buffer)

#### チュートリアル

- `write()` を細かく呼び出しているため、チラツキの原因となっている。出力する文字列をバッファリングして一度に書き出すように変更する。
- Cには動的な文字列型がないため、そのための仕組みを自分で実装する必要がある。
    - Cにはクラスもないので、構造体とそれを操作するコンストラクタ、デストラクタを関数で実装する。

#### 実践

Goにはstring型があるのでほぼコードを書かずに完了。  

### [3-12. Hide the cursor when repainting](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#hide-the-cursor-when-repainting)

#### チュートリアル

- チラツキの要因がもう一つあり、スクリーンをリフレッシュするときに一瞬カーソルが画面中央に表示されてしまうことがある。これを解消するために、リフレッシュ前にカーソルを非表示にし、リフレッシュ後に再表示する。
    - l コマンドでモードのリセットを行う。
    - h コマンドでモードのセットを行う。
    - それぞれ引数に "?25" を渡すことでカーソルの非表示/表示を行うことになるが、ターミナルによってはこの引数に対応していないことがある。その場合であっても単に無視されるだけで問題にはならない。

#### 実践

チュートリアル通りのコードを追加。  
手元のターミナルではもともとチラツキを確認することはなかった。  

### [3-13. Clear lines one at a time](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#clear-lines-one-at-a-time)

#### チュートリアル

- リフレッシュの際に画面全体を消去するのではなく、各行を消去することで最適化する。
    - J コマンドの代わりに K コマンドを使用する。
    - K コマンドは現在行の一部を消去する。引数を2にすると行全体、1にするとカーソルの左側、0にするとカーソルの右側を消去する。デフォルト引数は0であり、今回はこれを使用する。

#### 実践

チュートリアル通りのコードを追加。  

### [3-14. Welcome message](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#welcome-message)

#### チュートリアル

- エディタの名前とバージョンを表示する。
    - スクリーン幅が小さい場合は文字列を切り詰めて表示する。
- 次に、そのメッセージを左右中央に配置する。

#### 実践

Goではstring型やスライスを使用できるので簡単に実装できる。  
左右中央に表示する際も、`min()` や `strings.Repeat()` を使用することでコードが簡潔になる。  
Goは定数を利用できるため、バージョン文字列は定数とする。チュートリアルにはないが const というセクションコメントを追加しておく。  

### [3-15. Move the cursor](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#move-the-cursor)

#### チュートリアル

- まず、エディタ状態としてカーソル位置 x, y を追加する。
- 初期値は 0, 0 とする。C言語はインデックスが0に始まるため、極力0始まりのインデックスを使用する。
    - ターミナルの行列位置を指定するときは+1して1始まりの数値に変換する。
- 次に、wasdキーを使用してカーソルを移動できるようにする。
    - 一旦、wasdキーを上左下右のように矢印キーに見立てている。

#### 実践

チュートリアル通りのコードを追加。  
一旦、カーソル位置を 10, 10 に設定して正しく動作することを確認する。確認後、0, 0 に戻す。  

wasdキーの入力時の条件分岐を追加する。チュートリアルではswitch文のフォールスルーを用いて4文字分の処理を書いているが、Goでは1つのcase節に複数条件を並べることができるのでより簡潔に書くことができる（`fallthrough`キーワードを用いてフォールスルーさせることも可能）。  

### [3-16. Arrow keys](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#arrow-keys)

#### チュートリアル

- 矢印キーを入力するとエスケープシーケンスとなり "\x1b[" に "A" ～ "D" が続く3バイトのコードとなる。これを wasd キーに変換するコードを追加する。
    - 例えば、右矢印キーは "\x1b[C" であり、ESC, '[', 'C' が連続するのと同じ意味である。
    - ESC は `Ctrl-[` でも入力することが可能。
- 最終的に wasd キーのマッピングは解除して、矢印キーのみでカーソル移動できるようにする。

#### 実践

チュートリアル通りのコードを追加。  

GoにはEnumはないので、constとiotaを使用する。  
文字コードを表すのにrune型を使ってきたが、ここで文字コードではない 1000～1004 の数値を保持することになったため、int型に変更する。runeはUnicodeコードポイントを保持する型であり、1000～1004も別の文字を表すコード二対応してしまうため。  

`Ctrl-[` が ESC に対応するという点は、3-1章に出てきた、上位3ビットを落とすとCtrl+キーを押したときのコードに対応する事実が、英字だけでなく記号にも適用されることを意味している。  

### [3-17. Prevent moving the cursor off screen](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#prevent-moving-the-cursor-off-screen)

#### チュートリアル

- カーソルが画面外に行かないように境界値チェックを設ける。

#### 実践

チュートリアル通りのコードを追加。  

### [3-18. The `Page Up` and `Page Down` keys](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#the-page-up-and-page-down-keys)

#### チュートリアル

- `Page UP` `Page Down` キーに対応する。
- `Page UP` は "\x1b[5~"、`Page Down` は "\x1b[6~" が送られる。
- まずは、それぞれ画面の一番上と一番下にカーソル移動するコードを実装する。
    - スクリーンの行数だけ上矢印または下矢印のコードを内部的に発行する。

#### 実践

GoだとCに比べて少しコードが長くなる。三項演算子が使えないのは残念。  

### [3-18. The `Home` and `End` keys)[https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#the-home-and-end-keys]

#### チュートリアル

- `Home` `End` キーに対応する。ターミナルエミュレータによって送られるエスケープシーケンスにバリエーションがある。これらすべてに対応する。
- `Home` は "\x1b[1~" "\x1b[7~" "\x1b[H" "\x1bOH" のいずれか。
- `End` は "\x1b[4~" "\x1b[8~" "\x1b[F" "\x1bOF" のいずれか。
- それぞれ、現在行の左端、右端二移動するコードとする。

#### 実践

チュートリアル通りのコードを追加。  

### [3-19. The `Delete` key](https://viewsourcecode.org/snaptoken/kilo/03.rawInputAndOutput.html#the-delete-key)

#### チュートリアル

- `Delete` キーの押下を検出する。 "\x1b[3~" というエスケープシーケンスが送られる。
- 今のところ、キーが押されても何もしない。

#### 実践

チュートリアル通りのコードを追加。  

## [4. A text viewer](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html)

### [4-1. A line viewer](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#a-line-viewer)

#### チュートリアル

- まずは "Hello, World!" と固定的に表示するコードを追加する。
- 次にファイルを1行読み取るコードに変更する。
    - コマンドライン引数に渡された名称のファイルを開いて読み込む。
    - コマンドライン引数が渡されなければ空のデータからスタートする。
    - バージョン表示のウェルカムメッセージは、ファイルを読み込まなかったときだけ表示する。

#### 実践

1行分のテキストを保持するデータ型も、Goならstring型を使用するだけで済む。  
ファイルを開いて読み取る部分も、Goで簡潔に書ける。  
コマンドライン引数処理は `os.Args` を使用したが、後に、`flag` を利用するコードに変更することになるかもしれない。  

### [4-2. Multiple lines](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#multiple-lines)

#### チュートリアル

- 行バッファを複数行保持できるようにして、ファイルの全行を読み込み表示するコードに変更する。

#### 実践

string型のスライスを使用することで、チュートリアルが大部分に費やしている動的メモリ管理処理をほぼ省略することができる。

