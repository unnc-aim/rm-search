package common

import "testing"

func TestHTMLToText(t *testing.T) {
	htmlStr := "<p><span style=\"color: rgb(0, 0, 0); font-size: 14px;\"><em>完整运动：平移、旋转（必做）</em></span></p><p><span style=\"color: rgb(0, 0, 0); font-size: 14px;\"><em>[在此处上传视频]</em></span></p><p><span style=\"color: rgb(0, 0, 0); font-size: 14px;\"><em>连续发射50发弹丸，攻击距离5米处静止的与大装甲模块尺寸相同的目标，统计命中率并展示相关证明材料（选做）</em></span></p><p><em>[在此处上传视频]</em></p><p><span style=\"color: rgb(0, 0, 0); font-size: 14px;\"><em>自动识别并分别跟随平移、旋转装甲模块，连续发射30发弹丸击打装甲模块，统计并展示命中率（如复写纸痕迹）；同时展示装甲板识别的可视化程序运行效果（可参考考核细则图1）</em></span><em>（选做）</em></p><p><em>[在此处上传视频]</em></p><p><span style=\"color: rgb(0, 0, 0); font-size: 14px;\"><em>展示哨兵机器人在比赛场地中移动、定位、避障、路径规划的自动运行效果与可视化程序运行效果，其中程序运行效果的数据与展示的机器人实际运行相对应（可参考考核细则图2）</em></span><em>（选做）</em></p><p><em>[在此处上传视频]</em></p><p><span style=\"color: rgb(0, 0, 0); font-size: 14px;\"><em>展示哨兵机器人不同运行模式（如两点间巡逻/原地旋转/自动反击等）</em></span><em>（选做）</em></p><p><em>[在此处上传视频]</em></p>"
	text, err := HTMLToText(htmlStr)
	if err != nil {
		t.Fatalf("failed to convert HTML to text: %v", err)
	}
	t.Logf("text: %s", text)
}
