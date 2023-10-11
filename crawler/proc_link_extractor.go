package crawler

import (
	"context"
	"log"
	"net/url"

	"github.com/odit-bit/invoker/internal/pipeline"
)

var _ pipeline.Processor = (*linkExtractor)(nil)

type linkExtractor struct {
	pnd PrivateNetworkDetector
}

func newLinkExtractor(pnd PrivateNetworkDetector) *linkExtractor {
	return &linkExtractor{
		pnd: pnd,
	}
}

// Process implements pipeline.Processor.
func (le *linkExtractor) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload, _ := p.(*payload)
	relTo, err := url.Parse(payload.URL)
	if err != nil {
		return nil, err
	}

	//search <base> tag n resolve to absolute url
	// <base href="XXX">
	content := payload.RawContent.String()
	baseMatch := baseHrefRegex.FindStringSubmatch(content)
	if len(baseMatch) == 2 {
		base := resolveURL(relTo, trailingSlash(baseMatch[1]))
		if base != nil {
			relTo = base
		}
	}

	//find unique set of link
	seenMap := make(map[string]struct{})
	for _, match := range findLinkRegex.FindAllStringSubmatch(content, -1) {
		link := resolveURL(relTo, match[1])
		if !le.retainLink(relTo.Hostname(), link) {
			continue
		}

		// Truncate anchors and drop duplicates
		link.Fragment = ""
		linkStr := link.String()
		if _, seen := seenMap[linkStr]; seen {
			continue
		}

		// Skip URLs that point to files that cannot contain html content.
		if exclusionRegex.MatchString(linkStr) {
			continue
		}

		seenMap[linkStr] = struct{}{}
		if nofollowRegex.MatchString(match[0]) {
			payload.NoFollowLinks = append(payload.NoFollowLinks, linkStr)
		} else {
			payload.Links = append(payload.Links, linkStr)
		}
	}

	lenP := payload.RawContent.Len()
	if lenP == 0 {
		log.Printf("link extractor: len raw content is zero")
	}
	return payload, nil

}

func (le *linkExtractor) retainLink(srcHost string, link *url.URL) bool {
	// Skip links that could not be resolved
	if link == nil {
		return false
	}

	// Skip links with non http(s) schemes
	if link.Scheme != "http" && link.Scheme != "https" {
		return false
	}

	// Keep links to the same host
	if link.Hostname() == srcHost {
		return true
	}

	// Skip links that resolve to private networks
	if isPrivate, err := le.pnd.IsPrivate(link.Host); err != nil || isPrivate {
		return false
	}

	return true
}

func trailingSlash(s string) string {
	if s == "" {
		return "/"
	}

	if s[len(s)-1] != '/' {
		return s + "/"
	}
	return s
}

func resolveURL(relTo *url.URL, target string) *url.URL {
	tLen := len(target)
	if tLen == 0 {
		return nil
	}

	if tLen >= 1 && target[0] == '/' {
		if tLen >= 2 && target[1] == '/' {
			target = relTo.Scheme + ":" + target
		}
	}

	if targetURL, err := url.Parse(target); err == nil {
		return relTo.ResolveReference(targetURL)
	}

	return nil
}
