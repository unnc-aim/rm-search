package indexer

import (
	"testing"
)

func TestAnalyze(t *testing.T) {
	text := "RM2024华南理工大学华南虎哨兵机械结构设计报告"
	words := Analyze(text)
	t.Log(words)
}
