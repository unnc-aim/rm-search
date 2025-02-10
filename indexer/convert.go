package indexer

import (
	"encoding/json"
	"fmt"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/sirupsen/logrus"
)

// ConvertBbsPost 转换 BbsPost 信息
func ConvertBbsPost(id string, src []byte) ([]byte, error) {
	var post BbsPost
	if err := json.Unmarshal(src, &post); err != nil {
		return nil, err
	}

	var content string
	var err error
	switch post.ContentType {
	case common.BbsPostContentTypeHTML:
		content, err = common.HTMLToText(post.HtmlContent)
		if err != nil {
			logrus.Errorf("failed to convert HTML to text: %v", err)
		}
	case common.BbsPostContentTypeMarkdown:
		content, err = common.MarkdownToText(post.MarkdownContent)
		if err != nil {
			logrus.Errorf("failed to convert markdown to text: %v", err)
		}
	default:
		logrus.Errorf("unknown content type: %s", post.ContentType)
	}

	categoryLvl0 := make([]string, 0)
	categoryLvl1 := make([]string, 0)
	if len(post.Tags) == 0 {
		categoryLvl0 = []string{"未分类"}
		categoryLvl1 = []string{"未分类 > 未分类"}
	} else {
		for _, tag := range post.Tags {
			categoryLvl0 = append(categoryLvl0, tag.GroupName)
			categoryLvl1 = append(categoryLvl1, fmt.Sprintf("%s > %s", tag.GroupName, tag.Name))
		}
	}

	return json.Marshal(IndexEntity{
		BaseEntity: BaseEntity{
			Id:           id,
			Type:         EntityTypeBbsPost,
			Title:        post.Title,
			Content:      content,
			CategoryLvl0: categoryLvl0,
			CategoryLvl1: categoryLvl1,
		},
		BbsPost: &post,
	})
}
