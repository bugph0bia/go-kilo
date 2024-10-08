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

### [2-1. Press `q` to quit?](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#press-q-to-quit)

#### チュートリアル

- `q`を入力することでプログラムが終了するようにする。

#### 実践

`q` が入力された場合も狩猟するようにコードを修正。  

### [2-2. Turn off echoing](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-echoing)

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

### [2-3. Disable raw mode at exit](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#disable-raw-mode-at-exit)

#### チュートリアル

- プログラムの終了時にRAWモードを無効にする。

#### 実践

`MakeRaw()` の戻り値に変更前のターミナル属性が返ってくるので、これを保存しておいてプログラム終了時に実行する。  
Goでは `atexit()` の代わりに `defer` 文を使用可能。  
前回と同様、`disableRawMode()` は実装する必要なし。  

### [2-4. Turn off canonical mode](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-canonical-mode)

#### チュートリアル

- `ICANON`をOFFにすることでCANONICALモードをOFFにする。入力を行単位ではなくバイト単位で読むことになる。
    - `I` から始まるが適用先は `c_lflag` であることに注意。

#### 実践

`MakeRaw()` の中で、`ICANON`フラグのOFFも行われているため、ここでは何もすることはない。  

### [2-5. Display keypresses](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#display-keypresses)

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

### [2-6. Turn off `Ctrl-C` and `Ctrl-Z` signals](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-ctrl-c-and-ctrl-z-signals)

#### チュートリアル

- `SIGINT`をOFFにすることで`Ctrl-C`を無効化する。
- `SIGTSTP`をOFFにすることで`Ctrl-Z`を無効化する。macOSでは`Ctrl-Y`も無効化される。

#### 実践

`MakeRaw()` の中で、`ISIG`フラグのOFFも行われているため、ここでは何もすることはない。  

### [2-7. Disable `Ctrl-S` and `Ctrl-Q`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#disable-ctrl-s-and-ctrl-q)

#### チュートリアル

- `IXON`をOFFにすることで`Ctrl-S` `Ctrl-Q`を無効化する。

#### 実践

`MakeRaw()` の中で、`IXON`フラグのOFFも行われているため、ここでは何もすることはない。  

### [2-8. Disable `Ctrl-V`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#disable-ctrl-v)

#### チュートリアル

- `IEXTEN`をOFFにすることで`Ctrl-V`を無効化する。macOSでは`Ctrl-O`も修正される。

#### 実践

`MakeRaw()` の中で、`IEXTEN`フラグのOFFも行われているため、ここでは何もすることはない。  

### [2-9. Fix `Ctrl-M`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#fix-ctrl-m)

#### チュートリアル

- `Ctrl-A`から`Ctrl-Z`まで調べると1から26の数値に対応するのだが、`Ctrl-M`だけは13ではなく10になる。他に`Ctrl-J`と`Enter`も10となる。
    - ターミナルにはCR(13, '\r')をNL(10, '\n')に変換する機能が備わっている。
- `ICRNL`をOFFにすることでこの機能を無効化する。これで`Ctrl-M`と`Enter`は変換されず13となる。

#### 実践

`MakeRaw()` の中で、`ICRNL`フラグのOFFも行われているため、ここでは何もすることはない。  

### [2-10. Turn off all output processing](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#turn-off-all-output-processing)

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

### [2-11. Miscellaneous flags](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#miscellaneous-flags)

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

### [2-12. A timeout for `read()`](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#miscellaneous-flags)

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

### [2-13. Error handling](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#error-handling)

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

### [2-14. Sections](https://viewsourcecode.org/snaptoken/kilo/02.enteringRawMode.html#sections)

#### チュートリアル

- ソースコードにセクションコメントを追加する。

#### 実践

チュートリアルではソースコードをセクションコメントで分割しているが、それには倣わないこととする。  

### `MakeRaw()` を使用せずチュートリアルに合わせる方針に変更

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

チュートリアルでは、エラー用の関数 `die()` の中にスクリーン消去処理を書いているが、本コードでは `panic()` を利用している。  
また、`atexit()` の中にスクリーン消去処理を書くとエラーメッセージまで消えてしまうとあるが、Goの`panic()`はdeferを適切に処理してからエラーメッセージを表示してくれる。  
このことから、本コードではdeferを用いて前項で作成した `editorRefreshScreen()` を呼び出しすことでスクリーン消去することとした。  


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

チュートリアルではエディタの状態を保持するグローバル変数は `E` としているが、Goでは大文字の変数名は外部公開されてしまうため避けたい。そのため本コードでは `ec` という変数名にする。  

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

### [4-3. Vertical scrolling](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#vertical-scrolling)

#### チュートリアル

- 行方向のオフセットを管理することで垂直スクロールを表現する。
- カーソルの y 位置を制御する処理をすべて見直す。

#### 実践

チュートリアル通りのコードを追加。  

### [4-4. Horizontal scrolling](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#horizontal-scrolling)

#### チュートリアル

- 列方向のオフセットを管理することで垂直スクロールを表現する。
- カーソルの x 位置を制御する処理をすべて見直す。

#### 実践

チュートリアル通りのコードを追加。  

### [4-5. Limit scrolling to the right](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#limit-scrolling-to-the-right)

- 現在行と現在列の1つ次までしかカーソル移動しないように制限をする。

#### 実践

チュートリアル通りのコードを追加。  
この時点では、`End` キーや、上下移動時にうまく制限はかからない。TAB文字や多バイト文字があるときもうまくいかない。  

### [4-6. Snap cursor to end of line](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#snap-cursor-to-end-of-line)

#### チュートリアル

- 長い行の末尾にカーソルを置いて、上下の短い行に移動したときにカーソルを末尾にスナップする処理を追加する。

#### 実践

チュートリアル通りのコードを追加。  

### [4-7. Moving left at the start of a line](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#moving-left-at-the-start-of-a-line)

#### チュートリアル

- 行頭で左矢印キーを押したときに前の行の末尾に移動するようにする。

#### 実践

チュートリアル通りのコードを追加。  

### [4-8. Moving right at the end of a line](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#moving-right-at-the-end-of-a-line)

#### チュートリアル

- 行末で右矢印キーを押したときに次の行の先頭に移動するようにする。

#### 実践

次の行に移動する前にファイルの末尾でないことを確認すべきである。  
チュートリアルでは、 `else if (row && E.cx == row->size)` のように確認しており、ファイルの末尾にいる場合は `row == NULL` になっていることから、この条件でチェックが可能となっている。  
ただし、後半の `E.cx == row->size` は直前の if の条件の逆であり冗長のように思える。  
本コードでは、`row == nil` になることはないため、別の条件式でファイル末尾でないことをチェックする必要がある。  

### [4-9. Rendering tabs](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#rendering-tabs)

#### チュートリアル

- 現時点で、タブ文字の描画には下記の問題がある。
    - タブ文字が8文字分のスペースを使用している。
    - タブ文字が描画された後ろに、前に表示されていた文字列が消えずに残ってしまう。この挙動は、ターミナル上でEnterキーを複数回押して画面いっぱいにプロンプトを表示した後、エディタで Makefile を表示することで確認することができる。
- タブ文字を最大8文字のスペースに変換してレンダリングを行う。

#### 実践

レンダリングのコードを実装する前にリファクタリングをする。  
行バッファにstring型を利用できたことから `eRow` 構造体は必要ないと考えていたが、レンダリング用の変数も格納する必要が出てきたので、改めて構造体を作成する。  

Goでは文字列と構造体の使い方が異なるため、`editorUpdateRow`は少しチュートリアルとは異なるコードとなる。  
中で行うタブをスペースに変換する処理はほぼチュートリアル通りで、効率化のため、Goでも`render`に必要となるバイト数を予め計算してキャパシティに確保するコードとした。  

この時点では、タブ文字が含まれる行で左右キーによるカーソル移動をしてもうまく動作しない状態。  

### [4-10. Tabs and the cursor](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#tabs-and-the-cursor)

#### チュートリアル

- 実際のテキスト上の位置を表す cx に対してレンダリング位置を表す rx を追加する。
- cx から rx を計算する処理を用意して、スクロール発生時に計算させる。
- cx を使用してカーソル移動していた処理を、rx に変更する。

#### 実践

チュートリアル通りのコードを追加。  

### [4-11. Scrolling with `Page Up` and `Page Down`](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#scrolling-with-page-up-and-page-down)

#### チュートリアル

- `Page Up` と `Page Down` の処理を改良する。
- それぞれ、1ページ分上下スクロールを行い、カーソル位置も上端／下端に移動する処理とする。

#### 実践

チュートリアル通りのコードを追加。  
本コードではもともと pageUp / pageDown 用のif文を用意していたので、そちらに組み込むようにコードを実装。  

### [4-12. Move to the end of the line with `End`](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#move-to-the-end-of-the-line-with-end)

#### チュートリアル

- `End` の処理を改良する。
    - 画面の右端に移動するのではなく、現在行の末尾に移動する。
- `Home` は変更する必要がない。

#### 実践

チュートリアル通りのコードを追加。  

### [4-13. Status bar](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#status-bar)

#### チュートリアル

- 画面下端にステータスバーを置くスペースを確保する。
- ステータスバーを目立たせるために色を反転する。
    - エスケープシーケンス "\x1b[7m" で反転色にして、"\x1b[m" で元に戻すことができる。
    - m コマンドは指定する引数ごとに、それ以降に印字するテキスト属性を変更することができる。
        - 太字(1)、下線(4)、点滅(5)、反転色(7)など。まとめて "\x1b[1;4;5;7m" のように指定することもできる。
        - 引数 0 で全属性のクリア。これはデフォルト引数である。
- ステータスバーにファイル名と行数、現在行を表示する。

#### 実践

`Printf` の書式指定子は基本的にCとGoで同じ。  
"%.20s" は精度20の文字列であり、20文字を超える場合はそれ以上表示しないという意味になる。  

チュートリアルでは、現在行は表示するだけの余裕があるときにのみ表示するコードとなっているので、それに合わせる。  

### [4-14. Status message](https://viewsourcecode.org/snaptoken/kilo/04.aTextViewer.html#status-message)

#### チュートリアル

- ステータスバーの下にもう一行、ステータスメッセージのスペースを確保する。
- ステータスメッセージが表示されてから一定時間で消えるようにするために、タイムスタンプも保持する。
    - Cの `time_t` 型は、1970年1月1日 0:00:00 からの経過時間の秒数であるUNIXタイムスタンプで時刻を保持する。
    - メッセージがセットされてから5秒経過後にキー操作をするとメッセージが消える。

#### 実践

Cで可変長引数を利用した文字列出力を行う場合は、`va_list` 系のデータ型や関数を用いることになるがGoではもっと簡潔に利用することができる。  
`...` を利用することで可変長引数型を利用することができるし、`Printf`系の関数に渡すときは `x...` のようにアンパッキングして渡せば良い。  

Goの `time.Time` 型はUNIXタイムスタンプではなく 1年1月1日 0:00:00 からの経過時間を用いているらしい。  
いずれにしても使う際に意識する必要はない。  
記憶しておいた時刻から現在時刻までの経過時間を調べるには `time.Since` を用いる。`time.Now().Sub(xx)` でも同じ結果が得られる。  
経過時間は `time.Duration` 型で表現される。リテラルを用いる場合は `5 * time.Second` のように数値に専用の定数を乗算して表現することに注意。  

## [5. A text editor](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html)

### [5-1. Insert ordinary characters](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html#insert-ordinary-characters)

#### チュートリアル

- カーソル位置に一文字挿入する処理を追加する。

#### 実践

`editorUpdateRow()` はチュートリアルとは異なるシグネチャになっていたが、合わせることとした。引数に対象行のポインタを取って直接中を書き換えるコードのほうが、"Update Row" のイメージに合うため。  

それ以外に下記のコード変更を行う。  

- 文字のバイトコードを保持する変数をint型にしていたが、Goのイディオムに倣いrune型に統一する。
    - 多バイト文字については一旦考えないこととする。デバッグ時に読み込み対象とするファイルもASCIIのみで構成される Makefile にする。
- 最終行に移動してから文字挿入し続けると panic が発生するバグが `editorMoveCursor()` に存在した。新しい行に移動したときのカーソルX位置が0にリセットされていなかったことが原因。修正する。
    - 4-6. Snap cursor to end of line で作り込んだバグ。

### [5-2. Prevent inserting special characters](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html#prevent-inserting-special-characters)

- `Backspace` や `Enter` を押したときに特殊文字がそのまま挿入されてしまうことを防ぐ。
- `Backspace` には "\r" や "\n" のようなバックスラッシュエスケープ表現が無いので、定数でASCIIコードの 127 を直接定義する。`Ctrl-H`も使用できる。
    - ASCIIコード表では `Backspace` は 8 、`elete` は 127 だが、現在では `Backspace` が 127、`Delete` は "\x1b[3~" にマップされている。
- `Enter` は "\r" と表現できる。一旦 TODO コメントだけ残しておく。
- `Ctrl-L` はターミナルの画面リフレッシュ。これは無視する。
- これまでに実装したものを除いたエスケープシーケンス（ファンクションキーなど、すべて `\x1b` から始まる）についても、すべて無視する。

#### 実践

チュートリアル通りのコードを追加。  

### [5-3. Save to disk](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html#save-to-disk)

#### チュートリアル

- テキストバッファの内容を1つの文字列に変換する。
- 保持しているファイル名を利用してファイルを上書き保存する。ファイル名を保持していない場合の対応は後で実装する。
- ファイルの保存成功／失敗をメッセージエリアで伝える。起動直後のヘルプメッセージにも Ctrl-S について追記する。

#### 実践

ファイルの書き込みについては、チュートリアルと同じ低レベルのファイルオープンを使用して、同様のフラグとパーミッションを指定する。  
ファイルの読み込みのときはそれらを行わなかったので、対称性が無いようにも思える。  

既存ファイルへの上書きの際、ファイルを Truncate する処理は必須？安全な書き込みのために必要という説明であるが、ほぼ利用したことがない関数であり必要性をきちんと理解できていない。  
書き込む長さが元のファイルの内容より小さくなる場合に、Truncate していないとうまくいかないということかもしれないが、現時点ではそれを試すことができない。  

Goのコンパイラはシングルパスではないので、関数のプロトタイプ宣言を書くことはない。  

### [5-4. Dirty flag](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html#dirty-flag)

#### チュートリアル

- ファイルが、開いた時点／最後に保存した時点から変更されているかどうかをダーティフラグで保持する。
- ダーティフラグが立っている場合はステータスバーに "(modified)" と表示することでユーザに伝える。

#### 実践

フラグなのでbool型で表現することも考えたが、チュートリアルの説明に「どの程度ファイルが汚れているかを表現することもできるのでint型にしている」とあるので、それに倣うことにする。  
その場合、厳密にいえばもはやフラグではない。  
ただ、このチュートリアルの中ではフラグ（0か否か）の意味でしか使用しない模様。  

### [5-5. Quit confirmation](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html#quit-confirmation)

#### チュートリアル

- ダーティフラグを利用して、未保存の変更があるときに `Ctrl-Q` を押してプログラムを終了しようとしたときに、あと3回 `Ctrl-Q` を押すように要求する。

#### 実践

キー押下残り回数を保持する変数をチュートリアルでは静的変数としているが、Goには静的変数は無いためパッケージ変数で代用する。  
変数を利用する関数 `editorProcessKeypress()` の近くに定義する。  

本コードでは `editorProcessKeypress()` の戻り値でプログラム終了を表現しているが、今回途中リターンが出てきたため、名前付き戻り値に変更することでコードを簡潔化する。  

### [5-6. Simple backspacing](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html#simple-backspacing)

#### チュートリアル

- バックスペースを実装する。
    - `Backspace` `Ctrl-H` でカーソルの左側の1文字を削除、`Delete` はカーソル位置の文字を削除。
    - `Delete` はカーソルを右に1つ移動してから `Backspace` を押したことと同じである。

#### 実践

`editorRowDelChar()` の呼び出し箇所は、Cだと前置デクリメントで書きたくなるところだが、Goのインクリメントとデクリメントには前置スタイルは存在せず、また式ではなく文であるため、他の式の中に混ぜて書けないことに注意。  

### [5-7. Backspacing at the start of a line](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html#backspacing-at-the-start-of-a-line)

#### チュートリアル

- 行頭でバックスペース処理を行った場合、現在行の文字列を前の行の末尾に連結して、現在行を削除する。
    - 前章で説明された、`Delete` はカーソルを右に1つ移動してから `Backspace` という動作についても期待通りとなる。行末で `Delete` を押したときに正しく動作する。

#### 実践

スライスから要素を削除して前詰めする方法は複数存在するが、実行効率の良い `copy()` を使用する方法を採用した（[参考](https://www.how2go.dev/docs/standard/slice/#%e3%82%b9%e3%83%a9%e3%82%a4%e3%82%b9%e3%81%ae%e8%a6%81%e7%b4%a0%e3%82%92%e5%89%8a%e9%99%a4%e3%81%99%e3%82%8bcopy-%e3%82%92%e4%bd%bf%e3%81%a3%e3%81%9f%e5%a0%b4%e5%90%88)）。  

本コードでは `editorFreeRow()` は実装不要。  

行の連結が発生したときに、カーソル位置が前の行の末尾に移動しないことに違和感があるが、今後解消されるか？  

### [5-8. The `Enter` key](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html#the-enter-key)

#### チュートリアル

- `Enter` キーを押すことで、行を分割したり新しく挿入したりできるようにする。
    - 以前に作った `editorAppendRow()` を `editorInsertRow()` に改造してこれを実現する。
    - カーソル位置が行頭の場合は空行を挿入、それ以外の場合は現在行をカーソル位置で二分割する。

#### 実践

スライスへ要素を挿入するための効率的な方法を検討した（[参考](https://mattn.kaoriya.net/software/lang/go/20200404155447.htm)）。  
調査の過程で、golang v1.22 で追加された slices パッケージの存在を知り、`slices.Insert()` を利用することとする。  
前の章で実装したスライスから要素を削除する方法についても、`slices.Delete()` に置き換える。  

### [5-9. Save as...](https://viewsourcecode.org/snaptoken/kilo/05.aTextEditor.html#save-as)

#### チュートリアル

- 引数なしでプログラムを起動したときにファイルを保存する方法がないため、プロンプトでファイル名を入力できるようにする。
- プロンプト処理は、`Enter` で入力確定、`ESC` で入力中断、`Backspace` `Delete` で1文字削除する処理を用意する。

#### 実践

`editorPrompt()` は入力が確定されたか中断されたかを2つ目のbool型の戻り値で返すようにすることで、カンマokイディオムを利用できるようにする。  

## [6. Search](https://viewsourcecode.org/snaptoken/kilo/06.search.html)

#### チュートリアル

- `editorPrompt()` を利用して簡単な検索機能を実装する。
- `Ctrl-F` を検索機能にマッピングする。

#### 実践

チュートリアル通りのコードを追加したが、forループは少し自分の好みのコードに変えた。  

### [6-1. Incremental search](https://viewsourcecode.org/snaptoken/kilo/06.search.html#incremental-search)

#### チュートリアル

- さらに、インクリメンタル検索を実装する。
- `editorPrompt()` にコールバック関数を渡せるようにして、ユーザがキーを入力するたびに、それまでに入力された文字列と最後に入力された文字を渡すようにする。
- `Enter` `ESC` が入力されたら検索を終了する。

#### 実践

コールバック関数は、Goではクロージャで代用できる。  
`editorPrompt()` の中でしか使わない関数なので、クロージャにすることで利用するスコープを明確にし、パッケージ（グローバル）の名前空間を汚染することも防ぐことができる。  

### [6-2. Restore cursor position when cancelling search](https://viewsourcecode.org/snaptoken/kilo/06.search.html#restore-cursor-position-when-cancelling-search)

#### チュートリアル

- `ESC` で検索を中断したとき、カーソル位置を検索を始める前の状態に戻すようにする。

#### 実践

検索が中断されたことを知るために、カンマokイディオムが利用できる。  

### [6-3. Search forward and backward](https://viewsourcecode.org/snaptoken/kilo/06.search.html#search-forward-and-backward)

#### チュートリアル

- 矢印キーを使って、前方検索と後方検索を実現する。左キーと上キーで前方検索、右キーと下キーで後方検索とする。

#### 実践

チュートリアルでは `lastMatch` `direction` の2つの変数を `editorFindCallBack()` 内の静的変数としているが、Goでは静的変数は使えない。  
ただ、`editorFindCallBack()` を、`editorFind()` 内のクロージャとして実装してあるため、2つの変数は単純に `editorFind()` 内のローカル変数とすれば `editorFindCallBack()` に束縛されて同じ機能を実現することができる。  
Cの静的変数はプログラム起動時の一度しか初期化されないが、ローカル変数であれば毎回初期化される効果もある。これによって、`Enter` `ESC` キー押下時に行っている、次回の検索のための変数初期化を省略することができる。  

インデックス値を前後の指定した方向に進めながら走査する方法として、 `direction` を -1 または +1 にする手法はこれまでもよく使ったことがある。  
このとき、インデックスの取りうる範囲を超えるタイミングで数値をローテーションさせる必要がある。  
チュートリアルでは、if - else if でこれを実現しているが、本ツールでは % 演算子を使って `idx = (idx + length) % length` のようにワンライナーで記述することにした。  
これはGoのイディオムではなく私の好みであるが、Goでは if - else if を書くのに複数行必要になってしまうところ、一行に押さえることができたのでより効果的だと思う。  

この章に限ったことではないが、構造体変数を別の変数に代入したり関数の引数に渡すとき、値を渡すべきかアドレスを渡すべきか、悩む。  
C言語の経験に照らし合わせれば間違いなくアドレス渡しとすべきだが、Goのイディオム的には値渡しで良いという説明をどこかで見た記憶があり、値渡しとする。  

## [7. Syntax highlighting](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html)

### [7-1. colorful digits](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#colorful-digits)

#### チュートリアル

- 数字を赤色にする。
- 色の変更にはエスケープシーケンスを利用する（[参考](https://en.wikipedia.org/wiki/ANSI_escape_code)）。

#### 実践

文字色を設定するようなライブラリは利用せず、チュートリアルと同じようにエスケープシーケンスで色を付ける。  

### [7-2. Refactor syntax highlighting](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#refactor-syntax-highlighting)

#### チュートリアル

- `hl` というバイト配列を用意して、テキストのクラス（数値なのか、キーワードなのか、など）の情報を保持する。
- `hl` は `render` の文字列に対応するように、1文字ずつ色情報を格納する。
- `hl` の内容に合わせてANSIカラーコードを返す関数を用意して、行データを描画するときに色指定のエスケープシーケンスを挿入する。
    - 色に変化があるときだけエスケープシーケンスを挿入するように最適化する。

#### 実装

`editorUpdateSyntax()` でスライスの初期化が必要。各要素はゼロ値にリセットされるので `hlNumber` を埋める処理は不要にも思えるが、`hlNumber` がゼロ値であり続ける保証はないので、きちんと初期化も書くことにする。  
Goには `memset` に相当する関数はないのでループで初期化することになる。

`editorSyntaxToColor()` は map を使ってもよいかもしれないが、必ずしも 1:1 対応しないかもしれないので（default節にどこまで頼るかわからないので）、チュートリアル通りの関数とする。  

### [7-3. Colorful search results](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#colorful-search-results)

#### チュートリアル

- 検索中にマッチした文字列を青色にハイライトする。

#### 実践

チュートリアル通りのコードを追加。  

### [7-4. Restore syntax highlighting after search](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#restore-syntax-highlighting-after-search)

#### チュートリアル

- ユーザが検索を終えたらマッチ文字列のハイライトを解除する。
- そのために、マッチ文字列をハイライトする前に、対象の行インデックスと変更前のハイライト状態を保存してリストアできるようにする。

#### 実践

状態を保存する変数は、以前と同じようにクロージャに束縛されるローカル変数として用意することができる。  
ハイライト前の状態を保存する際、`[]byte` スライスの内容をコピーすることになるが、=演算子の代入では状態を保存したことにならないため `copy()` 関数でディープコピーする必要がある。  

### [7-5. Colorful numbers](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#colorful-numbers)

#### チュートリアル

- 区切り文字を判定する関数を作成する。
- 区切り文字のあとに来る連続する数字（小数点の "." も許容）をハイライトする。

#### 実践

区切り文字を判定するための処理は、CとGoで使用する関数が異なる。  
`isspace()`は、Goでは `unicode.IsSpace()` で置き換える。  
"\0" の判定は、Goでは文字列終端に利用されないため不要に思えるが、一応入れておく。  
`strchr()` については、Goでは `strings.ContainsAny()` を代わりに利用できる。このような文字種そのものを重視するようなケースにおいては、raw文字列リテラルを使用することとする。  

### [7-6. Detect filetype](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#detect-filetype)

- `editorSyntax` 構造体を作成して、ファイルタイプに関する情報を格納できるようにする。
    - まずはCのファイルタイプの情報を格納する。
- 現在のファイルタイプをステータスバーに表示する。
- シンタックスが更新されたら全行のハイライトを更新する。

#### 実践

Goのスライスは長さの情報を持つため、終端NULLや `HLDB_ENTRIES` のような #define により配列長を計算させるようなハックは必要ない。  
チュートリアルと同様、現在のシンタックス情報を保持する変数はポインタ型にしておき、`nil` でシンタックス無しを表現できるようにしておく。  

ファイルタイプの判定は愚直に行うのではなく、正規表現パターンマッチを使用することとする。  
拡張子の判定は ".c$" のようにしてファイル名の末尾を判定すれば良い。  

### [7-7. Colorful strings](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#colorful-strings)

#### チュートリアル

- 文字列リテラルをマゼンタ色にハイライトする。
- シングルクォーテーション、ダブルクォーテーションの両方に対応する。
    - バックスラッシュでエスケープされている場合は除く。

#### 実践

Goの `iota` はビットフィールドの定義にも利用できる。  

本コードでは while ループではなく for ループを使用しているので、インデックスをインクリメントするロジックを少し変える。  

### [7-8. Colorful single-line comments](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#colorful-single-line-comments)

#### チュートリアル

- 一行コメントをハイライトする。

#### 実践

`strncmp()` の代わりに `strings.HasPrefix()` を使用してコメントの開始を判定する。  

### [7-9. Colorful keywords](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#colorful-keywords)

#### チュートリアル

- Cの予約後のうち、型名以外をキーワード1、型名をキーワード2としてハイライトする。
- キーワード2の末尾に "|" をつけてキーワード1と区別する。

#### 実践

チュートリアルのように、キーワード1と2を区別するのに "|" の有無で判定するのは好みではないので、保持する変数そのものを区別することにする。  
その結果、ハイライト処理のロジックを2回実行することになるため、共通処理をクロージャに抜き出して冗長さを無くす。  

行末で改行したときの処理にバグがあったので修正する。  

### [7-10. Nonprintable characters](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#nonprintable-characters)

#### チュートリアル

- 印字不可能な文字を印字する。
    - `Ctrl-A` ～ `Ctrl-Z` は "A" ～ "Z" とする。
    - `Ctrl-@` は "@" とする。
    - それ以外の印字不可能文字は "?" とする。
    - これらを印字する場合は、色を反転させて白背景に黒文字で表示する。

#### 実践

チュートリアル通りのコードを追加。  

### [7-11. Colorful multiline comments](https://viewsourcecode.org/snaptoken/kilo/07.syntaxHighlighting.html#colorful-multiline-comments)

#### チュートリアル

- 複数行コメントのハイライトを実装する。色は一行コメントと同じ。
- 文字列リテラル中にコメントの開始・終了があっても無視する。
- 複数行コメントの中の一行コメントの開始は無視する。
- `eRow` 構造体に自分自身の行インデックスと複数行コメント内かどうかのフラグを追加して、複数行コメントの判定を行えるようにする。
    - 行データを挿入・削除したときにインデックスを更新する。
    - ある行のコメント状態に変化があった場合は次の行のシンタックスハイライトの更新が必要となる。

#### 実践

定数名はGoではキャメルケースを使用するため、分かり易さを重視して `hlMultiLineComment` とした。  

チュートリアルとはコードが少し異なるため、行インデックスを更新するタイミングやループの範囲に注意する必要がある。  
