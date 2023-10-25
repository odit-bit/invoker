package crawler

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	"github.com/odit-bit/pipeline"
)

var (
	// it only matches in line content of <title>content</title>
	// titleRegex = regexp.MustCompile(`(?i)<title.*?>(.*?)</title>`)

	titleRegex         = regexp.MustCompile(`(?is)<title.*?>(.*?)</title>`)
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
	payload := p.(*payload)
	lenP := payload.RawContent.Len()
	if lenP == 0 {
		return nil, fmt.Errorf("text extractor: length raw content is zero")
	}
	sanitizer := te.policyPool.Get().(*bluemonday.Policy)
	defer te.policyPool.Put(sanitizer)

	title, body := sanitizeBytes(sanitizer, &payload.RawContent)
	payload.Title, payload.TextContent = title, body

	if len(payload.TextContent) == 0 {
		return nil, nil
	}

	return payload, nil
}

// helper

func sanitizeString(sanitizer *bluemonday.Policy, buf *bytes.Buffer) (string, string) {
	// get <title> tag html and sub string
	// ex: ["<title> ..content.. </title>",  "..content..""]
	titleMatched := titleRegex.FindStringSubmatch(buf.String())

	var title string
	// log.Printf("DEBUG text extractor title matched content: %v", titleMatched)
	if len(titleMatched) == 2 {
		title = sanitizer.Sanitize(titleMatched[1])
		title = repeatedSpaceRegex.ReplaceAllString(title, " ")
		ok := isValidUTF8([]byte(title))
		title = strings.TrimSpace(title)
		if !ok {
			title = ""
		}
	}

	textContent := sanitizer.SanitizeReader(buf).String()
	textContent = repeatedSpaceRegex.ReplaceAllString(textContent, " ")
	textContent = strings.TrimSpace(textContent)
	ok := isValidUTF8([]byte(textContent))
	if !ok {
		textContent = ""
	}

	return title, textContent
}

var repeatSpaceBytes = []byte(" ")

func sanitizeBytes(sanitizer *bluemonday.Policy, buf *bytes.Buffer) (title, body []byte) {
	htmlBytes := buf.Bytes()

	matchedBytes := titleRegex.FindSubmatch(htmlBytes)
	if len(matchedBytes) == 2 {
		title = sanitizer.SanitizeBytes(matchedBytes[1])
		title = repeatedSpaceRegex.ReplaceAll(title, repeatSpaceBytes)
		title = bytes.TrimSpace(title)
		ok := isValidUTF8(title)
		if !ok {
			title = nil
		}
	}

	body = sanitizer.SanitizeReader(buf).Bytes()
	body = repeatedSpaceRegex.ReplaceAll(body, repeatSpaceBytes)
	body = bytes.TrimSpace(body)
	ok := isValidUTF8(body)
	if !ok {
		body = nil
	}

	return title, body

}

func isValidUTF8(input []byte) bool {
	for len(input) > 0 {
		r, size := utf8.DecodeRune(input)
		if r == utf8.RuneError && size == 1 {
			return false
		}
		input = input[size:]
	}
	return true
}
