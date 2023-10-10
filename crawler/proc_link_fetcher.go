package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/odit-bit/invoker/internal/pipeline"
)

type URLGetter interface {
	Get(ctx context.Context, url string) (*http.Response, error)
}

type PrivateNetworkDetector interface {
	IsPrivate(host string) (bool, error)
}

var _ pipeline.Processor = (*linkFetcher)(nil)

// linkFetcher operates on payload values emitted by the input source and
// attempts to retrieve the contents of each link by sending out HTTP GET requests.
// The retrieved link web page contents are stored within the payload's RawContent field
// and made available to the following stages of the pipeline.
// for url that lead to non html , private network, or non 200 status code will be skipped with silent error
type linkFetcher struct {
	urlGetter   URLGetter
	netDetector PrivateNetworkDetector
}

func newLinkFetcher(urlGetter URLGetter, netDetector PrivateNetworkDetector) *linkFetcher {
	return &linkFetcher{
		urlGetter:   urlGetter,
		netDetector: netDetector,
	}
}

// Process implements pipeline.Processor.
func (lf *linkFetcher) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	// p should crawler's payload struct
	payload, _ := p.(*payload)

	pURL := payload.URL
	// Skip URLs that point to files that cannot contain html content.
	if exclusionRegex.MatchString(pURL) {
		// log.Println("err parse url:", pURL)
		return nil, nil
	}

	// Never crawl links in private networks (e.g. link-local addresses).
	// This is a security risk!
	private, err := lf.isPrivate(pURL)
	if private || err != nil {
		// log.Println("err private network:", private)
		return nil, nil
	}

	//get url within timeout otherwise skipped
	if err := contentFromURL(ctx, lf.urlGetter, payload); err != nil {
		// log.Println(err)
		return nil, nil
	}

	return payload, nil
}

// check does the url pointed to private ip address
func (lf *linkFetcher) isPrivate(urlString string) (bool, error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return false, nil
	}
	return lf.netDetector.IsPrivate(parsedURL.Hostname())

}

// get content of url pointed
func contentFromURL(ctx context.Context, getter URLGetter, payload *payload) error {
	// url Getter
	// held crawl link in expensive connection
	res, err := getter.Get(ctx, payload.URL)
	if err != nil {
		return fmt.Errorf("link fetcher: %v", err)
	}

	// http.Response should not nil, skipped not success code
	if res == nil || res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("link fetcher: status nok ok (200)")
	}

	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "html") {
		return fmt.Errorf("link fetcher: non html content-type:%v", contentType)
	}

	_, err = io.Copy(&payload.RawContent, res.Body)
	_ = res.Body.Close()
	if err != nil {
		return fmt.Errorf("link fetcher: %v", err)
	}

	return nil
}
