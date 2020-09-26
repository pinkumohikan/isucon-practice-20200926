package main

import (
	"database/sql"
	"fmt"
	"isucon8/isucoin/controller"
	"isucon8/isucoin/model"
	"log"
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
	store := sessions.NewCookieStore([]byte(SessionSecret))

	const LogSendInterval = 1000 / 20
	go func (logs <- chan model.LogPayload) {
		for l := range logs {
			s := time.Now().UnixNano() / 1000

			logger, err := model.Logger(db)
			if err != nil {
				log.Printf("[WARN] new logger failed. tag: %s, v: %v, err:%s", l.Tag, l.Value, err)
				return
			}
			err = logger.Send(l.Tag, l.Value)
			if err != nil {
				log.Printf("[WARN] logger send failed. tag: %s, v: %v, err:%s", l.Tag, l.Value, err)
			}

			e := time.Now().UnixNano() / 1000
			elapsed := e - s
			interval := LogSendInterval - elapsed
			if interval > 0 {
				time.Sleep(LogSendInterval * time.Millisecond)
			}
		}
	}(model.SendLogChan)
	tradeChanceChan := make(chan bool, 9999)
	const TradeInterval = 10 * time.Millisecond

	go func (chances <-chan bool) {
		for _ = range chances {
			if err := model.RunTrade(db); err != nil {
				log.Printf("err: %s", err)
			}
			time.Sleep(TradeInterval)
		}
	}(tradeChanceChan)

	h := controller.NewHandler(db, store, tradeChanceChan)

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
