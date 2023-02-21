package cachehttp

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type httpGetter struct {
	baseURL string
}

// GetKeyFromGetter 从getter中获取key
func (h *httpGetter) GetKeyFromGetter(group string, key string) ([]byte, error) {
	// 拼接请求group和key的url
	u := fmt.Sprintf(
		"http://%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	// 发送请求
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	// 获取对应key的其他节点的响应
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}
