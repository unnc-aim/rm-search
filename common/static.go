package common

import (
	"strings"
)

var SourceToSelf = map[string]string{
	"https://rm-static.djicdn.com": "/rm-static.djicdn.com",
}

// RedirectStatic 将第三方静态资源重定向到自己的 CDN
func RedirectStatic(url string) string {
	for k, v := range SourceToSelf {
		url = strings.ReplaceAll(url, k, v)
	}
	return url
}

// RedirectStaticIfProd 将第三方静态资源重定向到自己的 CDN
func RedirectStaticIfProd(url string) string {
	if IsProd() {
		return RedirectStatic(url)
	}
	return url
}

// GetStaticSource 获取静态资源的源地址
func GetStaticSource(url string) string {
	for k, v := range SourceToSelf {
		url = strings.ReplaceAll(url, v, k)
	}
	return url
}
