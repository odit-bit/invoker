package xhttpclient

import (
	"context"
	"log"
	"net/http"
	"time"
)

type UrlGetter struct {
	cli     *http.Client
	Timeout time.Duration
}

// will create default http client with default timeout
func NewUrlGetter() *UrlGetter {

	transport := &http.Transport{
		// TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Use this if you want to skip certificate verification
	}

	ug := UrlGetter{
		cli:     &http.Client{Transport: transport},
		Timeout: 30,
	}
	return &ug
}

// will create default http client with given timeout
func NewUrlGetterWithTimeout(timeout time.Duration) *UrlGetter {

	if timeout == 0 {
		log.Fatal("timeout cannot 0")
	}

	transport := &http.Transport{
		// TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // Use this if you want to skip certificate verification
	}

	ug := UrlGetter{
		cli:     &http.Client{Transport: transport},
		Timeout: timeout,
	}
	return &ug
}

func (ug *UrlGetter) Get(ctx context.Context, url string) (*http.Response, error) {
	reqCtx, cancel := context.WithTimeout(ctx, ug.Timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := ug.cli.Do(req)
	return res, err
}
