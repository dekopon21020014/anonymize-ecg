package controller

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dekopon21020014/anonymize-mfer/mfer"
	"github.com/dekopon21020014/anonymize-mfer/xml"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func GetRouter() *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // フロントエンドのオリジン
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
	}))

	router.GET("/", getTop)
	router.POST("/", anonymizeMfer)

	return router
}

func getTop(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to LP"})
}

func anonymizeMfer(c *gin.Context) {
	// フロントのフォームから送信されたデータをすべて取得
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	password := c.PostForm("password")
	passwordConfirmation := c.PostForm("passwordConfirmation")

	// パスワードが送られてきているか
	if password == "" || passwordConfirmation == "" {
		c.String(http.StatusBadRequest, "Both password and password confirmation are required")
		return
	}

	// パスワードとパスワードの確認が一致しているか
	if password != passwordConfirmation {
		c.String(http.StatusBadRequest, "Passwords do not match")
		return
	}

	// formのなかのファイルを取得
	files := form.File
	if len(files) == 0 {
		c.String(http.StatusBadRequest, "No files found")
		return
	}

	// 匿名化されたファイルを保存するための構造体のスライス
	// なのでスライスの各要素は構造体です
	anonymizedFiles := []struct {
		Name    string
		Content []byte
	}{}
	anonymizedPrefix := "anonymized-" // 匿名化したときのファイル名のプレフィックス

	for _, fileHeaders := range files {
		for _, fileHeader := range fileHeaders {
			filename := fileHeader.Filename
			var isMfer, isXML bool
			// mwfじゃない，または，MWFじゃない時は処理しない
			if strings.HasSuffix(filename, ".mwf") || strings.HasSuffix(filename, ".MWF") {
				isMfer = true
			} else if strings.HasSuffix(filename, ".xml") || strings.HasSuffix(filename, ".XML") {
				isXML = true
			} else { // mferでもxmlでもないときは処理しない
				continue
			}

			// ファイルオープン
			file, err := fileHeader.Open()
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("open file err: %s", err.Error()))
				return
			}
			defer file.Close()
			fmt.Println(fileHeader.Filename)

			// ファイルオブジェクトをバイト列として読み出し
			data, err := io.ReadAll(file)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			var anonymizedData []byte
			// 匿名化処理
			if isMfer {
				ad, err := mfer.Anonymize(data)
				if err != nil {
					fmt.Println(err)
					c.String(http.StatusInternalServerError, fmt.Sprintf("process file err: %s", err.Error()))
					return
				}
				anonymizedData = ad
			} else if isXML {
				ad, err := xml.Anonymize(data)
				if err != nil {
					fmt.Println(err)
					c.String(http.StatusInternalServerError, fmt.Sprintf("process file err: %s", err.Error()))
					return
				}
				anonymizedData = ad
			}

			// 一旦変数(構造体)にする．そのあとすぐにスライスにappendする
			tmpAnonymizedFile := struct {
				Name    string
				Content []byte
			}{
				Name:    anonymizedPrefix + fileHeader.Filename,
				Content: anonymizedData,
			}

			// 匿名化されたデータをスライスに追加
			anonymizedFiles = append(anonymizedFiles, tmpAnonymizedFile)
		}
	}

	// ZIPファイルを作成するためのバッファ
	bufForResponse := new(bytes.Buffer)
	zipWriter := zip.NewWriter(bufForResponse)

	// 各ファイルをZIPに追加
	for _, file := range anonymizedFiles {
		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ZIPファイルの作成に失敗しました"})
			return
		}
		_, err = zipFile.Write(file.Content)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ファイルの書き込みに失敗しました"})
			return
		}
	}

	// ZIPファイルの書き込みを閉じる
	zipWriter.Close()

	// クライアントにZIPファイルを返却
	c.Header("Content-Disposition", "attachment; filename=anonymized-files.zip")
	c.Header("Content-Type", "application/zip")
	c.Data(http.StatusOK, "application/zip", bufForResponse.Bytes())
}
