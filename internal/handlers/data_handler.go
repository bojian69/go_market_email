package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go_market_email/internal/services"
)

type DataHandler struct {
	dataService *services.DataService
}

func NewDataHandler(dataService *services.DataService) *DataHandler {
	return &DataHandler{dataService: dataService}
}

// UploadFile 上传文件
func (h *DataHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败"})
		return
	}

	taskIDStr := c.PostForm("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	// 保存文件到临时目录
	tempPath := "/tmp/" + file.Filename
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 根据文件类型处理
	var data []map[string]interface{}
	if file.Header.Get("Content-Type") == "text/csv" {
		// 处理CSV文件
		content := ""
		// 读取文件内容...
		data, err = h.dataService.ImportCSVData(content, uint(taskID))
	} else {
		// 处理Excel文件
		data, err = h.dataService.ImportExcelData(tempPath, uint(taskID))
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// ExecuteSQL 执行SQL查询
func (h *DataHandler) ExecuteSQL(c *gin.Context) {
	var request struct {
		Query  string `json:"query" binding:"required"`
		TaskID uint   `json:"task_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.dataService.ExecuteSQLQuery(request.Query, request.TaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// SaveManualData 保存手动输入的数据
func (h *DataHandler) SaveManualData(c *gin.Context) {
	var request struct {
		TaskID      uint                     `json:"task_id" binding:"required"`
		Data        []map[string]interface{} `json:"data" binding:"required"`
		Description string                   `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.dataService.SaveManualData(request.Data, request.TaskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "数据保存成功"})
}