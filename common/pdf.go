package common

import (
	"context"
	"github.com/google/go-tika/tika"
	"io"
)

// PDFToHTML converts a PDF file to HTML.
func PDFToHTML(ctx context.Context, client *tika.Client, input io.Reader) (string, error) {
	htmlStr, err := client.Parse(ctx, input)
	if err != nil {
		return "", err
	}
	return htmlStr, nil
}

// PDFToText converts a PDF file to text.
func PDFToText(ctx context.Context, client *tika.Client, input io.Reader) (string, error) {
	htmlStr, err := PDFToHTML(ctx, client, input)
	if err != nil {
		return "", err
	}
	text, err := HTMLToText(htmlStr)
	if err != nil {
		return "", err
	}
	return text, nil
}
