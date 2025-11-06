package services

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"go_market_email/internal/models"
	"gorm.io/gorm"
)

type TemplateService struct {
	db *gorm.DB
}

func NewTemplateService(db *gorm.DB) *TemplateService {
	return &TemplateService{db: db}
}

// ExtractVariables 从模板中提取变量
func (s *TemplateService) ExtractVariables(content string) []string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(content, -1)
	
	variables := make([]string, 0)
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 {
			variable := strings.TrimSpace(match[1])
			if !seen[variable] {
				variables = append(variables, variable)
				seen[variable] = true
			}
		}
	}
	
	return variables
}

// CreateTemplate 创建邮件模板
func (s *TemplateService) CreateTemplate(template *models.EmailTemplate) error {
	// 提取变量
	variables := s.ExtractVariables(template.Content + " " + template.Subject)
	variablesJSON, _ := json.Marshal(variables)
	template.Variables = string(variablesJSON)
	
	// 设置版本号
	var maxVersion int
	s.db.Model(&models.EmailTemplate{}).
		Where("name = ? AND user_id = ? AND project_id = ?", 
			template.Name, template.UserID, template.ProjectID).
		Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)
	
	template.Version = maxVersion + 1
	
	return s.db.Create(template).Error
}

// UpdateTemplate 更新模板（创建新版本）
func (s *TemplateService) UpdateTemplate(id uint, updates *models.EmailTemplate) error {
	var existing models.EmailTemplate
	if err := s.db.First(&existing, id).Error; err != nil {
		return err
	}
	
	// 创建新版本
	newTemplate := models.EmailTemplate{
		Name:      existing.Name,
		Subject:   updates.Subject,
		Content:   updates.Content,
		UserID:    existing.UserID,
		ProjectID: existing.ProjectID,
		Version:   existing.Version + 1,
		Status:    "active",
	}
	
	// 提取变量
	variables := s.ExtractVariables(newTemplate.Content + " " + newTemplate.Subject)
	variablesJSON, _ := json.Marshal(variables)
	newTemplate.Variables = string(variablesJSON)
	
	// 停用旧版本
	s.db.Model(&existing).Update("status", "inactive")
	
	return s.db.Create(&newTemplate).Error
}

// GetTemplate 获取模板
func (s *TemplateService) GetTemplate(id uint) (*models.EmailTemplate, error) {
	var template models.EmailTemplate
	err := s.db.First(&template, id).Error
	return &template, err
}

// ListTemplates 获取模板列表
func (s *TemplateService) ListTemplates(userID, projectID uint, page, pageSize int) ([]models.EmailTemplate, int64, error) {
	var templates []models.EmailTemplate
	var total int64
	
	query := s.db.Model(&models.EmailTemplate{}).
		Where("user_id = ? AND project_id = ? AND status = ?", userID, projectID, "active")
	
	query.Count(&total)
	
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").Find(&templates).Error
	
	return templates, total, err
}

// GetTemplateVersions 获取模板版本历史
func (s *TemplateService) GetTemplateVersions(name string, userID, projectID uint) ([]models.EmailTemplate, error) {
	var templates []models.EmailTemplate
	err := s.db.Where("name = ? AND user_id = ? AND project_id = ?", 
		name, userID, projectID).
		Order("version DESC").Find(&templates).Error
	
	return templates, err
}

// DeleteTemplate 删除模板
func (s *TemplateService) DeleteTemplate(id uint) error {
	return s.db.Delete(&models.EmailTemplate{}, id).Error
}

// ReplaceVariables 替换模板变量
func (s *TemplateService) ReplaceVariables(template string, data map[string]interface{}) string {
	result := template
	
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		if valueStr, ok := value.(string); ok {
			result = strings.ReplaceAll(result, placeholder, valueStr)
		}
	}
	
	return result
}

// ValidateTemplate 验证模板
func (s *TemplateService) ValidateTemplate(template *models.EmailTemplate) error {
	if template.Name == "" {
		return errors.New("模板名称不能为空")
	}
	if template.Subject == "" {
		return errors.New("邮件主题不能为空")
	}
	if template.Content == "" {
		return errors.New("邮件内容不能为空")
	}
	
	// 检查模板大小
	contentSize := len([]byte(template.Content))
	if contentSize > 50*1024*1024 { // 50MB
		return errors.New("模板内容超过大小限制")
	}
	
	return nil
}