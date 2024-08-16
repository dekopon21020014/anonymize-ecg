package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/shikidalab/anonymize-ecg/model"
)

func ExportCSV(c *gin.Context) {
	db, err := model.GetDB(os.Getenv("DSN"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fileStruct, err := model.ExportPatientsToCSV(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileStruct.Name))
	c.Header("Content-Type", "text/csv")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Data(http.StatusOK, "text/csv", fileStruct.Content)
}

func SaveCSVFile() error {
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
