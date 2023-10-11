package crawler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/odit-bit/invoker/internal/pipeline"
)

type URLGetter interface {
	Get(url string) (*http.Response, error)
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
		log.Printf("link fetcher url: %v\nerror: %v \n", pURL, "not containt html content")
		return nil, nil
	}

	// Never crawl links in private networks (e.g. link-local addresses).
	// This is a security risk!
	private, err := lf.isPrivate(pURL)
	if private || err != nil {
		log.Printf("link fetcher error: %v url: %v\n", err, pURL)
		return nil, nil
	}

	//get url within timeout otherwise skipped
	if err := contentFromURL(ctx, lf.urlGetter, payload); err != nil {
		log.Printf("link fetcher error: %v url: %v\n", err, pURL)
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
	res, err := getter.Get(payload.URL)
	if err != nil {
		return fmt.Errorf("http request: %v", err)
	}

	// http.Response should not nil, skipped not success code
	if res == nil || res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("http response status nok ok (%v)", res.StatusCode)
	}

	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "html") {
		return fmt.Errorf("http response: non html content-type:%v", contentType)
	}

	_, err = io.Copy(&payload.RawContent, res.Body)
	if err != nil {
		return fmt.Errorf("copy response body: %v", err)
	}

	err = res.Body.Close()
	if err != nil {
		return fmt.Errorf("close response body: %v", err)
	}

	// log.Println("link fetcher content type", contentType, "url:", payload.URL)
	return nil
}
