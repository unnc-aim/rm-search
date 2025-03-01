package indexer

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/database/model"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
	"time"
)

const (
	BBSBaseURL      = "https://bbs.robomaster.com"
	AnnounceBaseURL = "https://www.robomaster.com/zh-CN/resource/pages/announcement"
)

var ErrBbsPostCannotIndex = errors.New("bbs post cannot be indexed")

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

	var url string
	switch post.Category {
	case "ARTICLE":
		url = fmt.Sprintf("%s/article/%d", BBSBaseURL, post.Id)
	case "FAQ":
		url = fmt.Sprintf("%s/faq/%d", BBSBaseURL, post.Id)
	case "WIKI":
		if len(post.BelongWikis) > 0 {
			url = fmt.Sprintf("%s/wiki/%d/%d", BBSBaseURL, post.BelongWikis[0].WikiId, post.Id)
		}
	default:
		logrus.Debugf("unknown category: %s, id: %d", post.Category, post.Id)
	}
	if url == "" {
		return nil, ErrBbsPostCannotIndex
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

	var authorNickname string
	var authorAvatar string
	if post.AuthorNickname != nil {
		authorNickname = *post.AuthorNickname
	}
	authorAvatar = post.AuthorAvatar

	createTime := time.Time(post.CreateAt).UnixMilli()
	updateTime := time.Time(post.UpdateAt).UnixMilli()

	return json.Marshal(IndexEntity{
		BaseEntity: BaseEntity{
			Id:             id,
			Type:           EntityTypeBbsPost,
			Title:          title,
			Content:        content,
			Image:          image,
			Url:            url,
			Season:         season,
			CategoryLvl0:   categoryLvl0,
			CategoryLvl1:   categoryLvl1,
			CollegeName:    collegeName,
			AuthorNickname: authorNickname,
			AuthorAvatar:   authorAvatar,
			CreateTime:     createTime,
			UpdateTime:     updateTime,
		},
		BbsPost: &post,
	})
}

// ConvertAnnounce 转换公告信息
func ConvertAnnounce(id string, src model.Announce) ([]byte, error) {
	var attachments []Attachment
	err := json.Unmarshal([]byte(src.Attachments), &attachments)
	if err != nil {
		logrus.Errorf("failed to unmarshal attachments: %v", err)
	}
	announce := Announce{
		Id:          src.ID,
		Title:       src.Title,
		Date:        src.Date,
		Context:     src.Context,
		Content:     src.Content,
		Attachments: attachments,
	}

	title := strings.TrimSpace(announce.Title)
	content := announce.Content
	url := fmt.Sprintf("%s/%d", AnnounceBaseURL, announce.Id)
	date := announce.Date.UnixMilli()

	return json.Marshal(IndexEntity{
		BaseEntity: BaseEntity{
			Id:             id,
			Type:           EntityTypeAnnounce,
			Title:          title,
			Content:        content,
			Image:          "",
			Url:            url,
			Season:         "",
			CategoryLvl0:   nil,
			CategoryLvl1:   nil,
			CollegeName:    nil,
			AuthorNickname: "",
			AuthorAvatar:   "",
			CreateTime:     date,
			UpdateTime:     date,
		},
		Announce: &announce,
	})
}
