package index

import (
	_ "embed"
	"fmt"
	"github.com/go-ego/gse"
	"github.com/sirupsen/logrus"
	"strings"
	"unicode/utf8"
)

var Seg gse.Segmenter

//go:embed dict/college_name.txt
var collegeNameDict string

func init() {
	LoadDict()
}

func LoadDict() {
	Seg = gse.Segmenter{}

	err := Seg.LoadDictEmbed("zh")
	if err != nil {
		panic(err)
	}
	logrus.Info("Load dict zh success")

	err = Seg.LoadDictStr(collegeNameDict)
	if err != nil {
		panic(err)
	}
	logrus.Info("Load dict college_name success")
}

var schoolKeywords = []string{"大学", "学院"}

// ExtractCollegeName 提取学校名称
func ExtractCollegeName(text string) []string {
	// 替换全角括号
	text = strings.ReplaceAll(text, "(", "（")
	text = strings.ReplaceAll(text, ")", "）")

	// 分词
	cut := Seg.Cut(text, true)

	var ret []string
	for i, seg := range cut {
		// 过滤掉长度小于3的片段
		if utf8.RuneCountInString(seg) < 3 {
			continue
		}
		for _, word := range schoolKeywords {
			// 包含大学的片段和后缀
			if strings.Contains(seg, word) {
				suffix := strings.TrimPrefix(seg, word)
				// 如果包含括号
				if strings.Contains(suffix, "（") && strings.Contains(suffix, "）") {
					ret = append(ret, seg)
					continue
				}
			}
			// 以大学结尾的片段
			if strings.HasSuffix(seg, word) {
				collegeName := seg
				// 如果还有超过3个片段
				if i+3 < len(cut) {
					// 补充校区名称
					if cut[i+1] == "（" && cut[i+3] == "）" {
						collegeName = fmt.Sprintf("%s（%s）", seg, cut[i+2])
					}
				}
				ret = append(ret, collegeName)
				continue
			}
		}
	}

	return ret
}
