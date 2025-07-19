package common

type WordCloudItem struct {
	Word  string `json:"word" gorm:"column:query"`
	Count int64  `json:"count"`
}
