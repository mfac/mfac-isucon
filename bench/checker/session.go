package checker

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/kfly8/mfac-isucon/bench/score"
)

const (
	UserAgent = "benchmarker"
)

var (
	targetHost string
)

type Session struct {
	Client    *http.Client
	Transport *http.Transport

	logger *log.Logger
}

func NewSession() *Session {
	w := &Session{
		logger: log.New(os.Stdout, "", 0),
	}

	jar, _ := cookiejar.New(&cookiejar.Options{})
	w.Transport = &http.Transport{}
	w.Client = &http.Client{
		Transport: w.Transport,
		Jar:       jar,
		Timeout:   time.Duration(10) * time.Second,
	}

	return w
}

func SetTargetHost(host string) (string, error) {
	parsedURL, err := url.Parse(host)
	if err != nil {
		return "", err
	}

	targetHost = ""

	// 完璧にチェックするのは難しい
	if parsedURL.Scheme == "http" {
		targetHost += parsedURL.Host
	} else if parsedURL.Scheme != "" && parsedURL.Scheme != "https" {
		targetHost += parsedURL.Scheme + ":" + parsedURL.Opaque
	} else {
		targetHost += host
		// return "", fmt.Errorf("不正なホスト名です")
	}

	return targetHost, nil
}

func (s *Session) NewRequest(method, uri string, body io.Reader) (*http.Request, error) {
	parsedURL, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}

	if parsedURL.Host == "" {
		parsedURL.Host = targetHost
	}

	req, err := http.NewRequest(method, parsedURL.String(), body)

	if err != nil {
		return nil, err
	}

	return req, err
}

func escapeQuotes(s string) string {
	return strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(s)
}

func (s *Session) RefreshClient() {
	jar, _ := cookiejar.New(&cookiejar.Options{})
	s.Transport = &http.Transport{}
	s.Client = &http.Client{
		Transport: s.Transport,
		Jar:       jar,
	}
}

func (s *Session) SendRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", UserAgent)

	return s.Client.Do(req)
}

func (s *Session) Success(point int64) {
	score.GetInstance().SetScore(point)
}

func (s *Session) Fail(point int64, req *http.Request, err error) error {
	score.GetInstance().SetFails(point)
	if req != nil {
		err = fmt.Errorf("%s (%s %s)", err, req.Method, req.URL.Path)
	}

	score.GetFailErrorsInstance().Append(err)
	return err
}
