package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shikidalab/anonymize-ecg/controller"
	"github.com/shikidalab/anonymize-ecg/model"
)

func main() {
	// .envファイルを読み込む
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	// 環境変数を取得
	dsn := os.Getenv("DSN")

	// dbのセットアップ
	err = model.SetupDB(dsn)
	if err != nil {
		log.Fatal(err)
	}

	// `-export` オプションを定義
	export := flag.Bool("export", false, "Export the data")

	// 引数を解析
	flag.Parse()

	// `-export` が指定された場合はcsvに吐き出して終了
	if *export {
		err := controller.SaveCSVFile()
		if err != nil {
			log.Fatalf("Error saving CSV file: %v", err)
		}
		return
	}

	// httpサーバのセットアップ
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // フロントエンドのオリジン
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	// サーバのルーティング設定
	router.GET("/", controller.GetTop)
	router.POST("/", controller.AnonymizeECG)
	router.GET("/download-csv", controller.ExportCSV)

	// サーバの起動
	router.Run()
}
