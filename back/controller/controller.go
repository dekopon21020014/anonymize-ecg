package controller

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/shikidalab/anonymize-ecg/mfer"
	"github.com/shikidalab/anonymize-ecg/model"
	"github.com/shikidalab/anonymize-ecg/xml"
)

func GetTop(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to LP"})
}

const (
	contentTypeZip        = "application/zip"
	contentDispositionFmt = "attachment; filename=%s"
)

var (
	errPasswordMismatch = errors.New("passwords do not match")
	errFileNameFormat   = errors.New("file name format is incorrect")
	errZipCreation      = errors.New("failed to create ZIP file")
	errFileWrite        = errors.New("failed to write file")
)

type File struct {
	Name    string
	Content []byte
}

func AnonymizeECG(c *gin.Context) {
	ch := make(chan []File)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	passwword, err := validatePassword(conn)
	if err != nil {
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte("ok"))
	if err != nil {
		log.Println("WriteMessage error:", err)
	}

	go receiveMessage(conn, ch)

	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	for files := range ch {
		anonymizedFiles, err := processFiles(files, passwword) // passwrod必須
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		for _, file := range anonymizedFiles {
			zipFile, err := zipWriter.Create(file.Name)
			if err != nil {
				fmt.Printf("%s: %v", errZipCreation, err)
				continue
			}
			_, err = zipFile.Write(file.Content)
			if err != nil {
				fmt.Printf("%s: %v", errFileWrite, err)
				continue
			}
		}
	}
	zipWriter.Close()
	sendZipResponse(c, zipBuffer, conn)
}

func validatePassword(conn *websocket.Conn) (string, error) {
	messageType, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error reading message:", err)
		return "", errors.New(" error reading message")
	}

	var creds struct {
		Type                 string `json:"type"`
		Password             string `json:"password"`
		PasswordConfirmation string `json:"passwordConfirmation"`
	}

	if messageType == websocket.TextMessage {
		err := json.Unmarshal(msg, &creds)
		if err != nil {
			return "", err
		}

		if creds.Password != creds.PasswordConfirmation {
			return "", errPasswordMismatch
		}
	}
	return creds.Password, nil
}

func receiveMessage(conn *websocket.Conn, ch chan []File) {
	for {
		// メッセージを受信する
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		// ZIPファイルをメモリ上で解凍する
		if messageType == websocket.BinaryMessage {
			// ZIPファイルをメモリ上で解凍する
			reader, err := zip.NewReader(bytes.NewReader(msg), int64(len(msg)))
			if err != nil {
				fmt.Println("Error creating ZIP reader:", err)
				continue
			}
			var files []File
			for _, file := range reader.File {
				// ZIPファイル内のファイルを開く
				rc, err := file.Open()
				if err != nil {
					fmt.Println("Error opening ZIP file entry:", err)
					continue
				}

				// ファイルの内容を読み込む
				fileContent, err := io.ReadAll(rc)
				rc.Close()
				if err != nil {
					fmt.Println("Error reading file content:", err)
					continue
				}

				// ファイル情報を構造体にまとめる
				files = append(files, File{
					Name:    file.Name,
					Content: fileContent,
				})
			}
			ch <- files

			// ファイルデータの処理（例: デバッグ出力）
			for _, file := range files {
				fmt.Printf("Received file: %s, size: %d bytes\n", file.Name, len(file.Content))
			}
		}
	}
	close(ch)
}

func processFiles(files []File, password string) ([]File, error) {
	var anonymizedFiles []File

	for _, file := range files {
		anonymizedFile, err := processFile(file, password)
		if err != nil {
			if errors.Is(err, errFileNameFormat) {
				continue
			}
			return nil, err
		}
		if anonymizedFile.Content != nil {
			anonymizedFiles = append(anonymizedFiles, anonymizedFile)
		}
	}

	return anonymizedFiles, nil
}

func processFile(file File, password string) (File, error) {
	fileType, err := getFileType(file.Name)
	if err != nil {
		return File{}, err
	}

	if fileType == "" {
		return File{}, nil // Skip non-MWF and non-XML files
	}

	patientID, date, err := parseFileName(file.Name)
	if err != nil {
		return File{}, err
	}

	anonymizedData, err := anonymizeData(file.Content, fileType)
	if err != nil {
		return File{}, fmt.Errorf("process file err: %s", err.Error())
	}

	hashedID, err := hashPatientID(patientID, password)
	if err != nil {
		return File{}, fmt.Errorf("hash patient ID err: %s", err.Error())
	}

	anonymizedFileName := fmt.Sprintf("%s_%s%s", hashedID, date, fileType)

	return File{
		Name:    anonymizedFileName,
		Content: anonymizedData,
	}, nil
}

func getFileType(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mwf":
		return ".mwf", nil
	case ".xml":
		return ".xml", nil
	default:
		return "", nil
	}
}

func parseFileName(filename string) (string, string, error) {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	parts := strings.Split(name, "_")
	if len(parts) != 2 {
		return "", "", errFileNameFormat
	}
	return parts[0], parts[1], nil
}

func anonymizeData(
	data []byte,
	fileType string,
) ([]byte, error) {
	switch fileType {
	case ".mwf":
		return mfer.Anonymize(data)
	case ".xml":
		return xml.Anonymize(data)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func hashPatientID(patientID, password string) (string, error) {
	db, err := model.GetDB(os.Getenv("DSN"))
	if err != nil {
		return "", err
	}
	defer db.Close()

	var hashedID string
	err = db.QueryRow("SELECT hashed_id FROM patients WHERE id = ?", patientID).Scan(&hashedID)
	if err == nil {
		// 既存のハッシュIDが見つかった場合、それを返す
		return hashedID, nil
	} else if err != sql.ErrNoRows {
		// sql.ErrNoRows 以外のエラーの場合、エラーを返す
		return "", err
	}

	// 新しいハッシュIDを生成
	newHashedID := sha256.Sum256([]byte(patientID + password))
	hashedIDStr := hex.EncodeToString(newHashedID[:])

	// 新しいハッシュIDをデータベースに保存
	_, err = db.Exec("INSERT INTO patients (id, hashed_id) VALUES (?, ?)", patientID, hashedIDStr)
	if err != nil {
		return "", fmt.Errorf("failed to save hashed ID: %v", err)
	}
	return hashedIDStr, nil
}

func sendZipResponse(c *gin.Context, zipBuffer *bytes.Buffer, conn *websocket.Conn) {

	// 現在の時刻を使用してZIPファイル名を生成
	anonymizedZipFileName := fmt.Sprintf("%s.zip", time.Now().Format("2006-01-02_15-04-05"))

	// メタデータを先に送信（例：ファイル名、サイズなど）
	metaData := map[string]string{
		"fileName": anonymizedZipFileName,
		"fileType": contentTypeZip,
	}
	if err := conn.WriteJSON(metaData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send metadata"})
		return
	}

	// ZIPファイルデータを送信
	if err := conn.WriteMessage(websocket.BinaryMessage, zipBuffer.Bytes()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send ZIP file"})
		return
	}
}
