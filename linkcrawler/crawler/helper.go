package crawler

import "regexp"

var (
	exclusionRegex = regexp.MustCompile(`(?i)\.(?:jpg|jpeg|png|gif|ico|css|js)$`)
	baseHrefRegex  = regexp.MustCompile(`(?i)<base.*?href\s*?=\s*?"(.*?)\s*?"`)
	findLinkRegex  = regexp.MustCompile(`(?i)<a.*?href\s*?=\s*?"\s*?(.*?)\s*?".*?>`)
	nofollowRegex  = regexp.MustCompile(`(?i)rel\s*?=\s*?"?nofollow"?`)
)
