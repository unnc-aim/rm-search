package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var ColumnNames = []string{
	"id",
	"remote_ip",
	"user_agent",
	"request_body",
	"request_length",
	"query",
	"status",
	"response_body",
	"response_length",
	"latency",
	"create_time",
	"update_time",
}

func main() {
	// 打开文件
	file, err := os.Open("recover/source.sql")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 创建正则表达式来匹配插入语句
	insertRegex := regexp.MustCompile("### INSERT INTO `rm_search`.`search_log`")

	// 创建扫描器来逐行读取文件
	scanner := bufio.NewScanner(file)
	// Increase the buffer size
	const maxCapacity = 1024 * 1024 * 1024 // 1GB
	buf := make([]byte, 0, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var allParams [][]string

	for scanner.Scan() {
		line := scanner.Text()
		if insertRegex.MatchString(line) {
			var params []string
			// 跳过第一行 ### SET
			if scanner.Scan() {
				// 读取接下来的 12 行
				for i := 0; i < 12; i++ {
					if scanner.Scan() {
						paramLine := scanner.Text()
						// 提取参数值
						parts := strings.SplitN(paramLine, "=", 2)
						if len(parts) == 2 {
							value := strings.TrimSpace(parts[1])
							if i == 0 {
								value = "NULL"
							}
							if i == 10 || i == 11 {
								// 解析时间 1741459042.049
								// 转成 MySQL 的时间戳 2022-12-31 23:59:59.999
								seconds, _ := strconv.ParseInt(strings.Split(value, ".")[0], 10, 64)
								millis, _ := strconv.ParseInt(strings.Split(value, ".")[1], 10, 64)
								t := time.Unix(seconds, millis*1e6)
								value = t.Format("2006-01-02 15:04:05.999")
								value = fmt.Sprintf("'%s'", value)
							}
							params = append(params, value)
						}
					}
				}
				allParams = append(allParams, params)
			}
		}
	}

	outputFile, err := os.Create(fmt.Sprintf("recover/output_%s.sql", time.Now().Format("20060102_150405")))
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}

	sql := fmt.Sprintf("INSERT INTO `rm_search`.`search_log` VALUES \n")
	_, err = outputFile.WriteString(sql)
	if err != nil {
		fmt.Println("Error writing to output file:", err)
		return
	}

	// 输出所有提取的参数
	var values []string
	for i, params := range allParams {
		for j, param := range params {
			if len(param) > 50 {
				param = param[:50] + "..." + param[len(param)-1:]
				params[j] = param
			}
			fmt.Printf("[%d] %s = %s\n", i, ColumnNames[j], param)
		}
		fmt.Println()

		values = append(values, fmt.Sprintf("(%s)", strings.Join(params, ", ")))
	}

	_, err = outputFile.WriteString(strings.Join(values, ",\n") + ";\n")
	if err != nil {
		fmt.Println("Error writing to output file:", err)
		return
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}
