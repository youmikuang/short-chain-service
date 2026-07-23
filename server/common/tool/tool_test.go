package tool

import (
	"strings"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"trailing slash removed", "https://baidu.com/", "https://baidu.com"},
		{"lowercase scheme/host", "HTTPS://Baidu.COM/a", "https://baidu.com/a"},
		{"query reordered + root path dropped", "https://baidu.com/?b=2&a=1", "https://baidu.com?a=1&b=2"},
		{"fragment dropped", "https://baidu.com/x#frag", "https://baidu.com/x"},
		{"trim spaces", "  https://baidu.com  ", "https://baidu.com"},
		{"already canonical", "https://baidu.com/path?a=1", "https://baidu.com/path?a=1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := NormalizeURL(c.in); got != c.want {
				t.Fatalf("NormalizeURL(%q)=%q, want %q", c.in, got, c.want)
			}
		})
	}
	// 解析失败应原样返回（仅去空格）
	if got := NormalizeURL("::::not a url::::"); got != "::::not a url::::" {
		t.Fatalf("NormalizeURL(invalid)=%q, want trimmed original", got)
	}
}

func TestExtractDomain(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"https://baidu.com/path", "baidu.com", false},
		{"http://example.com:8080/x", "example.com", false},
		{"https://sub.domain.example.com", "sub.domain.example.com", false},
		{"not a url", "", true},
		{"", "", true},
	}
	for _, c := range cases {
		got, err := ExtractDomain(c.in)
		if c.wantErr {
			if err == nil {
				t.Fatalf("ExtractDomain(%q) expected error, got %q", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ExtractDomain(%q) unexpected error: %v", c.in, err)
		}
		if got != c.want {
			t.Fatalf("ExtractDomain(%q)=%q, want %q", c.in, got, c.want)
		}
	}
}

func TestSha256Hex(t *testing.T) {
	if got := Sha256Hex(""); got != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Fatalf("Sha256Hex(\"\") = %q", got)
	}
	if got := Sha256Hex("abc"); got != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Fatalf("Sha256Hex(\"abc\") = %q", got)
	}
	// 相同输入应稳定
	if Sha256Hex("x") != Sha256Hex("x") {
		t.Fatal("Sha256Hex not stable")
	}
}

func TestRandString(t *testing.T) {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, n := range []int{0, 1, 8, 32} {
		s := RandString(n)
		if len(s) != n {
			t.Fatalf("RandString(%d) len=%d", n, len(s))
		}
		for _, r := range s {
			if !strings.ContainsRune(charset, r) {
				t.Fatalf("RandString contains invalid char %q", r)
			}
		}
	}
	// 两次应不同（极高概率）
	if RandString(32) == RandString(32) {
		t.Skip("RandString collision (extremely unlikely)")
	}
}

func TestBase62Encode(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0"},
		{35, "z"}, // 小写 z 位于索引 35
		{61, "Z"}, // 大写 Z 位于索引 61
		{62, "10"},
		{3844, "100"}, // 62*62
	}
	for _, c := range cases {
		if got := Base62Encode(c.in); got != c.want {
			t.Fatalf("Base62Encode(%d)=%q, want %q", c.in, got, c.want)
		}
	}
}

func TestSnowflake(t *testing.T) {
	sf := NewSnowflake(1)
	id1 := sf.NextID()
	id2 := sf.NextID()
	if id1 == 0 || id2 == 0 {
		t.Fatal("Snowflake produced zero id")
	}
	if id2 <= id1 {
		t.Fatalf("Snowflake ids not increasing: %d then %d", id1, id2)
	}
	// workerID 应被归一化到 [0,1023]
	if NewSnowflake(2048).workerID != 0 {
		t.Fatal("Snowflake workerID not mod 1024")
	}
}
