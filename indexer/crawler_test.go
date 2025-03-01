package indexer

import (
	"testing"
)

func TestGetBbsPost(t *testing.T) {
	post, err := GetBbsPost(54068)
	if err != nil {
		t.Fatalf("GetBbsPost error: %v", err)
	}
	t.Logf("BbsPost: %+v", post)
}

func TestGetBbsPost2(t *testing.T) {
	const startId = 54050
	const endId = 54100
	for i := startId; i < endId; i++ {
		post, err := GetBbsPost(int64(i))
		if err != nil {
			t.Fatalf("GetBbsPost error: %v", err)
		}
		t.Logf("[%d] BbsPost: %+v", i, post)
		if post.Data != nil {
			t.Logf("Data: %+v", post.Data)
		}
	}
}

func TestGetAnnounce(t *testing.T) {
	announce, err := GetAnnounce(1784)
	if err != nil {
		t.Fatalf("GetAnnounce error: %v", err)
	}
	t.Logf("Announce: %+v", announce)
}
