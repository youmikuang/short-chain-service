package tool

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Base62Encode 将 uint64 编码为 base62 字符串
func Base62Encode(num uint64) string {
	if num == 0 {
		return string(base62Chars[0])
	}
	var buf [12]byte
	i := len(buf)
	for num > 0 {
		i--
		buf[i] = base62Chars[num%62]
		num /= 62
	}
	return string(buf[i:])
}

// RandString 生成指定长度随机字符串（用于混淆/盐）
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = base62Chars[rand.Intn(len(base62Chars))]
	}
	return string(b)
}

// Sha256Hex 对字符串做 sha256 十六进制编码（API Key 入库哈希）
func Sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// ExtractDomain 从长链接中提取域名（用于黑名单校验）
func ExtractDomain(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	host := u.Hostname()
	if host == "" {
		return "", errors.New("empty host")
	}
	return host, nil
}

// NormalizeURL 归一化长链接，用于去重判断：去掉 fragment、小写化 scheme/host、
// 去掉仅含 "/" 的路径、按字典序重排 query 参数。使 https://baidu.com/ 与
// https://baidu.com、https://Baidu.com/?a=1&b=2 等视为同一链接。
// 解析失败则原样返回（仅去空格）。
func NormalizeURL(raw string) string {
	s := strings.TrimSpace(raw)
	u, err := url.Parse(s)
	if err != nil {
		return s
	}
	u.Fragment = ""
	if u.Scheme != "" {
		u.Scheme = strings.ToLower(u.Scheme)
	}
	if u.Host != "" {
		u.Host = strings.ToLower(u.Host)
	}
	if u.Path == "/" {
		u.Path = ""
	}
	u.RawQuery = u.Query().Encode() // Encode 会按 key 字典序排列
	return u.String()
}

// Snowflake 简易雪花 ID 生成器（Nginx 固定实例下 workerId 固定分配）
type Snowflake struct {
	mu       sync.Mutex
	epoch    int64
	workerID int64
	step     int64
}

// NewSnowflake workerID 取值范围 [0,1023]
func NewSnowflake(workerID int64) *Snowflake {
	return &Snowflake{
		epoch:    1700000000000, // 自定义起始时间戳(ms)
		workerID: workerID % 1024,
	}
}

// NextID 生成下一个 ID
func (s *Snowflake) NextID() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.step = (s.step + 1) & 0xFFF
	id := uint64((time.Now().UnixMilli()-s.epoch)<<22 | (s.workerID << 12) | s.step)
	return id
}
