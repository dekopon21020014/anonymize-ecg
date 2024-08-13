package controller

import (
	"fmt"
	"net/http"
	"os"

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
	c.Data(http.StatusOK, "text/csv", fileStruct.Content)
}
