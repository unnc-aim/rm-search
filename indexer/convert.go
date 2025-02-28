package indexer

import (
	"encoding/json"
	"fmt"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

// ConvertBbsPost 转换 BbsPost 信息
func ConvertBbsPost(id string, src []byte) ([]byte, error) {
	var post BbsPost
	if err := json.Unmarshal(src, &post); err != nil {
		return nil, err
	}
	title := strings.TrimSpace(post.Title)

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

	var image string
	if post.HeadImg != nil && *post.HeadImg != "" {
		type HeadImg []struct {
			Alt string `json:"alt"`
			Url string `json:"url"`
		}
		var headImg HeadImg
		if err := json.Unmarshal([]byte(*post.HeadImg), &headImg); err != nil {
			logrus.Debugf("failed to unmarshal head image: %s, err: %v", *post.HeadImg, err)
		}
		if len(headImg) > 0 {
			image = headImg[0].Url
		}
	}

	// TODO: 从其他字段中提取赛季信息
	var season string
	regex := regexp.MustCompile(`RM(201[4-9]|202[0-6])`)
	matches := regex.FindAllString(strings.ToUpper(title), -1)
	if len(matches) > 0 {
		season = matches[0]
	} else {
		season = "未分类"
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

	collegeName := make([]string, 0)
	if post.TeamInfo != nil {
		collegeName = append(collegeName, strings.Split(post.TeamInfo.CollegeName, ";")...)
	} else {
		collegeName = ExtractCollegeName(title)
		//if len(collegeName) > 0 {
		//	logrus.Debugf("extracted college name %s from title %s", collegeName, title)
		//}
	}
	if len(collegeName) == 0 {
		collegeName = []string{"未分类"}
	}

	return json.Marshal(IndexEntity{
		BaseEntity: BaseEntity{
			Id:           id,
			Type:         EntityTypeBbsPost,
			Title:        title,
			Content:      content,
			Image:        image,
			Season:       season,
			CategoryLvl0: categoryLvl0,
			CategoryLvl1: categoryLvl1,
			CollegeName:  collegeName,
		},
		BbsPost: &post,
	})
}
