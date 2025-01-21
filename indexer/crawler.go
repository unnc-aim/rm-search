package indexer

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetPostInfo 获取帖子信息
func GetPostInfo(id int64) (ret *PostInfoResp, err error) {
	url := fmt.Sprintf("https://bbs.robomaster.com/developers-server/rest/posts/info/%d", id)
	resp, err := http.Post(url, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
