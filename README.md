## Lambda by Go

Slack から AWS Chatbot 経由で Lambda 実行をする仕組みを構築した際の Lambda 関数

## zip 作成手順

[.zip ファイルアーカイブを使用して Go Lambda 関数をデプロイする](https://docs.aws.amazon.com/ja_jp/lambda/latest/dg/golang-package.html)を実施するための手順は以下。

### 1、build-lambda-zip ツールを GitHub からダウンロード

`go.exe get -u github.com/aws/aws-lambda-go/cmd/build-lambda-zip`を cmd で実行（初回のみ）

### 2、build-lambda-zip ツールで zip を作成

`set GOOS=linux`を cmd で実行（初回のみ）<br>
続いて以下のコマンドを順に実行して zip を作成する

```
go build -o main main.go
%USERPROFILE%\Go\bin\build-lambda-zip.exe -output main.zip main
```

zip 作成後は AWS Management console から zip を upload して Deploy をする

## zip 作成時に他の設定ファイルなどを取り込みたい時

以下のように、zip 化する際にその対象のファイルを指定すればよい

```
C:\Users\user\go\bin\build-lambda-zip.exe -output main.zip main config.yaml
```

参考：https://stackoverflow.com/questions/56607738/cannot-read-json-file-after-uploading-go-package-into-aws-lambda

## Cloud Watch Events から Lambda 実行

### PR マージ時の Event オブジェクトの中身

dev -> main への PR のマージを行うと以下のような Event オブジェクトが Lambda 関数に渡る

- isMerged
- event

の 2 つで PR のマージか？が判別可能そう

```json
{
  "version": "0",
  "id": "c0931ce3-ee9f-d9f5-205c-a0ab8e3c0f9f",
  "detail-type": "CodeCommit Pull Request State Change",
  "source": "aws.codecommit",
  "account": "xxxxxxxxxxxx",
  "time": "2021-08-16T08:32:25Z",
  "region": "ap-northeast-1",
  "resources": [
    "arn:aws:codecommit:ap-northeast-1:xxxxxxxxxxxx:slack-notification-test"
  ],
  "detail": {
    "author": "arn:aws:iam::xxxxxxxxxxxx:user/----------",
    "callerUserArn": "arn:aws:iam::xxxxxxxxxxxx:user/----------",
    "creationDate": "Mon Aug 16 08:31:23 UTC 2021",
    "destinationCommit": "dfdaf17904c816c9826d431187b4746d21406319",
    "destinationReference": "refs/heads/main",
    "event": "pullRequestMergeStatusUpdated",
    "isMerged": "True",
    "lastModifiedDate": "Mon Aug 16 08:32:14 UTC 2021",
    "mergeOption": "FAST_FORWARD_MERGE",
    "notificationBody": "A pull request event occurred in the following AWS CodeCommit repository: slack-notification-test. User: arn:aws:iam::xxxxxxxxxxxx:user/----------. Event: Updated. Pull request name: 4. Additional information: The pull request merge status has been updated. The status is merged. For more information, go to the AWS CodeCommit console https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/slack-notification-test/pull-requests/4?region=ap-northeast-1.",
    "pullRequestId": "4",
    "pullRequestStatus": "Closed",
    "repositoryNames": ["slack-notification-test"],
    "revisionId": "9dc053117e28d1da9cbb7044eddb2c6e815fc978d61183e8cb62a59997599037",
    "sourceCommit": "dced724ba1e54ef52319050306015d23a47f42d2",
    "sourceReference": "refs/heads/dev",
    "title": "test"
  }
}
```
