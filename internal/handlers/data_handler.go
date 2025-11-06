package handlers

import (
	"net/http"
	"os"
	"strconv"
	"strings"

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

	// 检查文件大小 (50MB)
	if file.Size > 50*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小不能超过50MB"})
		return
	}

	// 生成临时任务ID
	taskID := uint(1)
	if taskIDStr := c.PostForm("task_id"); taskIDStr != "" {
		if id, err := strconv.ParseUint(taskIDStr, 10, 32); err == nil {
			taskID = uint(id)
		}
	}

	// 保存文件到临时目录
	tempPath := "/tmp/" + file.Filename
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}

	// 根据文件扩展名处理
	var data []map[string]interface{}
	filename := strings.ToLower(file.Filename)
	if strings.HasSuffix(filename, ".csv") {
		// 读取CSV文件内容
		content, err := os.ReadFile(tempPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
			return
		}
		data, err = h.dataService.ImportCSVData(string(content), taskID)
	} else {
		// 处理Excel文件
		data, err = h.dataService.ImportExcelData(tempPath, taskID)
	}

	// 清理临时文件
	os.Remove(tempPath)

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