package indexer

import (
	"github.com/go-ego/gse"
	"strings"
)

var seg = gse.Segmenter{}

func Analyze(text string) []string {
	return seg.CutSearch(text, true)
}

func AnalyzeWhitespace(text string) string {
	analyze := Analyze(text)
	return strings.Join(analyze, " ")
}

func init() {
	err := seg.LoadDict()
	if err != nil {
		panic(err)
	}
}
