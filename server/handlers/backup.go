package handlers

import (
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

// ListBackups 获取备份记录列表
func ListBackups(c *gin.Context) {
	var records []models.BackupRecord

	query := database.GetDB().Model(&models.BackupRecord{})

	// 按年份筛选（从 application_date 提取年份）
	if yearStr := c.Query("year"); yearStr != "" {
		query = query.Where("application_date LIKE ?", yearStr+"%")
	}

	// 按部门筛选
	if deptIDStr := c.Query("department_id"); deptIDStr != "" {
		if deptID, err := strconv.Atoi(deptIDStr); err == nil {
			query = query.Where("department_id = ?", deptID)
		}
	}

	// 按备份工具筛选
	if tool := c.Query("backup_tool"); tool != "" {
		query = query.Where("backup_tool = ?", tool)
	}

	// 按关键词搜索
	if keyword := c.Query("keyword"); keyword != "" {
		kw := "%" + keyword + "%"
		kwLower := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("backup_target LIKE ? OR backup_tool LIKE ? OR "+
			"backup_source_asset_id IN (SELECT id FROM assets WHERE LOWER(computer_name) LIKE ?) OR "+
			"backup_medium_asset_id IN (SELECT id FROM assets WHERE LOWER(computer_name) LIKE ?)",
			kw, kw, kwLower, kwLower)
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

	if err := query.
		Preload("BackupSourceAsset").
		Preload("BackupMediumAsset").
		Preload("Department").
		Preload("Recoveries").
		Order("application_date DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": records, "total": total, "page_size": pageSize})
}

// CreateBackup 创建备份记录
func CreateBackup(c *gin.Context) {
	applicationDate := c.PostForm("application_date")
	backupSourceAssetIDStr := c.PostForm("backup_source_asset_id")
	backupTargetType := c.PostForm("backup_target_type")
	backupTarget := c.PostForm("backup_target")
	backupTool := c.PostForm("backup_tool")
	backupMediumAssetIDStr := c.PostForm("backup_medium_asset_id")
	backupFrequency := c.PostForm("backup_frequency")
	retentionPolicy := c.PostForm("retention_policy")
	fullBackupStrategy := c.PostForm("full_backup_strategy")
	departmentIDStr := c.PostForm("department_id")

	// 验证必填字段
	if applicationDate == "" || backupSourceAssetIDStr == "" || backupMediumAssetIDStr == "" || departmentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "申请日期、备份源、备份介质和所属部门不能为空"})
		return
	}

	backupSourceAssetID, err := strconv.Atoi(backupSourceAssetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "备份源资产ID格式不正确"})
		return
	}

	backupMediumAssetID, err := strconv.Atoi(backupMediumAssetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "备份介质资产ID格式不正确"})
		return
	}

	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部门ID格式不正确"})
		return
	}

	// 验证资产是否存在
	var sourceAsset models.Asset
	if err := database.GetDB().First(&sourceAsset, backupSourceAssetID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "选择的备份源资产不存在"})
		return
	}

	var mediumAsset models.Asset
	if err := database.GetDB().First(&mediumAsset, backupMediumAssetID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "选择的备份介质资产不存在"})
		return
	}

	var department models.Department
	if err := database.GetDB().First(&department, departmentID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "选择的部门不存在"})
		return
	}

	// 备份对象：类型为"系统"或"配置文件"时存固定值
	if backupTargetType == "系统" {
		backupTarget = "系统"
	} else if backupTargetType == "配置文件" {
		backupTarget = "配置文件"
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传申请表文件"})
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 PDF 格式文件"})
		return
	}

	// 从申请日期提取年份
	year := time.Now().Format("2006")
	if len(applicationDate) >= 4 {
		year = applicationDate[:4]
	}

	// 构建上传路径
	yearDir := filepath.Join(config.Cfg.Upload.BackupPath, year)
	if err := os.MkdirAll(yearDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "目录创建失败"})
		return
	}

	// 生成唯一文件名
	filename := formatFileName(file.Filename, ext)
	filePath := filepath.Join(yearDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	record := models.BackupRecord{
		ApplicationDate:     applicationDate,
		BackupSourceAssetID: uint(backupSourceAssetID),
		BackupTargetType:    backupTargetType,
		BackupTarget:        backupTarget,
		BackupTool:          backupTool,
		BackupMediumAssetID: uint(backupMediumAssetID),
		BackupFrequency:     backupFrequency,
		RetentionPolicy:     retentionPolicy,
		FullBackupStrategy:  fullBackupStrategy,
		DepartmentID:        uint(departmentID),
		FileName:            file.Filename,
		FilePath:            filePath,
		FileSize:            file.Size,
		FileType:            file.Header.Get("Content-Type"),
	}

	if err := database.GetDB().Create(&record).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存记录失败"})
		return
	}

	database.GetDB().Preload("BackupSourceAsset").Preload("BackupMediumAsset").Preload("Department").First(&record, record.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "ApplicationDate", FieldLabel: "申请日期", NewValue: applicationDate},
		{FieldName: "BackupSourceAssetID", FieldLabel: "备份源", NewValue: sourceAsset.ComputerName},
		{FieldName: "BackupTargetType", FieldLabel: "备份对象类型", NewValue: backupTargetType},
		{FieldName: "BackupTarget", FieldLabel: "备份对象", NewValue: backupTarget},
		{FieldName: "BackupTool", FieldLabel: "备份工具", NewValue: backupTool},
		{FieldName: "BackupMediumAssetID", FieldLabel: "备份介质", NewValue: mediumAsset.ComputerName},
		{FieldName: "BackupFrequency", FieldLabel: "备份频率", NewValue: backupFrequency},
		{FieldName: "RetentionPolicy", FieldLabel: "保留策略", NewValue: retentionPolicy},
		{FieldName: "FullBackupStrategy", FieldLabel: "全量备份策略", NewValue: fullBackupStrategy},
		{FieldName: "DepartmentID", FieldLabel: "所属部门", NewValue: department.Name},
		{FieldName: "FileName", FieldLabel: "申请表", NewValue: file.Filename},
	}

	services.LogOperation(username, displayName, "创建备份记录", "backup", record.ID, file.Filename, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": record})
}

// UpdateBackup 更新备份记录
func UpdateBackup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.BackupRecord
	// 初始加载不Preload关联，避免GORM Save时覆盖外键
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	applicationDate := c.PostForm("application_date")
	backupSourceAssetIDStr := c.PostForm("backup_source_asset_id")
	backupTargetType := c.PostForm("backup_target_type")
	backupTarget := c.PostForm("backup_target")
	backupTool := c.PostForm("backup_tool")
	backupMediumAssetIDStr := c.PostForm("backup_medium_asset_id")
	backupFrequency := c.PostForm("backup_frequency")
	retentionPolicy := c.PostForm("retention_policy")
	fullBackupStrategy := c.PostForm("full_backup_strategy")
	departmentIDStr := c.PostForm("department_id")

	// 记录旧值用于日志
	oldApplicationDate := record.ApplicationDate
	oldBackupSourceAssetID := record.BackupSourceAssetID
	oldBackupTargetType := record.BackupTargetType
	oldBackupTarget := record.BackupTarget
	oldBackupTool := record.BackupTool
	oldBackupMediumAssetID := record.BackupMediumAssetID
	oldBackupFrequency := record.BackupFrequency
	oldRetentionPolicy := record.RetentionPolicy
	oldFullBackupStrategy := record.FullBackupStrategy
	oldDepartmentID := record.DepartmentID
	oldFilePath := record.FilePath

	// 更新申请日期（年份变更时迁移文件）
	if applicationDate != "" && applicationDate != record.ApplicationDate {
		newYear := applicationDate[:4]
		oldYear := record.ApplicationDate[:4]
		if newYear != oldYear {
			if record.FilePath != "" {
				newYearDir := filepath.Join(config.Cfg.Upload.BackupPath, newYear)
				if err := os.MkdirAll(newYearDir, 0755); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "目录创建失败"})
					return
				}
				fileName := filepath.Base(record.FilePath)
				newFilePath := filepath.Join(newYearDir, fileName)
				if err := os.Rename(record.FilePath, newFilePath); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件迁移失败"})
					return
				}
				record.FilePath = newFilePath
			}
		}
		record.ApplicationDate = applicationDate
	}

	if backupSourceAssetIDStr != "" {
		if aid, err := strconv.Atoi(backupSourceAssetIDStr); err == nil {
			var asset models.Asset
			if err := database.GetDB().First(&asset, aid).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "选择的备份源资产不存在"})
				return
			}
			record.BackupSourceAssetID = uint(aid)
			record.BackupSourceAsset = models.Asset{}
		}
	}

	if backupTargetType != "" {
		record.BackupTargetType = backupTargetType
		if backupTargetType == "系统" {
			record.BackupTarget = "系统"
		} else if backupTargetType == "配置文件" {
			record.BackupTarget = "配置文件"
		} else if backupTarget != "" {
			record.BackupTarget = backupTarget
		}
	}

	if backupTool != "" {
		record.BackupTool = backupTool
	}

	if backupMediumAssetIDStr != "" {
		if aid, err := strconv.Atoi(backupMediumAssetIDStr); err == nil {
			var asset models.Asset
			if err := database.GetDB().First(&asset, aid).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "选择的备份介质资产不存在"})
				return
			}
			record.BackupMediumAssetID = uint(aid)
			record.BackupMediumAsset = models.Asset{}
		}
	}

	if backupFrequency != "" {
		record.BackupFrequency = backupFrequency
	}
	if retentionPolicy != "" {
		record.RetentionPolicy = retentionPolicy
	}
	if fullBackupStrategy != "" {
		record.FullBackupStrategy = fullBackupStrategy
	}

	if departmentIDStr != "" {
		if did, err := strconv.Atoi(departmentIDStr); err == nil {
			var dept models.Department
			if err := database.GetDB().First(&dept, did).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "选择的部门不存在"})
				return
			}
			record.DepartmentID = uint(did)
			record.Department = models.Department{}
		}
	}

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	database.GetDB().Preload("BackupSourceAsset").Preload("BackupMediumAsset").Preload("Department").First(&record, record.ID)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	var details []services.LogDetail

	if oldApplicationDate != record.ApplicationDate {
		details = append(details, services.LogDetail{FieldName: "ApplicationDate", FieldLabel: "申请日期", OldValue: oldApplicationDate, NewValue: record.ApplicationDate})
	}
	if oldBackupSourceAssetID != record.BackupSourceAssetID {
		var oldAsset models.Asset
		database.GetDB().First(&oldAsset, oldBackupSourceAssetID)
		details = append(details, services.LogDetail{FieldName: "BackupSourceAssetID", FieldLabel: "备份源", OldValue: oldAsset.ComputerName, NewValue: record.BackupSourceAsset.ComputerName})
	}
	if oldBackupTargetType != record.BackupTargetType {
		details = append(details, services.LogDetail{FieldName: "BackupTargetType", FieldLabel: "备份对象类型", OldValue: oldBackupTargetType, NewValue: record.BackupTargetType})
	}
	if oldBackupTarget != record.BackupTarget {
		details = append(details, services.LogDetail{FieldName: "BackupTarget", FieldLabel: "备份对象", OldValue: oldBackupTarget, NewValue: record.BackupTarget})
	}
	if oldBackupTool != record.BackupTool {
		details = append(details, services.LogDetail{FieldName: "BackupTool", FieldLabel: "备份工具", OldValue: oldBackupTool, NewValue: record.BackupTool})
	}
	if oldBackupMediumAssetID != record.BackupMediumAssetID {
		var oldAsset models.Asset
		database.GetDB().First(&oldAsset, oldBackupMediumAssetID)
		details = append(details, services.LogDetail{FieldName: "BackupMediumAssetID", FieldLabel: "备份介质", OldValue: oldAsset.ComputerName, NewValue: record.BackupMediumAsset.ComputerName})
	}
	if oldBackupFrequency != record.BackupFrequency {
		details = append(details, services.LogDetail{FieldName: "BackupFrequency", FieldLabel: "备份频率", OldValue: oldBackupFrequency, NewValue: record.BackupFrequency})
	}
	if oldRetentionPolicy != record.RetentionPolicy {
		details = append(details, services.LogDetail{FieldName: "RetentionPolicy", FieldLabel: "保留策略", OldValue: oldRetentionPolicy, NewValue: record.RetentionPolicy})
	}
	if oldFullBackupStrategy != record.FullBackupStrategy {
		details = append(details, services.LogDetail{FieldName: "FullBackupStrategy", FieldLabel: "全量备份策略", OldValue: oldFullBackupStrategy, NewValue: record.FullBackupStrategy})
	}
	if oldDepartmentID != record.DepartmentID {
		var oldDept models.Department
		database.GetDB().First(&oldDept, oldDepartmentID)
		details = append(details, services.LogDetail{FieldName: "DepartmentID", FieldLabel: "所属部门", OldValue: oldDept.Name, NewValue: record.Department.Name})
	}
	if oldFilePath != record.FilePath {
		details = append(details, services.LogDetail{FieldName: "FilePath", FieldLabel: "申请表路径", OldValue: filepath.Base(oldFilePath), NewValue: filepath.Base(record.FilePath)})
	}

	services.LogOperation(username, displayName, "更新备份记录", "backup", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DeleteBackup 删除备份记录
func DeleteBackup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.BackupRecord
	if err := database.GetDB().Preload("BackupSourceAsset").Preload("BackupMediumAsset").Preload("Department").First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	// 删除申请表文件
	if record.FilePath != "" {
		os.Remove(record.FilePath)
	}

	// 删除关联的恢复记录及其文件
	var recoveries []models.BackupRecovery
	database.GetDB().Where("backup_record_id = ?", record.ID).Find(&recoveries)
	for _, r := range recoveries {
		if r.FilePath != "" {
			os.Remove(r.FilePath)
		}
	}
	database.GetDB().Where("backup_record_id = ?", record.ID).Delete(&models.BackupRecovery{})

	// 软删除备份记录
	if err := database.GetDB().Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "ApplicationDate", FieldLabel: "申请日期", OldValue: record.ApplicationDate, NewValue: ""},
		{FieldName: "BackupSourceAssetID", FieldLabel: "备份源", OldValue: record.BackupSourceAsset.ComputerName, NewValue: ""},
		{FieldName: "BackupTargetType", FieldLabel: "备份对象类型", OldValue: record.BackupTargetType, NewValue: ""},
		{FieldName: "BackupTarget", FieldLabel: "备份对象", OldValue: record.BackupTarget, NewValue: ""},
		{FieldName: "BackupTool", FieldLabel: "备份工具", OldValue: record.BackupTool, NewValue: ""},
		{FieldName: "BackupMediumAssetID", FieldLabel: "备份介质", OldValue: record.BackupMediumAsset.ComputerName, NewValue: ""},
		{FieldName: "BackupFrequency", FieldLabel: "备份频率", OldValue: record.BackupFrequency, NewValue: ""},
		{FieldName: "RetentionPolicy", FieldLabel: "保留策略", OldValue: record.RetentionPolicy, NewValue: ""},
		{FieldName: "FullBackupStrategy", FieldLabel: "全量备份策略", OldValue: record.FullBackupStrategy, NewValue: ""},
		{FieldName: "DepartmentID", FieldLabel: "所属部门", OldValue: record.Department.Name, NewValue: ""},
		{FieldName: "FileName", FieldLabel: "申请表", OldValue: record.FileName, NewValue: ""},
	}

	services.LogOperation(username, displayName, "删除备份记录", "backup", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PreviewBackup 预览备份申请表
func PreviewBackup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.BackupRecord
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if record.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "申请表不存在"})
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

// DownloadBackup 下载备份申请表
func DownloadBackup(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.BackupRecord
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if record.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "申请表不存在"})
		return
	}

	if _, err := os.Stat(record.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+record.FileName+"\"")
	c.File(record.FilePath)
}

// CreateBackupRecovery 创建恢复还原记录
func CreateBackupRecovery(c *gin.Context) {
	backupID, _ := strconv.Atoi(c.Param("id"))
	var backupRecord models.BackupRecord
	if err := database.GetDB().First(&backupRecord, backupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "备份记录不存在"})
		return
	}

	recoveryType := c.PostForm("recovery_type")
	recoveryResult := c.PostForm("recovery_result")
	recoveryDate := c.PostForm("recovery_date")

	if recoveryType == "" || recoveryResult == "" || recoveryDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "恢复类型、恢复结果和恢复日期不能为空"})
		return
	}

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传恢复记录文件"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 PDF 格式文件"})
		return
	}

	// 从恢复日期提取年份
	year := time.Now().Format("2006")
	if len(recoveryDate) >= 4 {
		year = recoveryDate[:4]
	}

	yearDir := filepath.Join(config.Cfg.Upload.BackupRecoveryPath, year)
	if err := os.MkdirAll(yearDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "目录创建失败"})
		return
	}

	filename := formatFileName(file.Filename, ext)
	filePath := filepath.Join(yearDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	recovery := models.BackupRecovery{
		BackupRecordID: uint(backupID),
		RecoveryType:   recoveryType,
		RecoveryResult: recoveryResult,
		RecoveryDate:   recoveryDate,
		FileName:       file.Filename,
		FilePath:       filePath,
		FileSize:       file.Size,
		FileType:       file.Header.Get("Content-Type"),
	}

	if err := database.GetDB().Create(&recovery).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存恢复记录失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "RecoveryType", FieldLabel: "恢复类型", NewValue: recoveryType},
		{FieldName: "RecoveryResult", FieldLabel: "恢复结果", NewValue: recoveryResult},
		{FieldName: "RecoveryDate", FieldLabel: "恢复日期", NewValue: recoveryDate},
		{FieldName: "FileName", FieldLabel: "恢复与还原记录表", NewValue: file.Filename},
	}

	services.LogOperation(username, displayName, "创建恢复还原记录", "backup_recovery", recovery.ID, file.Filename, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": recovery})
}

// UpdateBackupRecovery 更新恢复还原记录
func UpdateBackupRecovery(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var recovery models.BackupRecovery
	if err := database.GetDB().First(&recovery, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "恢复记录不存在"})
		return
	}

	recoveryType := c.PostForm("recovery_type")
	recoveryResult := c.PostForm("recovery_result")
	recoveryDate := c.PostForm("recovery_date")

	oldRecoveryType := recovery.RecoveryType
	oldRecoveryResult := recovery.RecoveryResult
	oldRecoveryDate := recovery.RecoveryDate
	oldFilePath := recovery.FilePath

	// 年份变更时迁移文件
	if recoveryDate != "" && recoveryDate != recovery.RecoveryDate {
		newYear := recoveryDate[:4]
		oldYear := recovery.RecoveryDate[:4]
		if newYear != oldYear && recovery.FilePath != "" {
			newYearDir := filepath.Join(config.Cfg.Upload.BackupRecoveryPath, newYear)
			if err := os.MkdirAll(newYearDir, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "目录创建失败"})
				return
			}
			fileName := filepath.Base(recovery.FilePath)
			newFilePath := filepath.Join(newYearDir, fileName)
			if err := os.Rename(recovery.FilePath, newFilePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件迁移失败"})
				return
			}
			recovery.FilePath = newFilePath
		}
		recovery.RecoveryDate = recoveryDate
	}

	if recoveryType != "" {
		recovery.RecoveryType = recoveryType
	}
	if recoveryResult != "" {
		recovery.RecoveryResult = recoveryResult
	}

	if err := database.GetDB().Save(&recovery).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	var details []services.LogDetail

	if oldRecoveryType != recovery.RecoveryType {
		details = append(details, services.LogDetail{FieldName: "RecoveryType", FieldLabel: "恢复类型", OldValue: oldRecoveryType, NewValue: recovery.RecoveryType})
	}
	if oldRecoveryResult != recovery.RecoveryResult {
		details = append(details, services.LogDetail{FieldName: "RecoveryResult", FieldLabel: "恢复结果", OldValue: oldRecoveryResult, NewValue: recovery.RecoveryResult})
	}
	if oldRecoveryDate != recovery.RecoveryDate {
		details = append(details, services.LogDetail{FieldName: "RecoveryDate", FieldLabel: "恢复日期", OldValue: oldRecoveryDate, NewValue: recovery.RecoveryDate})
	}
	if oldFilePath != recovery.FilePath {
		details = append(details, services.LogDetail{FieldName: "FilePath", FieldLabel: "上传记录路径", OldValue: filepath.Base(oldFilePath), NewValue: filepath.Base(recovery.FilePath)})
	}

	services.LogOperation(username, displayName, "更新恢复还原记录", "backup_recovery", recovery.ID, recovery.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": recovery})
}

// DeleteBackupRecovery 删除恢复还原记录
func DeleteBackupRecovery(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var recovery models.BackupRecovery
	if err := database.GetDB().First(&recovery, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "恢复记录不存在"})
		return
	}

	// 删除物理文件
	if recovery.FilePath != "" {
		os.Remove(recovery.FilePath)
	}

	// 软删除记录
	if err := database.GetDB().Delete(&recovery).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "RecoveryType", FieldLabel: "恢复类型", OldValue: recovery.RecoveryType, NewValue: ""},
		{FieldName: "RecoveryResult", FieldLabel: "恢复结果", OldValue: recovery.RecoveryResult, NewValue: ""},
		{FieldName: "RecoveryDate", FieldLabel: "恢复日期", OldValue: recovery.RecoveryDate, NewValue: ""},
		{FieldName: "FileName", FieldLabel: "恢复与还原记录表", OldValue: recovery.FileName, NewValue: ""},
	}

	services.LogOperation(username, displayName, "删除恢复还原记录", "backup_recovery", recovery.ID, recovery.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PreviewBackupRecovery 预览恢复还原记录文件
func PreviewBackupRecovery(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var recovery models.BackupRecovery
	if err := database.GetDB().First(&recovery, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "恢复记录不存在"})
		return
	}

	if recovery.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	if _, err := os.Stat(recovery.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline; filename=\""+recovery.FileName+"\"")
	c.File(recovery.FilePath)
}

// DownloadBackupRecovery 下载恢复还原记录文件
func DownloadBackupRecovery(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var recovery models.BackupRecovery
	if err := database.GetDB().First(&recovery, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "恢复记录不存在"})
		return
	}

	if recovery.FilePath == "" {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	if _, err := os.Stat(recovery.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "文件不存在"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+recovery.FileName+"\"")
	c.File(recovery.FilePath)
}

// formatFileName 生成唯一文件名
func formatFileName(originalName, ext string) string {
	base := strings.TrimSuffix(filepath.Base(originalName), ext)
	return base + "_" + time.Now().Format("20060102150405") + ext
}
