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

## 1. Setup

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
