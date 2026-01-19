package index

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/scutrobotlab/rm-search/database/model"
	"github.com/sirupsen/logrus"
)

const (
	BBSBaseURL      = "https://bbs.robomaster.com"
	AnnounceBaseURL = "https://www.robomaster.com/zh-CN/resource/pages/announcement"
)

var ErrBbsPostCannotIndex = errors.New("bbs post cannot be indexed")

func getEntityId(entityType EntityType, id any) string {
	return fmt.Sprintf("%s-%v", entityType, id)
}

// ConvertBbsPost 转换 BbsPost 信息
func ConvertBbsPost(m *model.BbsPost) (*Entity, error) {
	id := getEntityId(EntityTypeBbsPost, m.ID)

	var post BbsPost
	if err := json.Unmarshal([]byte(m.Data), &post); err != nil {
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

	var source string
	var url string
	switch post.Category {
	case "ARTICLE":
		source = EntitySourceBbsPostArticle
		url = fmt.Sprintf("%s/article/%d", BBSBaseURL, post.Id)
	case "FAQ":
		source = EntitySourceBbsPostFAQ
		url = fmt.Sprintf("%s/faq/%d", BBSBaseURL, post.Id)
	case "WIKI":
		source = EntitySourceBbsPostWiki
		if len(post.BelongWikis) > 0 {
			url = fmt.Sprintf("%s/wiki/%d/%d", BBSBaseURL, post.BelongWikis[0].WikiId, post.Id)
		}
	default:
		logrus.Errorf("unknown category: %s, id: %d", post.Category, post.Id)
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
	authorAvatar = common.RedirectStaticIfProd(post.AuthorAvatar)

	createTime := time.Time(post.CreateAt).UnixMilli()
	updateTime := time.Time(post.UpdateAt).UnixMilli()

	return &Entity{
		BaseEntity: BaseEntity{
			Id:             id,
			Type:           EntityTypeBbsPost,
			Source:         source,
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
	}, nil
}

// ConvertAnnounce 转换公告信息
func ConvertAnnounce(src *model.Announce) (*Entity, error) {
	id := getEntityId(EntityTypeAnnounce, src.ID)

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

	return &Entity{
		BaseEntity: BaseEntity{
			Id:             id,
			Type:           EntityTypeAnnounce,
			Source:         EntitySourceAnnounce,
			Title:          title,
			Content:        content,
			Image:          "",
			Url:            url,
			Season:         "",
			CategoryLvl0:   []string{"官方信息"},
			CategoryLvl1:   []string{"官方信息 > 公告"},
			CollegeName:    nil,
			AuthorNickname: "RoboMaster",
			AuthorAvatar:   "/robomaster-10th.webp",
			CreateTime:     date,
			UpdateTime:     date,
		},
	}, nil
}

// ConvertAttachment 转换附件信息
func ConvertAttachment(src *model.Attachment) (*Entity, error) {
	id := getEntityId(EntityTypeAttachment, src.ID)

	var source string
	switch src.Type {
	case common.ContentTypePDF:
		source = EntitySourceAttachmentPDF
	default:
		logrus.Errorf("unknown attachment type: %s", src.Type)
	}

	return &Entity{BaseEntity: BaseEntity{
		Id:             id,
		Type:           EntityTypeAttachment,
		Source:         source,
		Title:          src.Name,
		Content:        src.Content,
		Image:          "/pdf-file-svgrepo-com.svg",
		Url:            src.URL,
		Season:         "",
		CategoryLvl0:   []string{"官方信息"},
		CategoryLvl1:   []string{"官方信息 > 附件"},
		CollegeName:    nil,
		AuthorNickname: "RoboMaster",
		AuthorAvatar:   "/robomaster-10th.webp",
		CreateTime:     src.LastModified,
		UpdateTime:     src.LastModified,
	}}, nil
}
