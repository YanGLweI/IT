package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"it-platform-server/config"
	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
)

// equalUintSlices 比较两个uint切片是否相等
func equalUintSlices(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ============================================================
// 模板管理
// ============================================================

// ListChangeRecordTemplates 获取模板历史版本列表
func ListChangeRecordTemplates(c *gin.Context) {
	var templates []models.ChangeRecordTemplate
	if err := database.GetDB().Order("is_current DESC, created_at DESC").Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": templates})
}

// GetCurrentChangeRecordTemplate 获取当前版本模板
func GetCurrentChangeRecordTemplate(c *gin.Context) {
	var template models.ChangeRecordTemplate
	if err := database.GetDB().Where("is_current = ?", true).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "暂无模板"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": template})
}

// UploadChangeRecordTemplate 上传新版本模板
func UploadChangeRecordTemplate(c *gin.Context) {
	version := c.PostForm("version")
	description := c.PostForm("description")

	if version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "版本号不能为空"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	// 仅允许 docx/doc/pdf
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".docx" && ext != ".doc" && ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 DOCX、DOC、PDF 格式文件"})
		return
	}

	uploadDir := config.Cfg.Upload.ChangeRecordTemplatePath
	os.MkdirAll(uploadDir, 0755)

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	// 事务：将旧版本 is_current 置为 0，插入新版本
	tx := database.GetDB().Begin()

	if err := tx.Model(&models.ChangeRecordTemplate{}).Where("is_current = ?", true).Update("is_current", false).Error; err != nil {
		tx.Rollback()
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新旧版本失败"})
		return
	}

	template := models.ChangeRecordTemplate{
		Version:     version,
		Description: description,
		FileName:    file.Filename,
		FilePath:    filePath,
		FileSize:    file.Size,
		FileType:    file.Header.Get("Content-Type"),
		IsCurrent:   true,
	}

	if err := tx.Create(&template).Error; err != nil {
		tx.Rollback()
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存模板失败"})
		return
	}

	tx.Commit()

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Version", FieldLabel: "版本号", NewValue: version},
		{FieldName: "Description", FieldLabel: "版本说明", NewValue: description},
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: template.FileName},
	}
	services.LogOperation(username, displayName, "上传变更记录表模板", "change_record_template", template.ID, template.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": template})
}

// DownloadChangeRecordTemplate 下载模板文件
func DownloadChangeRecordTemplate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var template models.ChangeRecordTemplate
	if err := database.GetDB().First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "模板不存在"})
		return
	}

	if _, err := os.Stat(template.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=\""+template.FileName+"\"")
	c.Header("Content-Length", fmt.Sprintf("%d", template.FileSize))
	c.File(template.FilePath)
}

// PreviewChangeRecordTemplate 预览模板文件
func PreviewChangeRecordTemplate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var template models.ChangeRecordTemplate
	if err := database.GetDB().First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "模板不存在"})
		return
	}

	if _, err := os.Stat(template.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	ext := strings.ToLower(filepath.Ext(template.FileName))
	switch ext {
	case ".pdf":
		c.Header("Content-Type", "application/pdf")
	case ".docx", ".doc":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	default:
		c.Header("Content-Type", "application/octet-stream")
	}
	c.Header("Content-Disposition", "inline; filename=\""+template.FileName+"\"")
	c.File(template.FilePath)
}

// DeleteChangeRecordTemplate 删除历史版本模板（不可删除当前版本）
func DeleteChangeRecordTemplate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var template models.ChangeRecordTemplate
	if err := database.GetDB().First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "模板不存在"})
		return
	}

	if template.IsCurrent {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不能删除当前版本模板"})
		return
	}

	os.Remove(template.FilePath)

	if err := database.GetDB().Delete(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("change_record_template")
	details := services.DiffStructs(template, models.ChangeRecordTemplate{}, fieldLabels)
	services.LogOperation(username, displayName, "删除变更记录表模板", "change_record_template", template.ID, template.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ============================================================
// 扫描件存档
// ============================================================

// ListChangeRecords 获取变更记录扫描件列表
func ListChangeRecords(c *gin.Context) {
	var records []models.ChangeRecord

	query := database.GetDB().Model(&models.ChangeRecord{})

	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", year)
		}
	}

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("description LIKE ?", "%"+keyword+"%")
	}

	// 按变更类型筛选
	if typeIDStr := c.Query("type_id"); typeIDStr != "" {
		var typeIDs []uint
		for _, idStr := range strings.Split(typeIDStr, ",") {
			if tid, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				typeIDs = append(typeIDs, uint(tid))
			}
		}
		if len(typeIDs) > 0 {
			query = query.Where("id IN (SELECT DISTINCT change_record_id FROM change_record_change_types WHERE change_type_id IN ?)", typeIDs)
		}
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var total int64
	query.Count(&total)

	if err := query.Preload("ChangeTypes").Order("implement_date DESC, year DESC, month DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": records, "total": total, "page_size": pageSize})
}

// parseDateForm 解析表单中的日期字段，空字符串返回nil
func parseDateForm(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateChangeRecord 上传变更记录扫描件
func CreateChangeRecord(c *gin.Context) {
	yearStr := c.PostForm("year")
	monthStr := c.PostForm("month")
	description := c.PostForm("description")

	if yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份和月份不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "月份格式不正确，应为1-12"})
		return
	}

	applyDate, err := parseDateForm(c.PostForm("apply_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "申请日期格式不正确，应为YYYY-MM-DD"})
		return
	}

	implementDate, err := parseDateForm(c.PostForm("implement_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "实施日期格式不正确，应为YYYY-MM-DD"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持PDF格式文件"})
		return
	}

	// 按年份归档
	yearDir := filepath.Join(config.Cfg.Upload.ChangeRecordPath, strconv.Itoa(year))
	os.MkdirAll(yearDir, 0755)

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	record := models.ChangeRecord{
		Year:          year,
		Month:         month,
		Description:   description,
		ApplyDate:     applyDate,
		ImplementDate: implementDate,
		FileName:      file.Filename,
		FilePath:      filePath,
		FileSize:      file.Size,
		FileType:      file.Header.Get("Content-Type"),
	}

	if err := database.GetDB().Create(&record).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存记录失败"})
		return
	}

	// 关联变更类型
	if typeIDs := c.PostForm("type_ids"); typeIDs != "" {
		var types []models.ChangeType
		ids := strings.Split(typeIDs, ",")
		for _, idStr := range ids {
			if tid, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				types = append(types, models.ChangeType{ID: uint(tid)})
			}
		}
		if len(types) > 0 {
			database.GetDB().Model(&record).Association("ChangeTypes").Replace(types)
		}
	}

	// 重新加载关联
	database.GetDB().Preload("ChangeTypes").First(&record, record.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Year", FieldLabel: "年份", NewValue: strconv.Itoa(year)},
		{FieldName: "Month", FieldLabel: "月份", NewValue: strconv.Itoa(month)},
		{FieldName: "Description", FieldLabel: "描述", NewValue: description},
	}
	
	// 添加申请日期
	if applyDate != nil {
		details = append(details, services.LogDetail{
			FieldName:  "ApplyDate",
			FieldLabel: "申请日期",
			OldValue:   "-",
			NewValue:   applyDate.Format("2006-01-02"),
		})
	} else {
		details = append(details, services.LogDetail{
			FieldName:  "ApplyDate",
			FieldLabel: "申请日期",
			OldValue:   "-",
			NewValue:   "-",
		})
	}
	
	// 添加实施日期
	if implementDate != nil {
		details = append(details, services.LogDetail{
			FieldName:  "ImplementDate",
			FieldLabel: "实施日期",
			OldValue:   "-",
			NewValue:   implementDate.Format("2006-01-02"),
		})
	} else {
		details = append(details, services.LogDetail{
			FieldName:  "ImplementDate",
			FieldLabel: "实施日期",
			OldValue:   "-",
			NewValue:   "-",
		})
	}
	
	details = append(details, services.LogDetail{
		FieldName:  "FileName",
		FieldLabel: "文件名",
		NewValue:   record.FileName,
	})
	
	// 添加变更类型信息
	var typeIDs []uint
	for _, ct := range record.ChangeTypes {
		typeIDs = append(typeIDs, ct.ID)
	}
	if len(typeIDs) > 0 {
		details = append(details, services.LogDetail{
			FieldName:  "ChangeTypes",
			FieldLabel: "变更类型",
			OldValue:   "[]",
			NewValue:   fmt.Sprintf("%v", typeIDs),
		})
	}
	
	services.LogOperation(username, displayName, "上传变更记录扫描件", "change_record", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": record})
}

// UpdateChangeRecord 更新变更记录扫描件元数据
func UpdateChangeRecord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.ChangeRecord
	// 预加载 ChangeTypes 关联，用于记录旧值
	if err := database.GetDB().Preload("ChangeTypes").First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	yearStr := c.PostForm("year")
	monthStr := c.PostForm("month")
	description := c.PostForm("description")

	if yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份和月份不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "月份格式不正确，应为1-12"})
		return
	}

	applyDate, err := parseDateForm(c.PostForm("apply_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "申请日期格式不正确，应为YYYY-MM-DD"})
		return
	}

	implementDate, err := parseDateForm(c.PostForm("implement_date"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "实施日期格式不正确，应为YYYY-MM-DD"})
		return
	}

	oldRecord := record
	oldYear := record.Year
	oldFilePath := record.FilePath
	// 保存旧的变更类型ID列表，用于日志记录
	var oldTypeIDs []uint
	for _, ct := range record.ChangeTypes {
		oldTypeIDs = append(oldTypeIDs, ct.ID)
	}

	// 年份变化时移动文件
	if year != oldYear {
		oldYearDir := filepath.Join(config.Cfg.Upload.ChangeRecordPath, strconv.Itoa(oldYear))
		newYearDir := filepath.Join(config.Cfg.Upload.ChangeRecordPath, strconv.Itoa(year))
		os.MkdirAll(newYearDir, 0755)

		fileName := filepath.Base(record.FilePath)
		newFilePath := filepath.Join(newYearDir, fileName)
		if err := os.Rename(oldFilePath, newFilePath); err != nil {
			if copyErr := services.CopyFile(oldFilePath, newFilePath); copyErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件移动失败"})
				return
			}
			os.Remove(oldFilePath)
		}
		record.FilePath = newFilePath
		os.Remove(oldYearDir)
	}

	record.Year = year
	record.Month = month
	record.Description = description
	record.ApplyDate = applyDate
	record.ImplementDate = implementDate

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 更新变更类型关联
	if typeIDs := c.PostForm("type_ids"); typeIDs != "" {
		var types []models.ChangeType
		ids := strings.Split(typeIDs, ",")
		for _, idStr := range ids {
			if tid, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
				types = append(types, models.ChangeType{ID: uint(tid)})
			}
		}
		if len(types) > 0 {
			database.GetDB().Model(&record).Association("ChangeTypes").Replace(types)
		}
	} else {
		// type_ids 为空字符串时清除关联
		database.GetDB().Model(&record).Association("ChangeTypes").Clear()
	}

	// 重新加载关联
	database.GetDB().Preload("ChangeTypes").First(&record, record.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("change_record")
	details := services.DiffStructs(oldRecord, record, fieldLabels)
	
	// 检查变更类型是否发生变化
	var newTypeIDs []uint
	for _, ct := range record.ChangeTypes {
		newTypeIDs = append(newTypeIDs, ct.ID)
	}
	
	if !equalUintSlices(oldTypeIDs, newTypeIDs) {
		oldStr := fmt.Sprintf("%v", oldTypeIDs)
		newStr := fmt.Sprintf("%v", newTypeIDs)
		details = append(details, services.LogDetail{
			FieldName:  "ChangeTypes",
			FieldLabel: "变更类型",
			OldValue:   oldStr,
			NewValue:   newStr,
		})
	}
	
	services.LogOperation(username, displayName, "更新变更记录扫描件", "change_record", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DeleteChangeRecord 删除变更记录扫描件
func DeleteChangeRecord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.ChangeRecord
	// 预加载 ChangeTypes 关联，用于记录旧值
	if err := database.GetDB().Preload("ChangeTypes").First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	os.Remove(record.FilePath)

	// 保存旧的变更类型ID列表，用于日志记录
	var oldTypeIDs []uint
	for _, ct := range record.ChangeTypes {
		oldTypeIDs = append(oldTypeIDs, ct.ID)
	}

	if err := database.GetDB().Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("change_record")
	details := services.DiffStructs(record, models.ChangeRecord{}, fieldLabels)
	
	// 添加变更类型信息
	if len(oldTypeIDs) > 0 {
		details = append(details, services.LogDetail{
			FieldName:  "ChangeTypes",
			FieldLabel: "变更类型",
			OldValue:   fmt.Sprintf("%v", oldTypeIDs),
			NewValue:   "[]",
		})
	}
	
	services.LogOperation(username, displayName, "删除变更记录扫描件", "change_record", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PreviewChangeRecord 预览变更记录扫描件
func PreviewChangeRecord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.ChangeRecord
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\""+record.FileName+"\"")
	c.File(record.FilePath)
}

// DownloadChangeRecord 下载变更记录扫描件
func DownloadChangeRecord(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.ChangeRecord
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=\""+record.FileName+"\"")
	c.Header("Content-Length", fmt.Sprintf("%d", record.FileSize))
	c.File(record.FilePath)
}
