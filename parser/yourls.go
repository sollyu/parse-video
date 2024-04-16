package parser

import (
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"net/url"
	"os"
)

func ShortUrl(longUrl string, title string) string {
	yourlsUrl := os.Getenv("YOURLS_URL")
	yourlsSignature := os.Getenv("YOURLS_SIGNATURE")
	if yourlsUrl == "" || yourlsSignature == "" {
		return longUrl
	}

	// 构建请求连接
	u, err := url.Parse(yourlsUrl)
	if err != nil {
		return longUrl
	}
	q := u.Query()
	q.Set("signature", yourlsSignature)
	q.Set("action", "shorturl")
	q.Set("url", longUrl)
	q.Set("title", title)
	q.Set("format", "json")
	u.RawQuery = q.Encode()
	u.Path = "/yourls-api.php"
	println(u.String())

	// 发送请求
	_client := resty.New()
	res, err := _client.R().
		SetHeader(HttpHeaderUserAgent, DefaultUserAgent).
		Get(u.String())
	if err != nil {
		return longUrl
	}

	// 解析返回结果
	_res := gjson.Parse(res.String())
	println(_res.String())
	switch _res.Get("statusCode").Int() {
	case 200:
		return _res.Get("shorturl").String()
	case 400:
		return _res.Get("shorturl").String()
	default:
		return longUrl
	}
}
