package indexer

import (
	"fmt"
	"time"
)

// Time 自定义时间类型，是 time.Time 的别名
type Time time.Time

// 自定义时间格式
const customTimeFormat = "2006-01-02T15:04:05.000+00:00"

// MarshalJSON 实现 json.Marshaler 接口
func (t Time) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format(customTimeFormat))
	return []byte(stamp), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (t *Time) UnmarshalJSON(data []byte) error {
	str := string(data)[1 : len(data)-1]
	parsedTime, err := time.Parse(customTimeFormat, str)
	if err != nil {
		return err
	}
	*t = Time(parsedTime)
	return nil
}
