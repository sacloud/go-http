> [!WARNING]
>
> このレポジトリの内容は下記に集約・移転しました。
> github.com/sacloud/sacloud-sdk-go
>
> 今後は上記のレポジトリで開発を行います。
> このライブラリを使用している方は、importを変更していただきますようお願いいたします。
>
> ```go
> // 変更前
> import "github.com/sacloud/go-http"
>
> // 変更後
> import "github.com/sacloud/sacloud-sdk-go/internal/go-http"
> ```

# sacloud/go-http

[![Go Reference](https://pkg.go.dev/badge/github.com/sacloud/go-http.svg)](https://pkg.go.dev/github.com/sacloud/go-http)
[![Tests](https://github.com/sacloud/go-http/workflows/Tests/badge.svg)](https://github.com/sacloud/go-http/actions/workflows/tests.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sacloud/go-http)](https://goreportcard.com/report/github.com/sacloud/go-http)

さくらのクラウド向けHTTPクライアントライブラリ

## 概要

さくらのクラウドの各種API(IaaS,ObjectStorage,PHYなど)で共通利用できるHTTPクライアント機能を提供します。

### 関連プロジェクト

- [sacloud/api-client-go](https://github.com/sacloud/api-client-go): sacloudプロダクト向けHTTP/APIクライアントライブラリ
  sacloud/go-httpをラップし環境変数やUsacloud互換のプロファイルの処理などを提供する。


## License

`go-http` Copyright (C) 2021-2025 The sacloud/go-http authors.

This project is published under [Apache 2.0 License](LICENSE).
