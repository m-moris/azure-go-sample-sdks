# Go sdk smaples for Azure

日本語向けのGoサンプルが少ない中、Azureにリソースを作成して試すのは敷居が高いと感じる人向け、ローカル環境で簡単にSDKの動作確認やデバッグができるサンプルを提供します。

## 対象SDK

| サービス名    |     |
| ------------- | --- |
| Storage Blob  | ○   |
| Storage Queue | ○   |
| Storage Table | ○   |

## 実行

Docker compose でストレージエミュレータを起動し、テストします。
Visual Studio Code などのIDEかのテストランナーでデバッグ実行しながら動作を確認できます。

```bash
$ docker compose up
$ go test -v -count=1 ./...
```

## リファレンス

- [Azure/azure-sdk-for-go: This repository is for active development of the Azure SDK for Go. For consumers of the SDK we recommend visiting our public developer docs at:](https://github.com/Azure/azure-sdk-for-go)


## 注意事項

- 古い Goライブラリが存在するので、混同しないように注意してください。
- エラーハンドリングは、`assert.Error(t, err)` でテストしていますが、実際のアプリケーションでは適切なエラーハンドリングを行ってください。
