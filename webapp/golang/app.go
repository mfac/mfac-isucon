package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var (
	db *sqlx.DB
)

const (
	PerPage = 20
)

type Memo struct {
	ID        int          `db:"id"         json:"id"`
	Lat       float64      `db:"lat"        json:"lat"`
	Lng       float64      `db:"lng"        json:"lng"`
	Body      string       `db:"body"       json:"body"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	Emojis    []EmojiCount `json:"emojis"`
}

type EmojiCount struct {
	Emoji string `db:"emoji" json:"emoji"`
	Count int    `db:"count" json:"count"`
}

type MemoEmoji struct {
	ID        int       `db:"id"`
	MemoID    int       `db:"memo_id" json:"memo_id"`
	Emoji     string    `db:"emoji" json:"emoji"`
	CreatedAt time.Time `db:"created_at"`
}

type MemoList struct {
	Memos   []Memo `json:"memos"`
	Page    int    `json:"page"`
	HasPrev bool   `json:"has_prev"`
	HasNext bool   `json:"has_next"`
}

type MemoMessage struct {
	Lat  float64
	Lng  float64
	Body string
}

func dbInitialize() {
	sqls := []string{
		"DELETE FROM memo WHERE id > 100001",
	}

	for _, sql := range sqls {
		db.Exec(sql)
	}
}

func writeJSONResponse(w http.ResponseWriter, result interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	encoder.Encode(result)
}

func decidePage(r *http.Request) (int, error) {
	m, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return 0, err
	}

	page := 0
	mpage := m.Get("page")
	if mpage == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(mpage)
		if err != nil {
			return 0, err
		}
	}
	return page, nil
}

func getInitialize(w http.ResponseWriter, r *http.Request) {
	dbInitialize()
	result := map[string]string{
		"result": "ok",
	}
	writeJSONResponse(w, result)
}

func getMemos(w http.ResponseWriter, r *http.Request) {
	page, err := decidePage(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	limit := PerPage + 1
	offset := PerPage * (page - 1)

	memos := []Memo{}
	dberr := db.Select(&memos,
		`SELECT id, body, X(latlng) lat, Y(latlng) lng, created_at FROM memo
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?`,
		limit, offset)
	if dberr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(dberr)
		return
	}

	hasPrev := page > 1
	hasNext := len(memos) == limit

	// 余分にとった分を削除
	if hasNext {
		memos = memos[:len(memos)-1]
	}

	memos, ierr := inflateMemos(memos)
	if ierr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(dberr)
		return
	}

	result := MemoList{
		Memos:   memos,
		HasPrev: hasPrev,
		HasNext: hasNext,
		Page:    page,
	}
	writeJSONResponse(w, result)
}

func getMemosByTag(c web.C, w http.ResponseWriter, r *http.Request) {
	page, err := decidePage(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	tag := c.URLParams["tag"]

	limit := PerPage + 1
	offset := PerPage * (page - 1)
	regexp := fmt.Sprintf("#%s[[:>:]]", tag)

	memos := []Memo{}
	dberr := db.Select(&memos,
		`SELECT id, body, X(latlng) lat, Y(latlng) lng, created_at FROM memo
        WHERE
          body REGEXP ?
        ORDER BY created_at DESC
        LIMIT ?
        OFFSET ?`,
		regexp, limit, offset)
	if dberr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(dberr)
		return
	}

	hasPrev := page > 1
	hasNext := len(memos) == limit

	// 余分にとった分を削除
	if hasNext {
		memos = memos[:len(memos)-1]
	}

	memos, ierr := inflateMemos(memos)
	if ierr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(dberr)
		return
	}

	result := MemoList{
		Memos:   memos,
		HasPrev: hasPrev,
		HasNext: hasNext,
		Page:    page,
	}
	writeJSONResponse(w, result)
}

func getAroundMemos(c web.C, w http.ResponseWriter, r *http.Request) {
	memoID := c.URLParams["memo_id"]
	selectedMemos := []Memo{}
	memoerr := db.Select(&selectedMemos,
		`SELECT id, X(latlng) lat, Y(latlng) lng FROM memo
        WHERE id = ?`,
		memoID)
	if memoerr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(memoerr)
		return
	}
	selectedMemo := selectedMemos[0]

	limit := PerPage

	memos := []Memo{}
	sql := fmt.Sprintf(
		`SELECT id, body, X(latlng) lat, Y(latlng) lng, created_at FROM memo
        ORDER BY 
          GLength(
            GeomFromText(
                CONCAT('LineString(%f %f,', X(latlng),' ',Y(latlng), ')')
            )
          )
        LIMIT ?`,
		selectedMemo.Lat, selectedMemo.Lng)

	dberr := db.Select(&memos, sql, limit)
	if dberr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(dberr)
		return
	}

	memos, ierr := inflateMemos(memos)
	if ierr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(dberr)
		return
	}

	result := MemoList{
		Memos:   memos,
		HasPrev: false,
		HasNext: false,
		Page:    1,
	}
	writeJSONResponse(w, result)
}

func postMemo(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("aaa", err)
		return
	}

	var msg MemoMessage
	err = json.Unmarshal(bytes, &msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("bbb", err)
		return
	}

	sql := fmt.Sprintf(
		`INSERT INTO memo
        (latlng, body, created_at, updated_at) VALUES 
        (GeomFromText('POINT(%f %f)'), ?, NOW(), NOW())`,
		msg.Lat, msg.Lng)

	result, err := db.Exec(sql, msg.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("ccc", err.Error())
		return
	}

	memoID, lerr := result.LastInsertId()
	if lerr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(lerr.Error())
		return
	}

	memos := []Memo{}
	dberr := db.Select(&memos,
		`SELECT id, body, X(latlng) lat, Y(latlng) lng, created_at FROM memo
        WHERE id = ?`,
		memoID)
	if dberr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(dberr)
		return
	}

	writeJSONResponse(w, memos[0])
}

func postMemoEmoji(c web.C, w http.ResponseWriter, r *http.Request) {
	emoji := c.URLParams["emoji"]
	memoID, err := strconv.Atoi(c.URLParams["memo_id"])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("ccc", err.Error())
		return
	}

	_, err = db.Exec(
		`INSERT INTO memo_emoji 
        (memo_id, emoji, created_at) VALUES
        (?, ?, NOW())`,
		memoID, emoji)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("ccc", err.Error())
		return
	}

	res := MemoEmoji{}
	res.MemoID = memoID
	res.Emoji = emoji
	writeJSONResponse(w, res)
}

func inflateMemos(memos []Memo) ([]Memo, error) {
	for i := 0; i < len(memos); i++ {
		var emojis []EmojiCount
		err := db.Select(&emojis,
			`SELECT emoji, COUNT(emoji) count
            FROM memo_emoji
            WHERE memo_id = ?
            GROUP BY emoji`,
			memos[i].ID)
		if err != nil {
			return nil, err
		}
		memos[i].Emojis = emojis
	}
	return memos, nil
}

func main() {
	host := os.Getenv("GEOMEMO_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("GEOMEMO_DB_PORT")
	if port == "" {
		port = "3306"
	}
	_, err := strconv.Atoi(port)
	if err != nil {
		log.Fatalf("Failed to read DB port number from an environment variable ISUCONP_DB_PORT.\nError: %s", err.Error())
	}
	user := os.Getenv("GEOMEMO_DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("GEOMEMO_DB_PASSWORD")
	dbname := os.Getenv("GEOMEMO_DB_NAME")
	if dbname == "" {
		dbname = "geomemo"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user,
		password,
		host,
		port,
		dbname,
	)

	db, err = sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s.", err.Error())
	}
	defer db.Close()

	goji.Get("/initialize", getInitialize)
	goji.Get("/memo", getMemos)
	goji.Get("/tag/:tag", getMemosByTag)
	goji.Get("/around/:memo_id", getAroundMemos)
	goji.Post("/memo", postMemo)
	goji.Post("/memo/:memo_id/emoji/:emoji", postMemoEmoji)

	goji.Get("/", http.FileServer(http.Dir("public")))
	goji.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("public/static"))))
	goji.Serve()
}
