package controller

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	anonymizedZipFileName = "anonymized-files.zip"
	contentTypeZip        = "application/zip"
	contentDispositionFmt = "attachment; filename=%s"
)

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

func getFilesFromForm(c *gin.Context) (map[string][]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("get form err: %s", err.Error())
	}

	files := form.File
	if len(files) == 0 {
		return nil, fmt.Errorf(noFilesFoundErr)
	}

	return files, nil
}

func processFiles(files map[string][]*multipart.FileHeader, password string) ([]struct {
	Name    string
	Content []byte
}, error) {
	var anonymizedFiles []struct {
		Name    string
		Content []byte
	}

	for _, fileHeaders := range files {
		for _, fileHeader := range fileHeaders {
			anonymizedFile, err := processFile(fileHeader, password)
			if err != nil {
				return nil, err
			}
			if anonymizedFile != nil {
				anonymizedFiles = append(anonymizedFiles, *anonymizedFile)
			}
		}
	}

	return anonymizedFiles, nil
}

func processFile(fileHeader *multipart.FileHeader, password string) (*struct {
	Name    string
	Content []byte
}, error) {
	filename := fileHeader.Filename
	fileType, err := getFileType(filename)
	if err != nil {
		return nil, err
	}

	if fileType == "" {
		return nil, nil // Skip non-MWF and non-XML files
	}

	patientID, date, err := parseFileName(filename)
	if err != nil {
		return nil, err
	}

	data, err := readFileContent(fileHeader)
	if err != nil {
		return nil, err
	}

	anonymizedData, err := anonymizeData(data, fileType)
	if err != nil {
		return nil, fmt.Errorf("process file err: %s", err.Error())
	}

	hashedID, err := hashPatientID(patientID, password)
	if err != nil {
		return nil, fmt.Errorf("hash patient ID err: %s", err.Error())
	}

	anonymizedFileName := fmt.Sprintf("%s_%s%s", hashedID, date, fileType)

	return &struct {
		Name    string
		Content []byte
	}{
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

func readFileContent(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open file err: %s", err.Error())
	}
	defer file.Close()

	return io.ReadAll(file)
}

func anonymizeData(data []byte, fileType string) ([]byte, error) {
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

func createZipFile(anonymizedFiles []struct {
	Name    string
	Content []byte
}) (*bytes.Buffer, error) {
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
	c.Header("Content-Disposition", fmt.Sprintf(contentDispositionFmt, anonymizedZipFileName))
	c.Header("Content-Type", contentTypeZip)
	c.Data(http.StatusOK, contentTypeZip, zipBuffer.Bytes())
}
