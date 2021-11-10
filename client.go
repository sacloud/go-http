// Copyright 2021 The sacloud/go-http authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// TODO 後で別プロジェクトへ切り出す
const (
	// APIAccessTokenEnvKey APIアクセストークンの環境変数名
	APIAccessTokenEnvKey = "SAKURACLOUD_ACCESS_TOKEN"
	// APIAccessSecretEnvKey APIアクセスシークレットの環境変数名
	APIAccessSecretEnvKey = "SAKURACLOUD_ACCESS_TOKEN_SECRET"
)

var (
	// DefaultUserAgent デフォルトのユーザーエージェント
	DefaultUserAgent = fmt.Sprintf(
		"go-http/v%s (%s/%s; +https://github.com/sacloud/go-http)",
		Version,
		runtime.GOOS,
		runtime.GOARCH,
	)

	// DefaultAcceptLanguage デフォルトのAcceptLanguage
	DefaultAcceptLanguage = ""

	// DefaultRetryMax デフォルトのリトライ回数
	DefaultRetryMax = 10

	// DefaultRetryWaitMin デフォルトのリトライ間隔(最小)
	DefaultRetryWaitMin = 1 * time.Second
	// DefaultRetryWaitMax デフォルトのリトライ間隔(最大)
	DefaultRetryWaitMax = 64 * time.Second

	// DefaultCheckRetryFunc デフォルトのリトライ判定Func
	DefaultCheckRetryFunc = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if err != nil {
			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}
		if resp.StatusCode == 0 || resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusLocked {
			return true, nil
		}
		return false, nil
	}
)

// Client さくらのクラウドAPI(secure.sakura.ad.jp)向けのHTTPクライアント
//
// レスポンスの状態に応じてリトライする仕組みを持つ
// デフォルトだとレスポンスステータスコード423、または503を受け取った場合にRetryMax回リトライする
//
// リトライ間隔はRetryMinからRetryMaxまで指数的に増加する(Exponential Backoff)
//
// リトライ時にcontext.Canceled、またはcontext.DeadlineExceededの場合はリトライしない
type Client struct {
	// AccessToken アクセストークン
	AccessToken string `validate:"required"`
	// AccessTokenSecret アクセストークンシークレット
	AccessTokenSecret string `validate:"required"`
	// ユーザーエージェント
	UserAgent string
	// Accept-Language
	AcceptLanguage string
	// Gzipを有効にするか
	Gzip bool
	// CheckRetryFunc リトライすべきか判定するためのfunc
	CheckRetryFunc func(ctx context.Context, resp *http.Response, err error) (bool, error)
	// リトライ回数
	RetryMax int
	// リトライ待ち時間(最小)
	RetryWaitMin time.Duration
	// リトライ待ち時間(最大)
	RetryWaitMax time.Duration
	// APIコール時に利用される*http.Client 未指定の場合http.DefaultClientが利用される
	HTTPClient *http.Client
	// RequestCustomizer リクエスト前に*http.Requestのカスタマイズを行うためのfunc
	RequestCustomizer RequestCustomizer
}

// NewClient APIクライアント作成
func NewClient(token, secret string) *Client {
	c := &Client{
		AccessToken:       token,
		AccessTokenSecret: secret,
	}
	return c
}

// NewClientFromEnv 環境変数からAPIキーを取得してAPIクライアントを作成する
func NewClientFromEnv() (*Client, error) {
	token := os.Getenv(APIAccessTokenEnvKey)
	if token == "" {
		return nil, fmt.Errorf("environment variable %q is required", APIAccessTokenEnvKey)
	}
	secret := os.Getenv(APIAccessSecretEnvKey)
	if secret == "" {
		return nil, fmt.Errorf("environment variable %q is required", APIAccessSecretEnvKey)
	}
	return NewClient(token, secret), nil
}

func (c *Client) init() {
	if c.UserAgent == "" {
		c.UserAgent = DefaultUserAgent
	}
	if c.AcceptLanguage == "" {
		c.AcceptLanguage = DefaultAcceptLanguage
	}
	if c.CheckRetryFunc == nil {
		c.CheckRetryFunc = DefaultCheckRetryFunc
	}
	if c.RetryMax == 0 {
		c.RetryMax = DefaultRetryMax
	}
	if c.RetryWaitMin == 0 {
		c.RetryWaitMin = DefaultRetryWaitMin
	}
	if c.RetryWaitMax == 0 {
		c.RetryWaitMax = DefaultRetryWaitMax
	}
}

func (c *Client) httpClient() *retryablehttp.Client {
	return &retryablehttp.Client{
		HTTPClient:   c.HTTPClient,
		RetryWaitMin: c.RetryWaitMin,
		RetryWaitMax: c.RetryWaitMax,
		RetryMax:     c.RetryMax,
		CheckRetry:   c.CheckRetryFunc,
		Backoff:      retryablehttp.DefaultBackoff,
	}
}

// Do APIコール実施
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.init()

	// set headers
	req.SetBasicAuth(c.AccessToken, c.AccessTokenSecret)
	if req.Header.Get("Content-Type") == "" && req.Body != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	if c.Gzip && req.Header.Get("Accept-Encoding") == "" {
		req.Header.Add("Accept-Encoding", "gzip")
	}
	if req.Header.Get("X-Requested-With") == "" {
		req.Header.Add("X-Requested-With", "XMLHttpRequest")
	}
	if req.Header.Get("X-Sakura-Bigint-As-Int") == "" {
		req.Header.Add("X-Sakura-Bigint-As-Int", "1") //Use BigInt on resource ids.
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}
	if req.Header.Get("Accept-Language") == "" && c.AcceptLanguage != "" {
		req.Header.Add("Accept-Language", c.AcceptLanguage)
	}

	if c.RequestCustomizer != nil {
		if err := c.RequestCustomizer(req); err != nil {
			return nil, err
		}

	}

	request, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, err
	}

	client := c.httpClient()

	// API call
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	if c.Gzip && resp.Header.Get("Content-Encoding") == "gzip" {
		body, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body = body
	}

	return resp, err
}
