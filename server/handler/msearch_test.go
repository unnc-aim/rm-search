package handler

import "testing"

func Test_extractQuery(t *testing.T) {
	type args struct {
		reqBody []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test extractQuery",
			args: args{
				reqBody: []byte(`{"index":"rm-search"}
{"aggs":{"categories.lvl0":{"terms":{"field":"category_lvl0.keyword","size":20}},"college_name":{"terms":{"field":"college_name.keyword","size":20}},"source":{"terms":{"field":"source.keyword","size":20}}},"query":{"function_score":{"query":{"bool":{"filter":[],"must":{"bool":{"should":[{"bool":{"should":[{"multi_match":{"query":"华南理工","fields":["title^5","content"],"fuzziness":"AUTO:4,8"}},{"multi_match":{"query":"华南理工","fields":["title^2.5","content"],"type":"bool_prefix"}}]}},{"multi_match":{"query":"华南理工","type":"phrase","fields":["title^10","content"]}}]}}}},"functions":[{"filter":{"range":{"create_time":{"gte":"now-1095d"}}},"gauss":{"create_time":{"origin":"now/d","scale":"1095d","offset":"30d","decay":0.5}}},{"filter":{"range":{"create_time":{"lt":"now-1095d"}}},"weight":0.5}],"score_mode":"first","boost_mode":"multiply"}},"size":10,"from":0,"_source":{"includes":["id","source","title","content","image","url","author_nickname","author_avatar","create_time"]},"highlight":{"pre_tags":["<em>"],"post_tags":["</em>"],"fields":{"title":{"number_of_fragments":0},"content":{"number_of_fragments":5,"fragment_size":100}}},"sort":{"_score":"desc"}}`),
			},
			want: "华南理工",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractQuery(tt.args.reqBody); got != tt.want {
				t.Errorf("extractQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
