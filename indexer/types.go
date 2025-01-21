package indexer

import "time"

type PostInfoResp RestResp[any]

type RestResp[T any] struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    *T     `json:"data,omitempty"`
}

type PostInfo struct {
	Id                int64         `json:"id"`
	Category          string        `json:"category"`
	State             string        `json:"state"`
	StateDesc         string        `json:"stateDesc"`
	AuditRejectReason interface{}   `json:"auditRejectReason"`
	Official          bool          `json:"official"`
	Top               bool          `json:"top"`
	Marrow            bool          `json:"marrow"`
	Title             string        `json:"title"`
	ContentType       string        `json:"contentType"`
	HtmlContent       string        `json:"htmlContent"`
	MarkdownContent   string        `json:"markdownContent"`
	HeadImg           string        `json:"headImg"`
	AuthorId          int64         `json:"authorId"`
	AuthorNickname    string        `json:"authorNickname"`
	AuthorAvatar      string        `json:"authorAvatar"`
	CreateAt          time.Time     `json:"createAt"`
	UpdateAt          time.Time     `json:"updateAt"`
	Views             int64         `json:"views"`
	Approvals         int64         `json:"approvals"`
	Comments          int64         `json:"comments"`
	OriginalAuthor    string        `json:"originalAuthor"`
	OriginalTitle     string        `json:"originalTitle"`
	OriginalUrl       string        `json:"originalUrl"`
	Directory         string        `json:"directory"`
	DifficultyScore   int64         `json:"difficultyScore"`
	Tags              []Tag         `json:"tags"`
	SolutionId        interface{}   `json:"solutionId"`
	Solution          interface{}   `json:"solution"`
	BelongWikis       []BelongWiki  `json:"belongWikis"`
	DynamicCategory   interface{}   `json:"dynamicCategory"`
	Draft             interface{}   `json:"draft"`
	JobCategoryId     interface{}   `json:"jobCategoryId"`
	CanComment        int64         `json:"canComment"`
	History           bool          `json:"history"`
	ContactText       string        `json:"contactText"`
	ContactPic        string        `json:"contactPic"`
	Attachments       []interface{} `json:"attachments"`
	References        []interface{} `json:"references"`
	BelongTeamContent bool          `json:"belongTeamContent"`
	TeamInfo          interface{}   `json:"teamInfo"`
}

type Tag struct {
	Id        int64       `json:"id"`
	GroupName string      `json:"groupName"`
	Name      string      `json:"name"`
	HeadImg   interface{} `json:"headImg"`
}

type BelongWiki struct {
	WikiId      int64       `json:"wikiId"`
	WikiName    string      `json:"wikiName"`
	WikiHeadImg string      `json:"wikiHeadImg"`
	NodeId      int64       `json:"nodeId"`
	Type        string      `json:"type"`
	Childs      interface{} `json:"childs"`
}
