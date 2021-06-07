---
title: CloudFormationで別アカウントのリソースをパラメータで受け取る際にやるべきこと
tags:
  - aws
  - cloudformation
date: "2021-06-07T09:45:00+09:00"
published: true
---

# 何をするべきか

最初に結論。クロスアカウントのリソースをパラメータで受け取る際には、CloudFormationテンプレートのパラメータのタイプを`String`, `CommaDelimitedList`のどちらかにする必要がある。

```yaml
Parameters:
  VPC:
    Type: String

  SecurityGroups:
    Type: CommaDelimitedList
```

# なぜStringなどしか指定できないのか

CloudFormationテンプレートのパラメータとしてVPCやSecurityGroupを受け取ろうとした場合、下記のように書くことが多いと思う。

```yaml
Parameters:
  VPC:
    Type: AWS::EC2::VPC::Id

  SecurityGroups:
    Type: List<AWS::EC2::SecurityGroup::Id>
```

単一アカウント上のリソースをパラメータに渡す際は上記がベストなのだが、**クロスアカウントでリソースを参照しようとした場合、これではデプロイできない**。

実際に、上記のようなCloudFormationテンプレートを用意して、デプロイ時にパラメータに別アカウントのリソースを入力した場合、下記エラーが発生してしまう。

```
Parameter validation failed: parameter value xxxx for parameter name yyy does not exists.
```

CloudFormation はデプロイ前に入力値と指定されたパラメータタイプが一致しているかなどのバリデーションを実施してくれる。
この際、パラメータに**AWS固有パラメータを指定していた場合は、対象のリソースが存在するかを確認してくれる**のだが、ここで指定されたリソースが存在しないと判断された場合は、上記エラーが報告される。

「別アカウントには存在するリソースなのに、なぜないと判断される？」と考えそうになるが、そもそもあるアカウント上から許可なく別アカウントのリソースの存在有無を確認できたらおかしい。
なので、CloudFormationは**実行されたアカウント内に対象のリソースが存在するかどうかを確認する**。

というわけで、クロスアカウントでリソースを参照する場合、 CloudFormationのパラメータタイプにはリソースの存在確認が入るAWS 固有パラメータ（`AWS::EC2::VPC::Id`, `List<AWS::EC2::SecurityGroup::Id>`など）ではなく、**`String`か`CommaDelimitedList`を利用する必要がある**。

なお、この話はしっかりとAWS公式のドキュメントに記載されている。

> テンプレートユーザーが異なる AWS アカウントからの入力値を入力できるようにする場合は、AWS 固有のタイプでパラメータを定義することはできません。代わりに、String タイプ (または CommaDelimitedList) タイプのパラメータを定義してください。

[パラメータ - CloudFormation](https://docs.aws.amazon.com/ja_jp/AWSCloudFormation/latest/UserGuide/parameters-section-structure.html#aws-specific-parameter-types)

# 結論

CloudFormationのパラメータバリデーションは、単一アカウントのみであればとても助かる機能であり、パラメータの入力ミスを減らしてくれる。
しかし、クロスアカウントでリソース参照する場合には使うことができないので、

```yaml
Parameters:
  VPC:
    Type: AWS::EC2::VPC::Id

  SecurityGroups:
    Type: List<AWS::EC2::SecurityGroup::Id>
```

ではなく、

```yaml
Parameters:
  VPC:
    Type: String

  SecurityGroups:
    Type: CommaDelimitedList
```

としましょう。
