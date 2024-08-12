package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shikidalab/anonymize-ecg/controller"
	"github.com/shikidalab/anonymize-ecg/model"
)

func main() {
	// .envファイルを読み込む
	err := godotenv.Load()
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

	// `-export` が指定された場合の挙動
	if *export {
		db, err := model.GetDB(dsn)
		if err != nil {
			log.Fatal(err)
		}

		// DBのファイルネームから拡張子を除外する(ex. database.sqlite -> database)
		outputFilename := strings.TrimSuffix(dsn, filepath.Ext(dsn))
		// 拡張子にcsvを採用する
		err = model.ExportPatientsToCSV(db, fmt.Sprintf("%s.csv", outputFilename))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Exporting data...")
		// ここにエクスポート処理を記述
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

	// サーバの起動
	router.Run()
}
