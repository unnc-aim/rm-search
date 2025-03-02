package index

import (
	"reflect"
	"testing"
)

func TestExtractCollegeName(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Test1",
			args: args{
				text: "我来自华南理工大学",
			},
			want: []string{"华南理工大学"},
		},
		{
			name: "Test2",
			args: args{
				text: "本次开源项目出自哈尔滨工业大学（深圳）南工骁鹰战队，作品仅用于技术交流，未经作者允许，不得作任何商业用途。",
			},
			want: []string{"哈尔滨工业大学（深圳）"},
		},
		{
			name: "Test3",
			args: args{
				text: "各位机甲大师大家好，这里给大家送上哈尔滨工业大学（威海）HERO战队的今年的赛季规划文档，今年我们的团队整体构架和制度有了很大的完善，同时也将各项文件都打包在附件里面一同送给大家。",
			},
			want: []string{"哈尔滨工业大学（威海）"},
		},
		{
			name: "Test4",
			args: args{
				text: "合肥工业大学(宣城校区)是一所以工为主，工、理、管、文、法、经、教育、艺术等多学科协调发展的全日制本科高校。",
			},
			want: []string{"合肥工业大学（宣城校区）"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractCollegeName(tt.args.text); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractCollegeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
