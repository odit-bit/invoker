package xhttpclient

import (
	"context"
	"net/http"
	"time"
)

var DefaultGetter = &UrlGetter{
	cli:     &http.Client{},
	timeout: 0,
}

type UrlGetter struct {
	cli     *http.Client
	timeout time.Duration
}

func (ug *UrlGetter) SetTimeout(ctxTimeout time.Duration) {
	ug.cli.Timeout = ctxTimeout
}

func (ug *UrlGetter) WithNoRedirect() {
	ug.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (ug *UrlGetter) Get(url string) (*http.Response, error) {
	ctx := context.TODO()
	if ug.timeout > time.Second {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), ug.timeout)
		ctx = ctxTimeout
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ug.cli.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
