package indexer

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

var (
	ErrStatusMethodNotAllowed = errors.New("status code: 405")
)

// GetBbsPost 获取帖子信息
func GetBbsPost(id int64) (ret *BbsPostResp, err error) {
	url := fmt.Sprintf("https://bbs.robomaster.com/developers-server/rest/posts/info/%d", id)
	resp, err := http.Post(url, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 405 {
			return nil, ErrStatusMethodNotAllowed
		}
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
