package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "log"
	"net/http"
	"os"
	_ "strconv"
	"time"

	"github.com/go-gorp/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// QuestionList Questionを複数格納するstruck
type QuestionList struct {
	Object string     `json:"object"`
	Type   string     `json:"type"`
	Data   []Question `json:"data"`
}

// Question Questionオブジェクトを扱うためのstruct
type Question struct {
	ID        string    `json:"id" db:"id"`
	Object    string    `json:"object" db:"object"`
	Username  string    `json:"username" db:"username"`
	EventID   string    `json:"event_id" db:"event_id"`
	ProgramID string    `json:"program_id" db:"program_id"`
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Like      int       `json:"like" db:"like_count"`
}

var redisPool *redis.Pool

// QuestionCreateHandler QuestionオブジェクトをDBとRedisに書き込む
func QuestionCreateHandler(w http.ResponseWriter, r *http.Request) {

	// DBとRedisに書き込むためのstiruct Object を生成。POST REQUEST のBodyから値を取得
	var question Question
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&question)

	/* POST REQUEST の BODY に含まれていない値の生成 */
	// uuid
	newUUID := uuid.New()
	question.ID = newUUID.String()

	// object
	question.Object = "question"

	// username
	// TODO: Cookieからsessionidを取得して、Redisに存在する場合は、usernameを取得してquestionオブジェクトに格納する
	question.Username = "anonymous"

	// event_id URLに含まれている event_id を取得して、questionオブジェクトに格納
	vars := mux.Vars(r)
	eventID := vars["event_id"]
	question.EventID = eventID

	// いいねの数
	question.Like = 0

	// 時刻の取得
	now := time.Now()
	question.UpdatedAt = now
	question.CreatedAt = now

	// debug
	//w.Write([]byte("comment: " + question.Comment + "\n" +
	//	"ID: " + question.ID + "\n" +
	//	"Object: " + question.Object + "\n" +
	//	"eventID: " + question.EventID + "\n" +
	//	"programID: " + question.ProgramID + "\n" +
	//	"username: " + question.Username + "\n" +
	//	"Like: " + strconv.Itoa(question.Like) + "\n"))

	dbmap, err := initDb()
	defer dbmap.Db.Close()

	// debug SQL Trace
	//dbmap.TraceOn("", log.New(os.Stdout, "gorptest: ", log.Lmicroseconds))

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	/* データの挿入 */
	err = dbmap.Insert(&question)

	if err != nil {
		fmt.Printf("%+v", err)
		return
	}

	//var buf bytes.Buffer
	//enc := json.NewEncoder(&buf)
	//if err := enc.Encode(question); err != nil {
	//	log.Fatal(err)
	//}

	//fmt.Println(buf.String())

}

// QuestionListHandler QuestionオブジェクトをRedisから取得する。存在しない場合はDBから取得し、Redisへ格納する
// TODO: pagenationなどのパラメータ制御。まだ仮実装
func QuestionListHandler(w http.ResponseWriter, r *http.Request) {
	// URLに含まれている event_id を取得
	vars := mux.Vars(r)
	eventID := vars["event_id"]

	// Redisにデータが存在するか確認
	//redisConnection := getRedisConnection()
	//defer redisConnection.Close()

	// DBからデータを取得
	//dbmap, err := initDb()
	//defer dbmap.Db.Close()

	//if err != nil {
	//	causeErr := errors.Cause(err)
	//	fmt.Printf("%+v", causeErr)
	//	return
	//}

	//var questions []Question
	//_, err = dbmap.Select(&questions, "select * from questions")

	//if err != nil {
	//	causeErr := errors.Cause(err)
	//	fmt.Printf("%+v", causeErr)
	//	return
	//}

	// DB or Redis から取得したデータのtimezoneをAsia/Tokyoと指定
	//locationTokyo, err := time.LoadLocation("Asia/Tokyo")
	//for i := range questions {
	//	questions[i].CreatedAt = questions[i].CreatedAt.In(locationTokyo)
	//	questions[i].UpdatedAt = questions[i].UpdatedAt.In(locationTokyo)
	//}

	// DB or Redisから取得したデータをQuestionListの構造体に格納
	//var questionList QuestionList
	//questionList.Data = questions
	//questionList.Object = "list"
	//questionList.Type = "question"

	questionList := getQuestionList(eventID)

	/* JSONの整形 */
	// QuestionのStructをjsonとして変換
	jsonBytes, _ := json.Marshal(questionList)

	// 整形用のバッファを作成し、整形を実行
	out := new(bytes.Buffer)
	// プリフィックスなし、スペース2つでインデント
	json.Indent(out, jsonBytes, "", "  ")

	w.Write([]byte(out.String()))

	// redis test1
	//w2, _ := redisConnection.Do("SET", "testjon", questions[0])
	//r2, _ := redisConnection.Do("GET", "testjon")
	//fmt.Print(w2)
	//fmt.Print(r2)

	// redis test2
	// sorted list
	//redisConnection.Do("ZADD", "testsort1", 1, "a")
	//redisConnection.Do("ZADD", "testsort1", 2, "b")
	//redisConnection.Do("ZADD", "testsort1", 3, "c")
	//redisConnection.Do("ZADD", "testsort1", 4, "d")
	//redisConnection.Do("ZADD", "testsort1", 5, "e")
	//r2, _ := redis.Strings(redisConnection.Do("ZRANGE", "testsort1", 0, -1, "WITHSCORES"))
	//fmt.Print(r2)

	// redis test3
	// sorted list
	//redisConnection.Do("ZADD", "testsort2", questions[0].CreatedAt.Unix(), "quest0")
	//redisConnection.Do("ZADD", "testsort2", questions[1].CreatedAt.Unix(), "quest1")
	//r2, _ := redis.Strings(redisConnection.Do("ZRANGE", "testsort2", 0, -1, "WITHSCORES"))
	//fmt.Print(r2)

	// redis test4
	// JSON in hash
	//serialized0, err := json.Marshal(questions[0])
	//serialized1, err := json.Marshal(questions[1])
	//redisConnection.Do("HSET", "testhash1", "1", serialized0)
	//redisConnection.Do("HSET", "testhash1", "2", serialized1)
	//data, _ := redis.Bytes(redisConnection.Do("HGET", "testhash1", "1"))
	//deserialized := new(Question)
	//json.Unmarshal(data, deserialized)
	//fmt.Print(deserialized.Comment)
}

// getQuestions RedisとDBからデータを取得する
func getQuestionList(eventID string) (questionList QuestionList) {
	redisConn := getRedisConnection()
	defer redisConn.Close()

	/* Redisにデータが存在するか確認する。 */
	questionsKey, likeSortedKey, createdSortedKey := getQuestionsKey(eventID)

	// 3種類のKeyが存在しない場合はデータが何かしら不足しているため、データの同期を行う
	hasQuestionsKey := redisHasKey(redisConn, questionsKey)
	hasLikeSortedKey := redisHasKey(redisConn, likeSortedKey)
	hasCreatedSortedKey := redisHasKey(redisConn, createdSortedKey)

	if !hasQuestionsKey || !hasLikeSortedKey || !hasCreatedSortedKey {
		syncQuestion(eventID)
	}

	/* temp */
	// DBからデータを取得
	dbmap, err := initDb()
	defer dbmap.Db.Close()

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	var questions []Question
	_, err = dbmap.Select(&questions, "select * from questions")

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	// DB or Redis から取得したデータのtimezoneをAsia/Tokyoと指定
	locationTokyo, err := time.LoadLocation("Asia/Tokyo")
	for i := range questions {
		questions[i].CreatedAt = questions[i].CreatedAt.In(locationTokyo)
		questions[i].UpdatedAt = questions[i].UpdatedAt.In(locationTokyo)
	}

	questionList.Data = questions
	questionList.Object = "list"
	questionList.Type = "question"
	return questionList
}

// syncQuestion DBとRedisのデータを同期する
// RedisのデータがTTLなどで存在していない場合にsyncQuestionを使用する
func syncQuestion(eventID string) {
	redisConnection := getRedisConnection()
	defer redisConnection.Close()

	// DBからデータを取得
	dbmap, err := initDb()
	defer dbmap.Db.Close()

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	var questions []Question
	_, err = dbmap.Select(&questions, "SELECT * FROM questions WHERE event_id = '"+eventID+"'")

	if err != nil {
		causeErr := errors.Cause(err)
		fmt.Printf("%+v", causeErr)
		return
	}

	// DB or Redis から取得したデータのtimezoneをUTCからAsia/Tokyoと指定
	locationTokyo, err := time.LoadLocation("Asia/Tokyo")
	for i := range questions {
		questions[i].CreatedAt = questions[i].CreatedAt.In(locationTokyo)
		questions[i].UpdatedAt = questions[i].UpdatedAt.In(locationTokyo)
	}

	//Redisで利用するKeyを取得
	questionsKey, likeSortedKey, createdSortedKey := getQuestionsKey(eventID)

	//DBのデータをRedisに同期する。
	for _, question := range questions {
		//HashMap SerializedされたJSONデータを格納
		serializedJSON, _ := json.Marshal(question)
		redisConnection.Do("HSET", questionsKey, question.ID, serializedJSON)

		//SortedSet(Like)
		redisConnection.Do("ZADD", likeSortedKey, question.Like, question.ID)

		//SortedSet(CreatedAt)
		redisConnection.Do("ZADD", createdSortedKey, question.CreatedAt.Unix(), question.ID)
	}
}

// getQuestionsKey Redisで使用するQuestionsを格納するkeyを取得
func getQuestionsKey(eventID string) (questionsKey string, likeSortedKey string, createdSortedKey string) {
	questionsKey = "questions_" + eventID
	likeSortedKey = questionsKey + "_like"
	createdSortedKey = questionsKey + "_created"

	return questionsKey, likeSortedKey, createdSortedKey
}

// redisHasKey
func redisHasKey(conn redis.Conn, key string) bool {
	hasInt, _ := conn.Do("HEXISTS", key)

	var hasKey bool
	if hasInt == 1 {
		hasKey = true
	} else {
		hasKey = false
	}

	return hasKey
}

// initDb 環境変数を利用しDBへのConnectionを取得する(sqldriverでconnection poolが実装されているらしい)
func initDb() (dbmap *gorp.DbMap, err error) {
	dbms := "mysql"
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	protocol := "tcp(" + os.Getenv("DB_URL") + ")"
	dbname := "qicoo"
	option := "?parseTime=true"

	connect := user + ":" + password + "@" + protocol + "/" + dbname + option
	db, err := sql.Open(dbms, connect)

	if err != nil {
		return nil, errors.Wrap(err, "error on initDb()")
	}

	// structの構造体とDBのTableを紐づける
	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}
	dbmap.AddTableWithName(Question{}, "questions")

	return dbmap, nil
}

// initRedisPool RedisConnectionPoolからconnectionを取り出す
func initRedisPool() {
	url := os.Getenv("REDIS_URL")

	// idle connection limit:3    active connection limit:1000
	pool := &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1000,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", url) },
	}

	redisPool = pool
}

// getRedisConnection
func getRedisConnection() (conn redis.Conn) {
	return redisPool.Get()
}

func main() {
	r := mux.NewRouter()

	// 初期設定
	initRedisPool()

	// route QuestionCreate
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("POST").
		HandlerFunc(QuestionCreateHandler)

	// route QuestionList
	r.Path("/v1/{event_id:[a-zA-Z0-9-_]+}/questions").
		Methods("GET").
		HandlerFunc(QuestionListHandler)

	http.ListenAndServe(":8080", r)
}
