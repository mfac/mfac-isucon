package main

import (
	"encoding/json"
	"fmt"

	"github.com/kfly8/mfac-isucon/bench/checker"
)

type APIMemoList struct {
	Memos   []APIMemo `json:"memos"`
	Page    int       `json:"page"`
	HasPrev bool      `json:"has_prev"`
	HasNext bool      `json:"has_next"`
}

type APIMemo struct {
	ID        int     `json:"id"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Body      string  `json:"body"`
	CreatedAt string  `json:"created_at"`
	Emojis    []struct {
		Emoji string `json:"emoji"`
		Count int    `json:"count"`
	} `json:"emojis"`
}

type APIMemoEmoji struct {
	Emoji  Emoji `json:"emoji"`
	MemoID int   `json:"memo_id"`
}

func memoScenario(s *checker.Session, page int) {
	memo := checker.NewAPIAction("GET", fmt.Sprintf("/memo?page=%d", page))
	memo.Description = "/memo APIが表示できること"
	memo.CheckFunc = func(jsonBytes []byte) error {
		data := new(APIMemoList)
		err := json.Unmarshal(jsonBytes, data)
		if err != nil {
			return err
		}

		return nil
	}
	err := memo.Play(s)
	if err != nil {
		fmt.Println(err)
	}
}

func memoMoreAndMoreScenario(s *checker.Session, page int) {
	for i := 1; i < page; i++ {
		memoScenario(s, i)
	}
}

func tagScenario(s *checker.Session, tag Tag, page int) {
	memo := checker.NewAPIAction("GET", fmt.Sprintf("/tag/%s?page=%d", tag, page))
	memo.Description = "/tag APIが表示できること"
	memo.CheckFunc = func(jsonBytes []byte) error {
		data := new(APIMemoList)
		err := json.Unmarshal(jsonBytes, data)
		if err != nil {
			return err
		}

		return nil
	}
	err := memo.Play(s)
	if err != nil {
		fmt.Println(err)
	}
}

func aroundScenario(s *checker.Session, memo Memo) {
	around := checker.NewAPIAction("GET", fmt.Sprintf("/around/%d", memo.ID))

	around.Description = "/around APIが表示できること"
	around.CheckFunc = func(jsonBytes []byte) error {
		data := new(APIMemoList)
		err := json.Unmarshal(jsonBytes, data)
		if err != nil {
			return err
		}

		return nil
	}
	err := around.Play(s)
	if err != nil {
		fmt.Println(err)
	}
}

func tagMoreAndMoreScenario(s *checker.Session, tag Tag, page int) {
	for i := 1; i < page; i++ {
		tagScenario(s, tag, i)
	}
}

func postMemoScenario(s *checker.Session) {
	memo := checker.NewAPIAction("POST", "/memo")

	// TODO ランダマイズ
	lat := "52.32143"
	lng := "23.24"
	body := "yayayaya"

	memo.PostData = map[string]string{
		"lat":  lat,
		"lng":  lng,
		"body": body,
	}

	memo.Description = "/memo が作成できること"
	memo.CheckFunc = func(jsonBytes []byte) error {
		data := new(APIMemo)
		err := json.Unmarshal(jsonBytes, data)
		if err != nil {
			return err
		}

		return nil
	}
	err := memo.Play(s)
	if err != nil {
		fmt.Println(err)
	}
}

func postMemoEmojiScenario(s *checker.Session, memoID int, emoji Emoji) {
	memo := checker.NewAPIAction("POST", fmt.Sprintf("/memo/%d/emoji/%s", memoID, emoji))

	memo.Description = "/memo emoji が作成できること"
	memo.CheckFunc = func(jsonBytes []byte) error {
		data := new(APIMemoEmoji)
		err := json.Unmarshal(jsonBytes, data)
		if err != nil {
			return err
		}

		return nil
	}
	err := memo.Play(s)
	if err != nil {
		fmt.Println(err)
	}
}
