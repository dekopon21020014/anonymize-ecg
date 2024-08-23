package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shikidalab/anonymize-ecg/controller"
	"github.com/shikidalab/anonymize-ecg/model"
)

func main() {
	// ログファイルの設定
	if err := os.MkdirAll("log", os.ModePerm); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	loc, _ := time.LoadLocation("Asia/Tokyo")
	logFileName := fmt.Sprintf("log/%s.log", time.Now().In(loc).Format("2006-01-02")) // 実行した日付のログファイルをつくる

	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// 標準出力とファイルに同時にログを書き込むためのMultiWriterを作成
	multiWriter := io.MultiWriter(os.Stdout, f)

	// logをstdoutとログファイルの両方に出す
	log.SetOutput(multiWriter)
	log.Println("main function was started")

	// .envファイルを読み込む
	err = godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// dbの立ち上げ
	dsn := os.Getenv("DSN")
	err = model.SetupDB(dsn)
	if err != nil {
		log.Fatal(err)
	}

	// `-export` オプションを定義
	export := flag.Bool("export", false, "Export the data")

	// `-export` が指定された場合はcsvに吐き出して終了
	flag.Parse()
	if *export {
		err := controller.SaveCSVFile()
		if err != nil {
			log.Fatalf("Error saving CSV file: %v", err)
		}
		return
	}

	// ginのログ出力先をstdoutとlogファイルの両方に指定
	gin.DefaultWriter = multiWriter
	gin.DefaultErrorWriter = multiWriter

	// httpサーバのセットアップ
	router := gin.Default()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("FRONT_ORIGIN")}, // フロントエンドのオリジン
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	// サーバのルーティング設定
	router.GET("/", controller.GetTop)
	router.GET("/upload", controller.AnonymizeECG)
	router.GET("/download-csv", controller.ExportCSV)

	// サーバの起動
	router.Run()
}
