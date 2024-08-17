package controller

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shikidalab/anonymize-ecg/mfer"
	"github.com/shikidalab/anonymize-ecg/model"
	"github.com/shikidalab/anonymize-ecg/xml"
)

func GetTop(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to LP"})
}

const (
	passwordMismatchErr   = "passwords do not match"
	noFilesFoundErr       = "no files found"
	fileNameFormatErr     = "file name format is incorrect"
	zipCreationErr        = "failed to create ZIP file"
	fileWriteErr          = "failed to write file"
	contentTypeZip        = "application/zip"
	contentDispositionFmt = "attachment; filename=%s"
)

type File struct {
	Name    string
	Content []byte
}

func AnonymizeECG(c *gin.Context) {
	password, err := validatePasswords(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	files, err := getFilesFromForm(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	anonymizedFiles, err := processFiles(files, password)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	zipBuffer, err := createZipFile(anonymizedFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sendZipResponse(c, zipBuffer)
}

func validatePasswords(c *gin.Context) (string, error) {
	password := c.PostForm("password")
	passwordConfirmation := c.PostForm("passwordConfirmation")

	if password == "" || passwordConfirmation == "" {
		return "", fmt.Errorf("both password and password confirmation are required")
	}

	if password != passwordConfirmation {
		return "", fmt.Errorf(passwordMismatchErr)
	}

	return password, nil
}

func getFilesFromForm(c *gin.Context) ([]File, error) {
	// フォームからZIPファイルを取得
	file, err := c.FormFile("zipfile")
	if err != nil {
		return nil, fmt.Errorf("failed to get uploaded file: %v", err)
	}

	// ZIPファイルを開く
	srcFile, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer srcFile.Close()

	// ZIPファイルの内容をメモリ上に読み込む
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, srcFile); err != nil {
		return nil, fmt.Errorf("failed to read uploaded file: %v", err)
	}

	// メモリ上に読み込んだデータをZIPリーダーに渡す
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %v", err)
	}

	// ZIP内のファイルをメモリ上で解凍
	var files []File
	for _, f := range zipReader.File {
		// フォルダはスキップ
		if f.FileInfo().IsDir() {
			continue
		}

		// ファイルを開いて内容を読み込む
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file in zip: %v", err)
		}
		defer rc.Close()

		// ファイル内容をメモリに読み込む
		var fileBuf bytes.Buffer
		if _, err := io.Copy(&fileBuf, rc); err != nil {
			return nil, fmt.Errorf("failed to read file in zip: %v", err)
		}

		// ファイル名をキーとして内容をバイトスライスに保存
		files = append(files, File{
			Name:    f.Name,
			Content: fileBuf.Bytes(),
		})
	}

	return files, nil
}

func processFiles(files []File, password string) ([]File, error) {
	var anonymizedFiles []File

	for _, file := range files {
		anonymizedFile, err := processFile(file, password)
		if err != nil {
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
		return "", "", fmt.Errorf(fileNameFormatErr)
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
	fmt.Printf("%s, %s, %s\n", patientID, password, hashedIDStr)

	// 新しいハッシュIDをデータベースに保存
	_, err = db.Exec("INSERT INTO patients (id, hashed_id) VALUES (?, ?)", patientID, hashedIDStr)
	if err != nil {
		return "", fmt.Errorf("failed to save hashed ID: %v", err)
	}
	return hashedIDStr, nil
}

func createZipFile(anonymizedFiles []File) (*bytes.Buffer, error) {
	bufForResponse := new(bytes.Buffer)
	zipWriter := zip.NewWriter(bufForResponse)
	defer zipWriter.Close()

	for _, file := range anonymizedFiles {
		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", zipCreationErr, err)
		}
		_, err = zipFile.Write(file.Content)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", fileWriteErr, err)
		}
	}

	return bufForResponse, nil
}

func sendZipResponse(c *gin.Context, zipBuffer *bytes.Buffer) {
	anonymizedZipFileName := fmt.Sprintf("%s.zip", time.Now().Format("2006-01-02_15-04-05"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", anonymizedZipFileName))
	c.Header("Content-Type", contentTypeZip)
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Data(http.StatusOK, contentTypeZip, zipBuffer.Bytes())
}
