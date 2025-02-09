package indexer

import (
	"encoding/json"
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

	return json.Marshal(IndexEntity{
		BaseEntity: BaseEntity{
			Id:              id,
			Type:            EntityTypeBbsPost,
			Title:           post.Title,
			TitleAnalyzed:   AnalyzeWhitespace(post.Title),
			Content:         content,
			ContentAnalyzed: AnalyzeWhitespace(content),
		},
		BbsPost: &post,
	})
}
