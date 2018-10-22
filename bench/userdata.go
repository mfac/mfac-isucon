package main

import (
	"bufio"
	"errors"
	"os"
)

func prepareUserdata(userdata string) ([]Memo, []Tag, []Emoji, error) {
	if userdata == "" {
		return nil, nil, nil, errors.New("userdataディレクトリが指定されていません")
	}
	info, err := os.Stat(userdata)
	if err != nil {
		return nil, nil, nil, err
	}
	if !info.IsDir() {
		return nil, nil, nil, errors.New("userdataがディレクトリではありません")
	}

	memos := []Memo{}
	{
		file, err := os.Open(userdata + "/memos.txt")
		if err != nil {
			return nil, nil, nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		i := 1
		for scanner.Scan() {
			text := scanner.Text()
			memos = append(memos, Memo{Body: text, ID: i})
			i++
		}
	}

	var tags []Tag
	{
		file, err := os.Open(userdata + "/tags.txt")
		if err != nil {
			return nil, nil, nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		i := 1
		for scanner.Scan() {
			text := scanner.Text()
			tags = append(tags, Tag(text))
			i++
		}
	}

	var emojis []Emoji
	{
		file, err := os.Open(userdata + "/emojis.txt")
		if err != nil {
			return nil, nil, nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		i := 1
		for scanner.Scan() {
			text := scanner.Text()
			emojis = append(emojis, Emoji(text))
			i++
		}
	}

	return memos, tags, emojis, err
}
