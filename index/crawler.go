package index

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/scutrobotlab/rm-search/common"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrStatusMethodNotAllowed = errors.New("status method not allowed")
	ErrStatusNotFound         = errors.New("status not found")
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

// GetAnnounce 获取公告信息
func GetAnnounce(id int64) (ret *Announce, err error) {
	url := fmt.Sprintf("https://www.robomaster.com/zh-CN/resource/pages/announcement/%d", id)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	bodyStr := string(body)
	if strings.Contains(bodyStr, "您访问的页面不存在") {
		return nil, ErrStatusNotFound
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrapf(err, "parse html")
	}

	mainContent, err := GetMainContent(doc)
	if err != nil {
		return nil, errors.Wrapf(err, "get main content")
	}
	var mainTitle *html.Node
	var mainDate *html.Node
	var mainContext *html.Node
	for c := mainContent.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			for _, attr := range c.Attr {
				switch attr.Val {
				case "main-title":
					mainTitle = c
				case "main-date":
					mainDate = c
				case "main-context":
					mainContext = c
				default:
					continue
				}
			}
		}
	}

	var title string
	if mainTitle != nil {
		title = mainTitle.FirstChild.Data
	}
	var date time.Time
	if mainDate != nil {
		for c := mainDate.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				for _, attr := range c.Attr {
					if attr.Key == "class" && attr.Val == "article-time" {
						dateStr := c.FirstChild.Data
						date, err = time.Parse("2006-01-02", dateStr)
					}
				}
			}
		}
	}
	var context string
	if mainContext != nil {
		var buf bytes.Buffer
		err = html.Render(&buf, mainContext)
		if err != nil {
			logrus.Errorf("render html: %v", err)
		} else {
			context = buf.String()
		}
	}
	var content string
	if mainContext != nil {
		content, err = common.HTMLNodeToText(mainContext)
		if err != nil {
			logrus.Errorf("html node to text: %v", err)
		}
	}
	attachments := ParseAttachment(mainContext)

	return &Announce{
		Id:          id,
		Title:       title,
		Date:        date,
		Context:     context,
		Content:     content,
		Attachments: attachments,
	}, nil
}

func GetMainContent(n *html.Node) (*html.Node, error) {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && attr.Val == "main-content" {
				return n, nil
			}
		}
	}

	// 递归处理子节点
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		context, err := GetMainContent(c)
		if err == nil {
			return context, nil
		}
	}

	return nil, fmt.Errorf("main context not found")
}

// ParseAttachment 解析附件
func ParseAttachment(n *html.Node) (ret []Attachment) {
	if n == nil {
		return []Attachment{}
	}

	var traverse func(n *html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.HasSuffix(attr.Val, ".pdf") {
					ret = append(ret, Attachment{
						Src:  attr.Val,
						Name: n.FirstChild.Data,
					})
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return ret
}
