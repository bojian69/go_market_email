package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	
	"github.com/go-redis/redis/v8"
	"github.com/xuri/excelize/v2"
	"go_market_email/internal/utils"
	"gorm.io/gorm"
)

type DataService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewDataService(db *gorm.DB, rdb *redis.Client) *DataService {
	return &DataService{db: db, rdb: rdb}
}

// ImportExcelData 导入Excel数据
func (s *DataService) ImportExcelData(filePath string, taskID uint) ([]map[string]interface{}, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	
	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel文件至少需要包含标题行和数据行")
	}
	
	// 第一行作为标题
	headers := rows[0]
	var data []map[string]interface{}
	
	// 处理数据行
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		record := make(map[string]interface{})
		
		for j, header := range headers {
			if j < len(row) {
				record[header] = row[j]
			} else {
				record[header] = ""
			}
		}
		data = append(data, record)
	}
	
	// 存储到Redis
	key := utils.EmailDataKey + strconv.Itoa(int(taskID))
	dataJSON, _ := json.Marshal(data)
	
	ctx := context.Background()
	err = s.rdb.Set(ctx, key, dataJSON, 24*time.Hour).Err()
	if err != nil {
		return nil, err
	}
	
	return data, nil
}

// ImportCSVData 导入CSV数据
func (s *DataService) ImportCSVData(content string, taskID uint) ([]map[string]interface{}, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("CSV文件至少需要包含标题行和数据行")
	}
	
	// 解析标题行
	headers := strings.Split(lines[0], ",")
	for i, header := range headers {
		headers[i] = strings.TrimSpace(header)
	}
	
	var data []map[string]interface{}
	
	// 处理数据行
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		
		values := strings.Split(line, ",")
		record := make(map[string]interface{})
		
		for j, header := range headers {
			if j < len(values) {
				record[header] = strings.TrimSpace(values[j])
			} else {
				record[header] = ""
			}
		}
		data = append(data, record)
	}
	
	// 存储到Redis
	key := utils.EmailDataKey + strconv.Itoa(int(taskID))
	dataJSON, _ := json.Marshal(data)
	
	ctx := context.Background()
	err := s.rdb.Set(ctx, key, dataJSON, 24*time.Hour).Err()
	if err != nil {
		return nil, err
	}
	
	return data, nil
}

// ExecuteSQLQuery 执行SQL查询
func (s *DataService) ExecuteSQLQuery(query string, taskID uint) ([]map[string]interface{}, error) {
	// 安全检查：只允许SELECT和INSERT语句
	upperQuery := strings.ToUpper(strings.TrimSpace(query))
	if !strings.HasPrefix(upperQuery, "SELECT") && !strings.HasPrefix(upperQuery, "INSERT") {
		return nil, fmt.Errorf("只允许执行SELECT和INSERT语句")
	}
	
	// 禁止危险操作
	dangerousKeywords := []string{"DROP", "DELETE", "UPDATE", "TRUNCATE", "ALTER"}
	for _, keyword := range dangerousKeywords {
		if strings.Contains(upperQuery, keyword) {
			return nil, fmt.Errorf("查询包含危险关键字: %s", keyword)
		}
	}
	
	sqlDB, err := s.db.DB()
	if err != nil {
		return nil, err
	}
	
	rows, err := sqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	
	var data []map[string]interface{}
	
	// 处理查询结果
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		
		record := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			
			// 处理不同数据类型
			if b, ok := val.([]byte); ok {
				record[col] = string(b)
			} else if val == nil {
				record[col] = ""
			} else {
				record[col] = val
			}
		}
		
		data = append(data, record)
	}
	
	// 存储到Redis
	key := utils.EmailDataKey + strconv.Itoa(int(taskID))
	dataJSON, _ := json.Marshal(data)
	
	ctx := context.Background()
	err = s.rdb.Set(ctx, key, dataJSON, 24*time.Hour).Err()
	if err != nil {
		return nil, err
	}
	
	return data, nil
}

// SaveManualData 保存手动输入的数据
func (s *DataService) SaveManualData(data []map[string]interface{}, taskID uint) error {
	key := utils.EmailDataKey + strconv.Itoa(int(taskID))
	dataJSON, _ := json.Marshal(data)
	
	ctx := context.Background()
	return s.rdb.Set(ctx, key, dataJSON, 24*time.Hour).Err()
}

// GetTaskData 获取任务数据
func (s *DataService) GetTaskData(taskID uint) ([]map[string]interface{}, error) {
	key := utils.EmailDataKey + strconv.Itoa(int(taskID))
	
	ctx := context.Background()
	dataJSON, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []map[string]interface{}{}, nil
		}
		return nil, err
	}
	
	var data []map[string]interface{}
	err = json.Unmarshal([]byte(dataJSON), &data)
	return data, err
}

// DeleteTaskData 删除任务数据
func (s *DataService) DeleteTaskData(taskID uint) error {
	key := utils.EmailDataKey + strconv.Itoa(int(taskID))
	
	ctx := context.Background()
	return s.rdb.Del(ctx, key).Err()
}

// ValidateDataStructure 验证数据结构
func (s *DataService) ValidateDataStructure(data []map[string]interface{}, requiredFields []string) error {
	if len(data) == 0 {
		return fmt.Errorf("数据不能为空")
	}
	
	// 检查必需字段
	for _, field := range requiredFields {
		found := false
		for key := range data[0] {
			if key == field {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("缺少必需字段: %s", field)
		}
	}
	
	return nil
}