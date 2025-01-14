package main

import (
	"database/sql"
	"fmt"
	"isucon8/isucoin/controller"
	"isucon8/isucoin/model"
	"isucon8/isulogger"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	gctx "github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
)

const (
	SessionSecret = "tonymoris"
)

func init() {
	var err error
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Panicln(err)
	}
	time.Local = loc
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv("ISU_" + key); ok {
		return v
	}
	return def
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var (
		port   = getEnv("APP_PORT", "5000")
		dbhost = getEnv("DB_HOST", "127.0.0.1")
		dbport = getEnv("DB_PORT", "3306")
		dbuser = getEnv("DB_USER", "root")
		dbpass = getEnv("DB_PASSWORD", "")
		dbname = getEnv("DB_NAME", "isucoin")
		public = getEnv("PUBLIC_DIR", "public")
	)

	dbusrpass := dbuser
	if dbpass != "" {
		dbusrpass += ":" + dbpass
	}

	dsn := fmt.Sprintf(`%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&charset=utf8mb4`, dbusrpass, dbhost, dbport, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("mysql connect failed. err: %s", err)
	}
	db.SetMaxIdleConns(30)
	db.SetMaxOpenConns(30)

	store := sessions.NewCookieStore([]byte(SessionSecret))

	go func () {
		t := time.NewTicker(time.Millisecond * 1000)
		for _ = range t.C {
			if len(model.BufferedLogs) > 0 {
				logger, err := model.Logger(db)
				if err != nil {
					log.Printf("Log sending error. err=%s", err)
					return
				}

				model.BufferedLogsMutex.Lock()
				log.Printf("ログを %d件 まとめて送信中...", len(model.BufferedLogs))
				var logs []isulogger.Log
				for _, l := range model.BufferedLogs {
					logs = append(logs, isulogger.Log{
						Tag:  l.Tag,
						Time: time.Now(),
						Data: l.Value,
					})
				}
				model.BufferedLogs = nil
				model.BufferedLogsMutex.Unlock()

				if err := logger.SendBulk(logs); err != nil {
					log.Printf("Log sending error. err=%s", err)
				}
			}
		}
	}()

	go func () {
		t := time.NewTicker(time.Millisecond * 50)
		for _ = range t.C {
			if err := model.RunTrade(db); err != nil {
				log.Printf("err: %s", err)
			}
		}
	}()

	h := controller.NewHandler(db, store)
	model.InitializeCandleStack(&controller.BaseTime)
	go h.InfoUpdate()
	router := httprouter.New()
	router.POST("/initialize", h.Initialize)
	router.POST("/signup", h.Signup)
	router.POST("/signin", h.Signin)
	router.POST("/signout", h.Signout)
	router.GET("/info", h.Info)
	router.POST("/orders", h.AddOrders)
	router.GET("/orders", h.GetOrders)
	router.DELETE("/order/:id", h.DeleteOrders)
	router.NotFound = http.FileServer(http.Dir(public)).ServeHTTP

	addr := ":" + port
	log.Printf("[INFO] start server %s", addr)
	log.Fatal(http.ListenAndServe(addr, gctx.ClearHandler(h.CommonMiddleware(router))))
}
