package checker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

const (
	suceessGetScore  = 1
	suceessPostScore = 2

	failErrorScore     = 10
	failExceptionScore = 20
)

type APIAction struct {
	Method             string
	Path               string
	PostData           map[string]string
	Headers            map[string]string
	ExpectedStatusCode int
	ExpectedHeaders    map[string]string
	Description        string
	CheckFunc          func(jsonBytes []byte) error
}

func NewAPIAction(method, path string) *APIAction {
	return &APIAction{
		Method:             method,
		Path:               path,
		ExpectedStatusCode: http.StatusOK,
	}
}

func (a *APIAction) Play(s *Session) error {
	jsonBytes, err := json.Marshal(a.PostData)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	buf := bytes.NewBuffer(jsonBytes)
	req, err := s.NewRequest(a.Method, a.Path, buf)
	defer req.Body.Close()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return s.Fail(failExceptionScore, req, errors.New("リクエストに失敗しました (主催者に連絡してください)"))
	}

	for key, val := range a.Headers {
		req.Header.Add(key, val)
	}

	if req.Method == "POST" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	res, err := s.SendRequest(req)

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return s.Fail(failExceptionScore, req, errors.New("リクエストがタイムアウトしました"))
		}
		fmt.Fprintln(os.Stderr, err)
		return s.Fail(failExceptionScore, req, errors.New("リクエストに失敗しました"))
	}

	defer res.Body.Close()

	if res.StatusCode != a.ExpectedStatusCode {
		return s.Fail(failErrorScore, res.Request, fmt.Errorf("Response code should be %d, got %d", a.ExpectedStatusCode, res.StatusCode))
	}

	if a.CheckFunc != nil {
		jsonBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return s.Fail(
				failErrorScore,
				res.Request,
				err,
			)
		}

		err = a.CheckFunc(jsonBytes)
		if err != nil {
			return s.Fail(
				failErrorScore,
				res.Request,
				err,
			)
		}
	}

	s.Success(suceessGetScore)

	if a.Method == "POST" {
		s.Success(suceessPostScore)
	}

	return nil
}
