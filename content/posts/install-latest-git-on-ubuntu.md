---
title: "Proxy下のUbuntuでGitを最新版にアップデートするまで"
date: 2021-06-02T08:45:00+09:00
tags:
  - git
  - ubuntu
published: false
---

# 結論

最初に結論。下記コマンドで最新版にアップデート可能

```sh
$ sudo -E add-apt-repository ppa:git-core/ppa
$ sudo apt update
$ sudo apt upgrade
```

下記は、これに至るまでの流れを書いておく

# Git のデフォルトブランチ名を変更したい

新たな開発環境（Ubuntu）の整備をしていた際に、「Git のデフォルトブランチの名前を`main`にしないとなぁ」と思い、下記のコマンドでデフォルトブランチを変更・・・

```sh
$ git config --global init.defaultBranch main
$ git init
$ git branch
* master
```

・・・できてない。

バージョンが古すぎて、デフォルトブランチ名の変更機能がないのかもと思い、バージョンを確認。

```sh
$ git version
git version 2.25.1
```

案の定、変更機能が追加される前のバージョンだった。（機能追加は、Git 2.28.0 以降）
というわけで、Git のバージョンアップをしなければならない。

# Ubuntu に最新版をインストール

公式のドキュメント（[Download for Linux and Unix](https://git-scm.com/download/linux)）を参考にバージョンアップを実施。

```sh
$ sudo add-apt-repository ppa:git-core/ppa
$ sudo apt update
$ sudo apt upgrade
```

上記でバージョンアップ完了かと思いきや、最初のコマンドで下記エラーが出てしまう。

```sh
$ sudo add-apt-repository ppa:git-core/ppa
Cannot add PPA: 'ppa:~git-core/ubuntu/ppa'
ERROR: '~git-core' user or team does not exist.
```

最初は？となったが、sudo したのでプロキシ周りの設定（`http_proxy`, `https_proxy`）が引き継がれていないのでは？と気づき、環境変数を引き継いで再度実行したらエラーが解決した。

```sh
$ sudo -E add-apt-repository ppa:git-core/ppa
```

というわけで、下記コマンドで Git の最新版をインストールでき、無事にデフォルトブランチ名を変えることができた。

```sh
# 最新版のGitをインストール
$ sudo -E add-apt-repository ppa:git-core/ppa
$ sudo apt update
$ sudo apt upgrade

# Gitのバージョンを確認
$ git version
git version 2.31.1

# デフォルトブランチ名を変更
$ git config --global init.defaultBranch main
$ git init
$ git branch
* main
```
