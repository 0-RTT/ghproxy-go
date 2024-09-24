package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	sizeLimit = 1024 * 1024 * 1024 * 1
	host      = "127.0.0.1"
	port      = 8080
)

var (
	exps = []*regexp.Regexp{
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:releases|archive)/.*$`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:blob|raw)/.*$`),
		regexp.MustCompile(`^(?:https?://)?github\.com/([^/]+)/([^/]+)/(?:info|git-).*$`),
		regexp.MustCompile(`^(?:https?://)?raw\.github(?:usercontent|)\.com/([^/]+)/([^/]+)/.+?/.+$`),
		regexp.MustCompile(`^(?:https?://)?gist\.github\.com/([^/]+)/.+?/.+$`),
	}
	requestPool = sync.Pool{
		New: func() interface{} {
			return &http.Request{}
		},
	}
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "https://github.com/0-RTT/ghproxy-go")
	})
	router.NoRoute(handler)
	err := router.Run(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func handler(c *gin.Context) {
	rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/")
	re := regexp.MustCompile(`^(http:|https:)?/?/?(.*)`)
	matches := re.FindStringSubmatch(rawPath)
	rawPath = "https://" + matches[2]
	matches = checkURL(rawPath)
	if matches == nil {
		c.String(http.StatusForbidden, "Invalid input.")
		return
	}
	if exps[1].MatchString(rawPath) {
		rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
	}
	proxy(c, rawPath)
}

func proxy(c *gin.Context, u string) {
	req := requestPool.Get().(*http.Request)
	defer requestPool.Put(req)
	var err error
	req, err = http.NewRequest(c.Request.Method, u, c.Request.Body)
	if err != nil {
		http.Error(c.Writer, fmt.Sprintf("server error %v", err), http.StatusInternalServerError)
		return
	}
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Del("Host")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(c.Writer, fmt.Sprintf("server error %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if contentLength, ok := resp.Header["Content-Length"]; ok {
		if size, err := strconv.Atoi(contentLength[0]); err == nil && size > sizeLimit {
			finalURL := resp.Request.URL.String()
			c.Redirect(http.StatusMovedPermanently, finalURL)
			return
		}
	}
	c.Header("Cache-Control", "public, max-age=604800")
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Writer.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		return
	}
}

func checkURL(u string) []string {
	for _, exp := range exps {
		if matches := exp.FindStringSubmatch(u); matches != nil {
			return matches[1:]
		}
	}
	return nil
}
