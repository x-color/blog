---
title: '[Go] YAMLファイルを埋め込み構造体に読み込む'
tags:
  - golang
date: "2021-07-13T08:30:00+09:00"
published: true
---

以前、Go言語でYAMLファイルを埋め込み構造体に読み込もうとした際に、JSONを読み込むのと同じように実装したらうまく読み込めなかったことがあった。
その時調べた内容をまとめておく。

# 結論

このようなYAMLファイルがあった場合、下記のように構造体の埋め込みフィールドに`inline`タグをつけることで埋め込み構造体に読み込むことができる。

**YAMLファイル**

```yaml
# tmp.yaml
wheels: 4
tons: 10
```

**読み込むためのコード**

```go
package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Car struct {
	Wheels int `yaml:"wheels"`
}

type Truck struct {
	// 埋め込み のフィールドに 'inline' タグをつける
	Car  `yaml:",inline"`
	Tons int `yaml:"tons"`
}

func main() {
	truck := Truck{}
	b, _ := os.ReadFile("tmp.yaml")
	yaml.Unmarshal(b, &truck)
	fmt.Printf("Wheels: %+v\n", truck.Wheels)
	fmt.Printf("Tons: %v\n", truck.Tons)
}
```

下記が実行結果。

```sh
$ go run main.go
Wheels: 4
Tons: 10
```

# そもそもJSONではもっと簡単

JSONではタグを気にする必要なく埋め込み構造体へ読み込み可能。

**読み込み対象のJSONファイル**

```yaml
# tmp.json
{
    "wheels": 4,
    "tons": 10
}
```

**サンプルコード**

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Car struct {
	Wheels int `json:"wheels"`
}

type Truck struct {
	Car // タグが必要ない
	Tons int `json:"tons"`
}

func main() {
	truck := Truck{}
	b, _ := os.ReadFile("tmp.json")
	json.Unmarshal(b, &truck)
	fmt.Printf("Wheels: %+v\n", truck.Wheels)
	fmt.Printf("Tons: %v\n", truck.Tons)
}
```

**実行結果**

```sh
$ go run main.go
Wheels: 4
Tons: 10
```

# YAMLでも同じようにすると？

構造体を下記のように変更してYAMLファイルを読み込んでみる。

```go
type Car struct {
	Wheels int `yaml:"wheels"`
}

type Truck struct {
	Car // JSONと同様にタグなしにしてみる
	Tons int `yaml:"tons"`
}
```

**実行結果**

```sh
$ go run main.go
Wheels: 0
Tons: 10
```

上記のように`Wheels`が読み込めていない（`0`なのはintの初期値なため）。

# ではどうすればよい？

Go言語は大体の場合、ドキュメント見れば対処法や原因が書いてある。
なので今回も`yaml.v2`パッケージのドキュメントを見に行くと`inline`タグについて下記の記載がある。

> Inline the field, which must be a struct or a map,
> causing all of its fields or keys to be processed as if
> they were part of the outer struct. For maps, keys must
> not conflict with the yaml keys of other struct fields.

[yaml · pkg.go.dev](https://pkg.go.dev/gopkg.in/yaml.v2#Marshal)

要は`inline`タグをつけるとその構造体のフィールドはその構造体の上位の構造体のフィールドとして扱われるとのこと。

例として、`inline`タグを使用している下記構造体があったとする。

```go
type Car struct {
	Wheels int `yaml:"wheels"`
}

type Truck struct {
	Car `yaml:",inline"`　// inline タグを追加
	Tons int `yaml:"tons"`
}
```

これは、`yaml.v2`からすると下記と同等となる。

```go
type Truck struct {
	Wheels int `yaml:"wheels"`
	Tons int `yaml:"tons"`
}
```

そのため、`inline`タグを付けることによって埋め込み構造体にYAMLファイルを読み込むことができるようになる。

というわけで、これで埋め込み構造体にYAMLファイルを読み込むことができるようになった。
