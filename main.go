package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableKeepAlives:  false,
		},
	}

	blacklist = make(map[string]struct{})
)

func main() {
	loadBlacklist("blacklist.txt")
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Work has started!!! Source code is available at: https://github.com/0-RTT/ghproxy-go")
	})
	router.NoRoute(handler)
	err := router.Run(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func loadBlacklist(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error loading blacklist: %v\n", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		user := strings.TrimSpace(scanner.Text())
		if user != "" {
			blacklist[user] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading blacklist: %v\n", err)
	}
}

func handler(c *gin.Context) {
	rawPath := strings.TrimPrefix(c.Request.URL.RequestURI(), "/")
	re := regexp.MustCompile(`^(http:|https:)?/?/?(.*)`)
	matches := re.FindStringSubmatch(rawPath)
	rawPath = "https://" + matches[2]
	matched := false
	var user string
	for _, exp := range exps {
		if match := exp.FindStringSubmatch(rawPath); match != nil {
			matched = true
			user = match[1]
			break
		}
	}
	if !matched {
		c.String(http.StatusForbidden, "Invalid input.")
		return
	}
	if _, blocked := blacklist[user]; blocked {
		c.String(http.StatusForbidden, "Access denied.")
		return
	}
	if exps[1].MatchString(rawPath) {
		rawPath = strings.Replace(rawPath, "/blob/", "/raw/", 1)
	}
	proxy(c, rawPath)
}

func proxy(c *gin.Context, u string) {
	req, err := http.NewRequest(c.Request.Method, u, c.Request.Body)
	if err != nil {
		http.Error(c.Writer, fmt.Sprintf("server error %v", err), http.StatusInternalServerError)
		return
	}
	for key, values := range c.Request.Header {
		if key == "Content-Type" || key == "User-Agent" {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	req.Header.Del("Host")
	resp, err := httpClient.Do(req)
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
	c.Header("Cache-Control", "max-age=604800")
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
