package utils

import (
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

func RemoveHtmlTags(s string, doRet bool) (string, error) {
	if doRet {
		s = regexp.MustCompile(`<br\s*/>`).ReplaceAllString(s, "\n")
		s = regexp.MustCompile(`</\s*p>\s*<p>`).ReplaceAllString(s, "\n\n")
	} else {
		s = regexp.MustCompile(`<br\s*/>`).ReplaceAllString(s, " ")
		s = regexp.MustCompile(`</\s*p>\s*<p>`).ReplaceAllString(s, " ")
	}
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return "", err
	}
	var extract func(n *html.Node, sb *strings.Builder) *strings.Builder
	extract = func(n *html.Node, sb *strings.Builder) *strings.Builder {
		if n.Type == html.TextNode {
			sb.WriteString(html.UnescapeString(n.Data))
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c, sb)
		}

		return sb
	}
	var sb strings.Builder
	return extract(doc, &sb).String(), nil
}

func RemoveHtmlTagsWithRet(s string) (string, error) {
	return RemoveHtmlTags(s, true)
}
