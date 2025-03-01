package common

import (
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"strings"
)

// HTMLToText 将HTML字符串转换为纯文本
func HTMLToText(htmlStr string) (string, error) {
	// 解析HTML字符串
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return "", errors.Wrap(err, "failed to parse HTML")
	}

	return HTMLNodeToText(doc)
}

// HTMLNodeToText 将HTML节点转换为纯文本
func HTMLNodeToText(n *html.Node) (string, error) {
	var text strings.Builder
	// 递归遍历HTML节点树
	var traverse func(n *html.Node)
	traverse = func(n *html.Node) {
		// 如果节点类型是文本节点
		if n.Type == html.TextNode {
			// 去除前后空白字符并追加到文本构建器中
			text.WriteString(strings.TrimSpace(n.Data))
			// 如果文本不为空，追加一个空格
			if text.Len() > 0 && !strings.HasSuffix(text.String(), " ") {
				text.WriteByte(' ')
			}
		}
		// 递归遍历子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	// 从根节点开始遍历
	traverse(n)

	// 返回最终的纯文本
	return strings.TrimSpace(text.String()), nil
}
