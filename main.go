package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	sizeLimit = 1024 * 1024 * 1024 * 1 // 1 GB 限制
	host      = "127.0.0.1"
	port      = 8080
)

var (
	// 正则表达式匹配 GitHub 相关 URL
	exps = []*regexp.Regexp{
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:releases|archive)/.*$`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*$`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:info|git-).*$`),
		regexp.MustCompile(`^(?:https?://)?raw\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.+?/.+$`),
		regexp.MustCompile(`^(?:https?://)?gist\.github\.com/([^/]+)/.+?/.+$`),
	}

	// HTTP 客户端连接池
	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableKeepAlives:  false,
		},
	}
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 根路径重定向到 GitHub 项目
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "https://github.com/0-RTT/ghproxy-go")
	})

	// 处理其他未匹配的路由
	router.NoRoute(handler)

	err := router.Run(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func handler(c *gin.Context) {
	rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/")
	re := regexp.MustCompile(`^(http:|https:)?/?/?(.*)`)

	// 提取 URL 的实际部分
	matches := re.FindStringSubmatch(rawPath)
	rawPath = "https://" + matches[2]

	// 检查 URL 是否匹配 GitHub 相关模式
	matched := false
	for _, exp := range exps {
		if exp.MatchString(rawPath) {
			matched = true
			break
		}
	}

	if !matched {
		c.String(http.StatusForbidden, "Invalid input.")
		return
	}

	// 如果匹配的是 /blob/ URL，将其转换为 /raw/
	if exps[1].MatchString(rawPath) {
		rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
	}

	// 执行代理请求
	proxy(c, rawPath)
}

func proxy(c *gin.Context, u string) {
	// 直接创建新的 HTTP 请求
	req, err := http.NewRequest(c.Request.Method, u, c.Request.Body)
	if err != nil {
		http.Error(c.Writer, fmt.Sprintf("server error %v", err), http.StatusInternalServerError)
		return
	}

	// 只复制必要的请求头，减少不必要的开销
	for key, values := range c.Request.Header {
		if key == "Content-Type" || key == "User-Agent" {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	req.Header.Del("Host")

	// 使用连接池中的 httpClient 发出请求
	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(c.Writer, fmt.Sprintf("server error %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 检查响应的 Content-Length 是否超过限制
	if contentLength, ok := resp.Header["Content-Length"]; ok {
		if size, err := strconv.Atoi(contentLength[0]); err == nil && size > sizeLimit {
			finalURL := resp.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, finalURL)
			return
		}
	}

	// 设置缓存控制头
	c.Header("Cache-Control", "max-age=604800")

	// 将响应头写回给客户端
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// 返回响应状态码
	c.Writer.WriteHeader(resp.StatusCode)

	// 将响应体复制到客户端
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		return
	}
}
