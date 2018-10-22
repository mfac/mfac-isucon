package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"

	"github.com/kfly8/mfac-isucon/bench/checker"
	"github.com/kfly8/mfac-isucon/bench/score"
	"github.com/kfly8/mfac-isucon/bench/util"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota

	FailThreshold     = 5
	InitializeTimeout = time.Duration(10) * time.Second
	BenchmarkTimeout  = 60 * time.Second
	WaitAfterTimeout  = 10

	MemosPerPage = 20
)

// CLI is the command line object
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
}

type Memo struct {
	ID        int
	Lat       float64
	Lng       float64
	Body      string
	UpdatedAt string
	CreatedAt string
}

type Emoji string

type MemoEmoji struct {
	ID        int
	MemoID    int
	Emoji     Emoji
	CreatedAt string
}

type Tag string

type Output struct {
	Pass     bool     `json:"pass"`
	Score    int64    `json:"score"`
	Suceess  int64    `json:"success"`
	Fail     int64    `json:"fail"`
	Messages []string `json:"messages"`
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		target   string
		userdata string

		version bool
		debug   bool
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	flags.StringVar(&target, "target", "", "")
	flags.StringVar(&target, "t", "", "(Short)")

	flags.StringVar(&userdata, "userdata", "", "userdata directory")
	flags.StringVar(&userdata, "u", "", "userdata directory")

	flags.BoolVar(&version, "version", false, "Print version information and quit.")

	flags.BoolVar(&debug, "debug", false, "Debug mode")
	flags.BoolVar(&debug, "d", false, "Debug mode")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		return ExitCodeError
	}

	// Show version
	if version {
		fmt.Fprintf(cli.errStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	targetHost, terr := checker.SetTargetHost(target)
	if terr != nil {
		outputNeedToContactUs(terr.Error())
		return ExitCodeError
	}

	initialize := make(chan bool)

	setupInitialize(targetHost, initialize)

	memos, tags, emojis, err := prepareUserdata(userdata)
	if err != nil {
		outputNeedToContactUs(err.Error())
		return ExitCodeError
	}

	initReq := <-initialize

	if !initReq {
		fmt.Println(outputResultJson(false, []string{"初期化リクエストに失敗しました"}))

		return ExitCodeError
	}

	// 最初にDOMチェックなどをやってしまい、通らなければさっさと失敗させる
	memoScenario(checker.NewSession(), 1)
	tagScenario(checker.NewSession(), randomTag(tags), 1)
	aroundScenario(checker.NewSession(), randomMemo(memos))
	postMemoScenario(checker.NewSession())
	postMemoEmojiScenario(checker.NewSession(), randomMemo(memos).ID, randomEmoji(emojis))

	if score.GetInstance().GetFails() > 0 {
		fmt.Println(outputResultJson(false, score.GetFailErrorsStringSlice()))
		return ExitCodeError
	}

	memoMoreAndMoreScenarioCh := makeChanBool(1)
	tagMoreAndMoreScenarioCh := makeChanBool(1)
	aroundScenarioCh := makeChanBool(1)
	postMemoScenarioCh := makeChanBool(1)
	postMemoEmojiScenarioCh := makeChanBool(1)

	timeoutCh := time.After(BenchmarkTimeout)

L:
	for {
		select {
		case <-memoMoreAndMoreScenarioCh:
			go func() {
				memoMoreAndMoreScenario(checker.NewSession(), randomMemoPage(memos))
				memoMoreAndMoreScenarioCh <- true
			}()
		case <-tagMoreAndMoreScenarioCh:
			go func() {
				tagMoreAndMoreScenario(checker.NewSession(), randomTag(tags), randomMemoPage(memos))
				tagMoreAndMoreScenarioCh <- true
			}()
		case <-aroundScenarioCh:
			go func() {
				aroundScenario(checker.NewSession(), randomMemo(memos))
				aroundScenarioCh <- true
			}()
		case <-postMemoScenarioCh:
			go func() {
				postMemoScenario(checker.NewSession())
				postMemoScenarioCh <- true
			}()
		case <-postMemoEmojiScenarioCh:
			go func() {
				postMemoEmojiScenario(checker.NewSession(), randomMemo(memos).ID, randomEmoji(emojis))
				postMemoEmojiScenarioCh <- true
			}()

		case <-timeoutCh:
			break L
		}
	}

	time.Sleep(WaitAfterTimeout)

	msgs := []string{}

	if !debug {
		msgs = score.GetFailErrorsStringSlice()
	} else {
		msgs = score.GetFailRawErrorsStringSlice()
	}

	fmt.Println(outputResultJson(true, msgs))

	return ExitCodeOK
}

func outputResultJson(pass bool, messages []string) string {
	output := Output{
		Pass:     pass,
		Score:    score.GetInstance().GetScore(),
		Suceess:  score.GetInstance().GetSucesses(),
		Fail:     score.GetInstance().GetFails(),
		Messages: messages,
	}

	b, _ := json.Marshal(output)

	return string(b)
}

// 主催者に連絡して欲しいエラー
func outputNeedToContactUs(message string) {
	fmt.Println(outputResultJson(false, []string{"！！！主催者に連絡してください！！！", message}))
}

func makeChanBool(len int) chan bool {
	ch := make(chan bool, len)
	for i := 0; i < len; i++ {
		ch <- true
	}
	return ch
}

func randomMemo(memos []Memo) Memo {
	return memos[util.RandomNumber(len(memos))]
}

func randomMemoPage(memos []Memo) int {
	d := float64(len(memos)) / MemosPerPage
	return util.RandomNumber(int(math.Ceil(d)))
}

func randomTag(tags []Tag) Tag {
	return tags[util.RandomNumber(len(tags))]
}

func randomEmoji(emojis []Emoji) Emoji {
	return emojis[util.RandomNumber(len(emojis))]
}

func setupInitialize(targetHost string, initialize chan bool) {
	go func(targetHost string) {
		client := &http.Client{
			Timeout: InitializeTimeout,
		}

		parsedURL, _ := url.Parse("/initialize")
		parsedURL.Scheme = "http"
		parsedURL.Host = targetHost

		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		req.Header.Set("User-Agent", checker.UserAgent)

		res, err := client.Do(req)

		if err != nil {
			fmt.Println(err)
			initialize <- false
			return
		}
		defer res.Body.Close()
		initialize <- true
	}(targetHost)
}
