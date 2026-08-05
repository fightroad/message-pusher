package common

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"html/template"
)

func OpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		log.Println(err)
	}
}

func GetIp() (ip string) {
	ips, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		return ip
	}

	for _, a := range ips {
		if ipNet, ok := a.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				if strings.HasPrefix(ip, "10") {
					return
				}
				if strings.HasPrefix(ip, "172") {
					return
				}
				if strings.HasPrefix(ip, "192.168") {
					return
				}
				ip = ""
			}
		}
	}
	return
}

var sizeKB = 1024
var sizeMB = sizeKB * 1024
var sizeGB = sizeMB * 1024

func Bytes2Size(num int64) string {
	numStr := ""
	unit := "B"
	if num/int64(sizeGB) > 1 {
		numStr = fmt.Sprintf("%.2f", float64(num)/float64(sizeGB))
		unit = "GB"
	} else if num/int64(sizeMB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeMB)))
		unit = "MB"
	} else if num/int64(sizeKB) > 1 {
		numStr = fmt.Sprintf("%d", int(float64(num)/float64(sizeKB)))
		unit = "KB"
	} else {
		numStr = fmt.Sprintf("%d", num)
	}
	return numStr + " " + unit
}

func Seconds2Time(num int) (time string) {
	if num/31104000 > 0 {
		time += strconv.Itoa(num/31104000) + " 年 "
		num %= 31104000
	}
	if num/2592000 > 0 {
		time += strconv.Itoa(num/2592000) + " 个月 "
		num %= 2592000
	}
	if num/86400 > 0 {
		time += strconv.Itoa(num/86400) + " 天 "
		num %= 86400
	}
	if num/3600 > 0 {
		time += strconv.Itoa(num/3600) + " 小时 "
		num %= 3600
	}
	if num/60 > 0 {
		time += strconv.Itoa(num/60) + " 分钟 "
		num %= 60
	}
	time += strconv.Itoa(num) + " 秒"
	return
}

func Interface2String(inter interface{}) string {
	switch v := inter.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	}
	return "Not Implemented"
}

func UnescapeHTML(x string) interface{} {
	return template.HTML(x)
}

func IntMax(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func GetUUID() string {
	code := uuid.New().String()
	code = strings.Replace(code, "-", "", -1)
	return code
}

func Max(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

func Markdown2HTML(markdown string) (HTML string, err error) {
	if markdown == "" {
		return "", nil
	}
	var buf bytes.Buffer
	goldMarkEntity := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
		),
	)
	err = goldMarkEntity.Convert([]byte(markdown), &buf)
	if err != nil {
		return fmt.Sprintf("Markdown 渲染出错：%s", err.Error()), err
	}
	HTML = buf.String()
	return
}

func IsHTMLContent(s string) bool {
	t := strings.TrimSpace(s)
	return t != "" && t[0] == '<'
}

var htmlTagPattern = regexp.MustCompile(`(?s)<[^>]*>`)

// StripHTMLTags removes HTML tags for plain-text channels (e.g. WeChat templates).
func StripHTMLTags(s string) string {
	if s == "" {
		return ""
	}
	return strings.TrimSpace(html.UnescapeString(htmlTagPattern.ReplaceAllString(s, "")))
}

func GetTimestamp() int64 {
	return time.Now().Unix()
}

func Replace(s, old, new string, n int) string {
	new = strings.TrimPrefix(strings.TrimSuffix(fmt.Sprintf("%q", new), "\""), "\"")
	return strings.Replace(s, old, new, n)
}
