---
title: CloudFormationで作成するリソースのタグ付けを強制する
date: 2021-05-31T08:00:00+09:00
tags:
  - aws
  - cloudformation
  - openpolicyagent
published: false
---

# タグ付けを忘れがち

普段の業務では単一の AWS アカウントを複数のプロダクトの開発に利用している。
その場合に気になってくるのが**コスト**。
気づいたときには「今月の請求が。。」みたいなことにならないためにも、どのプロダクトがどのリソースを利用しており、それぞれどの程度利用料が発生しているのかを把握することはとても重要。

このニーズを満たすために、AWS では[コスト配分タグ](https://docs.aws.amazon.com/ja_jp/awsaccountbilling/latest/aboutv2/cost-alloc-tags.html)という機能がある。これを用いることで、タグごとにリソースのコストを把握する事ができる。

自分は普段、なにかを構築する際には CloudFormation を利用する事が多い。
コスト配分タグを最大限活用するためにも可能な限りリソースにはタグ付けをするようにしている。

しかし、多くのリソースを構築していると一部リソースのタグ付けを忘れてしまうことがある。おかげでタグで追跡できないリソースが作成され、「これコストかかっているけど、どのプロダクトのリソース？」といったことが発生する。

※CloudFormation スタックのタグ付けの伝搬でいいのではと最初思っていたのだが、意外とタグが伝搬されないリソースが多いので各種リソースに明示的につけるようにしている。

# 忘れないために自動でテストしよう

レビュー時に、テンプレートで定義されたすべてのリソースが正しくタグ付けされているかをチェックするのもいいが、そもそも「どのリソースはタグ付け可能だっけ？」となったりしてあまりにも大変。

というわけで人の手を借りずに自動でチェックさせるために、タグ付け可能なリソースがすべて適切なタグ付けがされているかをテストするための仕組みを[Open Policy Agent](https://www.openpolicyagent.org/)を用いて実装した。

<a href="https://github.com/x-color/opa-cfn-tag-policy"><img src="https://github-link-card.s3.ap-northeast-1.amazonaws.com/x-color/opa-cfn-tag-policy.png" width="460px"></a>

ちなみに、CloudFormation は高頻度で更新がかかり、随時新たなリソースの追加やタグ付けがサポートされる。
そのため、このリポジトリでは GitHub Actions を利用して、CloudFormation の更新を追いかけ、更新があるたびにポリシーファイルを更新している。
**利用する場合はなるべく最新版をダウンロードして利用してもらいたい。**

# テストしてみる

実際に利用してみる。今回は、**タグ付け可能なリソース全てに「System」タグが付けられていること**をテストする。

なお、[ここ](https://github.com/x-color/opa-cfn-tag-policy/tree/main/example)にもサンプルのポリシーや CloudFormation テンプレートを用意している。

## テストのための準備

ディレクトリ構成は下記とする

```sh
.
├── policy
│   └── deny.rego # 実際のポリシーを定義
└── templates
    └── template.yaml # タグ付けされていないリソースが定義されたCloudFormationテンプレート
```

テスト対象とする CloudFormation テンプレートは下記とする。（`templates/template.yaml`）

```yaml
AWSTemplateFormatVersion: 2010-09-09

Parameters:
  VpcCidr:
    Type: String

Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: !Ref VpcCidr
```

このテンプレートでは、指定された Cidr ブロックを持つ VPC を作成するが、「System」タグを付け忘れている。

## タグ付けチェック用のポリシーファイルの作成

次に`policy/deny.rego` に 「System」タグが付与されていない場合に警告されるポリシーを定義する。

```rego:policy/deny.rego
# policy/deny.rego
package policy

import data.cloudformation as cfn # 「opa-cfn-tag-policy」のポリシーをインポート

deny[reason] {
    some id
    rs := input.Resources[id] # *1
    not cfn.resource_has_tag(rs, "System") # *2
    reason = sprintf("No 'System' tag: %v", [id])  # *3
}
```

処理の流れは下記となっている。

1. 処理中のリソース名を`id`に格納しつつ、すべてのリソースを取得
2. 対象のリソースが「System」タグを持っているか確認
3. 「System」タグを持っていないリソースに対してリソース名を含んだ警告文を返す

なお、このあとのテスト時の手順に入っているが、テスト実行前に上記リポジトリ内のタグ付けされているかチェックする関数を定義している[ポリシーファイル](https://github.com/x-color/opa-cfn-tag-policy/blob/main/policy/cfn_tag.rego)をダウンロードする必要がある。

## ポリシーを用いてテスト

Open Policy Agent の CLI を用いてチェックする場合は下記コマンドでチェック可能。

```sh
# 「opa-cfn-tag-policy」リポジトリ内のタグ付けポリシーをダウンロード
$ curl -L -o policy/cfn_tag.rego https://raw.githubusercontent.com/x-color/opa-cfn-tag-policy/main/policy/cfn_tag.rego

# テストの実施
$ opa eval -d policy -i templates/template.yaml data.policy
{
  "result": [
    {
      "expressions": [
        {
          "value": {
            "deny": [
              "No 'System' tag: VPC"  # 「VPC」に「System」タグがないと報告されている
            ]
          },
          "text": "data.policy",
          "location": {
            "row": 1,
            "col": 1
          }
        }
      ]
    }
  ]
}
```

[Conftest](https://www.conftest.dev/)を用いるとより簡単にテストすることが可能。

```sh
# 「opa-cfn-tag-policy」リポジトリ内のタグ付けポリシーをダウンロード
$ conftest pull https://raw.githubusercontent.com/x-color/opa-cfn-tag-policy/main/policy/cfn_tag.rego

# テストの実施
$ conftest test -n policy templates
FAIL - templates/vpc-template.yaml - policy - No 'System' tag: VPC

1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions
```

# さいごに

実際に作成したテストの仕組みは CI/CD パイプラインに組み込んでおり、おかげで CloudFormation テンプレートで作成するリソースのタグ付けを忘れることがなくなった。

タグ付け忘れで、用途不明なリソースができていてコスト管理などで困っている場合は、ぜひ利用してみてほしい。
