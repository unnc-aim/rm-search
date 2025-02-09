package common

import "github.com/russross/blackfriday/v2"

// MarkdownToText converts markdown to plain text
func MarkdownToText(markdown string) (string, error) {
	html := blackfriday.Run([]byte(markdown))
	return HTMLToText(string(html))
}

// MarkdownToHTML converts markdown to HTML
func MarkdownToHTML(markdown string) (string, error) {
	html := blackfriday.Run([]byte(markdown))
	return string(html), nil
}
