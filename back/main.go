package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

	// `-export` が指定された場合はcsvに吐き出して終了
	if *export {
		err := saveCSVFile()
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

func saveCSVFile() error {
	// データベースへの接続を取得
	db, err := model.GetDB(os.Getenv("DSN"))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close() // データベース接続を確実に閉じる

	// CSVデータをエクスポート
	csv, err := model.ExportPatientsToCSV(db)
	if err != nil {
		return fmt.Errorf("failed to export patients to CSV: %v", err)
	}

	// 環境変数から保存先ディレクトリを取得
	saveDir := os.Getenv("SAVE_DIR")
	if saveDir == "" {
		return fmt.Errorf("SAVE_DIR is not set")
	}

	// 保存先ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", saveDir, err)
	}

	// 保存するファイルのフルパスを作成
	filePath := filepath.Join(saveDir, csv.Name)

	// ファイルを作成
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer file.Close() // 閉じる処理を忘れずに追加

	// ファイルにデータを書き込む
	_, err = file.Write(csv.Content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", filePath, err)
	}

	fmt.Printf("CSV file exported successfully to: %s\n", filePath)
	return nil
}
