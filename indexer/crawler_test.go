package indexer

import (
	"testing"
)

func TestGetPostInfo(t *testing.T) {
	post, err := GetPostInfo(54068)
	if err != nil {
		t.Fatalf("GetPostInfo error: %v", err)
	}
	t.Logf("PostInfo: %+v", post)
}

func TestGetPostInfo2(t *testing.T) {
	const startId = 54050
	const endId = 54100
	for i := startId; i < endId; i++ {
		post, err := GetPostInfo(int64(i))
		if err != nil {
			t.Fatalf("GetPostInfo error: %v", err)
		}
		t.Logf("[%d] PostInfo: %+v", i, post)
		if post.Data != nil {
			t.Logf("Data: %+v", post.Data)
		}
	}
}
