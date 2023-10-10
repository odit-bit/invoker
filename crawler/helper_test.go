package crawler

import "testing"

func Test_baseHref(t *testing.T) {
	href := `<base href="">`

	match := baseHrefRegex.FindStringSubmatch(href)

	url := trailingSlash(match[1])
	if url != "/" {
		t.Error("not matched ")
	}
}
