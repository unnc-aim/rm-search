package common

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// MustMarshal 序列化
func MustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// CompareJSON 比较两个 JSON 并打印不一致的位置
func CompareJSON(json1, json2 []byte, ignorePaths []string) bool {
	var v1, v2 interface{}
	// 解析第一个 JSON 数据
	if err := json.Unmarshal(json1, &v1); err != nil {
		fmt.Printf("Error unmarshaling JSON 1: %v\n", err)
		return false
	}
	// 解析第二个 JSON 数据
	if err := json.Unmarshal(json2, &v2); err != nil {
		fmt.Printf("Error unmarshaling JSON 2: %v\n", err)
		return false
	}
	// 调用递归比较函数
	return compareValues(v1, v2, "", ignorePaths)
}

// compareValues 递归比较两个值
func compareValues(v1, v2 interface{}, path string, ignorePaths []string) bool {
	// 检查当前路径是否在忽略列表中
	if shouldIgnore(path, ignorePaths) {
		return true
	}

	t1 := reflect.TypeOf(v1)
	t2 := reflect.TypeOf(v2)

	// 检查类型是否相同
	if t1 != t2 {
		fmt.Printf("Type mismatch at path %s: %v vs %v\n", path, t1, t2)
		return false
	}

	switch v1 := v1.(type) {
	case map[string]interface{}:
		v2 := v2.(map[string]interface{})
		for k, val1 := range v1 {
			newPath := appendPath(path, k)
			if val2, ok := v2[k]; ok {
				if !compareValues(val1, val2, newPath, ignorePaths) {
					return false
				}
			} else {
				fmt.Printf("Key %s not found in JSON 2 at path %s\n", k, path)
				return false
			}
		}
		for k := range v2 {
			if !containsMapKey(v1, k) {
				fmt.Printf("Key %s not found in JSON 1 at path %s\n", k, path)
				return false
			}
		}
	case []interface{}:
		v2 := v2.([]interface{})
		if len(v1) != len(v2) {
			fmt.Printf("Array length mismatch at path %s: %d vs %d\n", path, len(v1), len(v2))
			return false
		}
		for i := range v1 {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			if !compareValues(v1[i], v2[i], newPath, ignorePaths) {
				return false
			}
		}
	default:
		if !reflect.DeepEqual(v1, v2) {
			fmt.Printf("Value mismatch at path %s: %v vs %v\n", path, v1, v2)
			return false
		}
	}
	return true
}

// shouldIgnore 检查当前路径是否需要忽略
func shouldIgnore(path string, ignorePaths []string) bool {
	for _, ignorePath := range ignorePaths {
		if path == ignorePath {
			return true
		}
	}
	return false
}

// appendPath 拼接路径
func appendPath(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

// containsMapKey 检查映射中是否包含某个键
func containsMapKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}
