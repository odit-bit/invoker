package crawler

import (
	"context"
	"html"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
	"github.com/odit-bit/invoker/internal/pipeline"
)

var (
	titleRegex         = regexp.MustCompile(`(?i)<title.*?>(.*?)</title>`)
	repeatedSpaceRegex = regexp.MustCompile(`\s+`)
)

var _ pipeline.Processor = (*textExtractor)(nil)

type textExtractor struct {
	// sanitize html
	policyPool sync.Pool
}

func newTextExtractor() *textExtractor {
	return &textExtractor{
		policyPool: sync.Pool{
			New: func() any {
				return bluemonday.StrictPolicy()
			},
		},
	}
}

// Process implements pipeline.Processor.
func (te *textExtractor) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload, ok := p.(*payload)
	if !ok {
		// log.Println("proc text extractor:", payload)
		// return nil, fmt.Errorf("text extractor :not crawler payload")
		return nil, nil
	}

	sanitizer := te.policyPool.Get().(*bluemonday.Policy)

	titleMatch := titleRegex.FindStringSubmatch(payload.RawContent.String())
	if len(titleMatch) == 2 {
		payload.Title = sanitizeTitle(sanitizer, titleMatch[1])
	}

	payload.TextContent = sanitizeContent(sanitizer, &payload.RawContent)

	te.policyPool.Put(sanitizer)
	return payload, nil
}

// helper

// sanit
func sanitizeTitle(sanitizer *bluemonday.Policy, titleMatch string) string {

	// sanitize the rawContent (html)
	t := sanitizer.Sanitize(titleMatch)

	// replace repeated space
	t2 := repeatedSpaceRegex.ReplaceAllString(t, " ")

	//replace non-utf8

	//
	t3 := html.UnescapeString(t2)
	t4 := strings.TrimSpace(t3)
	t5 := cleanText(t4)
	return t5
}

func sanitizeContent(sanitizer *bluemonday.Policy, r io.Reader) string {

	// sanitize the rawContent (html)
	t := sanitizer.SanitizeReader(r).String()

	// replace repeated space
	t2 := repeatedSpaceRegex.ReplaceAllString(t, " ")

	//
	t3 := html.UnescapeString(t2)

	t4 := cleanText(t3)

	t5 := strings.TrimSpace(t4)

	return t5
}

func cleanText(text string) string {
	// This function removes non-UTF-8 characters from the input text
	cleanedText := make([]rune, 0, len(text))
	for _, r := range text {
		if r == 0xFFFD { // Replace invalid UTF-8 characters
			continue
		}
		cleanedText = append(cleanedText, r)
	}
	return string(cleanedText)
}
