package xfilters

import "strings"

// SensitiveWords 简单敏感词表（生产可换 AC 自动机 / 词典服务）
var SensitiveWords = []string{"spam", "钓鱼", "phishing"}

// ContainsSensitive 检测是否包含敏感词
func ContainsSensitive(s string) bool {
	for _, w := range SensitiveWords {
		if strings.Contains(s, w) {
			return true
		}
	}
	return false
}
