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

// ListUserChangeHistories 获取用户变更记录历史列表
func ListUserChangeHistories(c *gin.Context) {
	var records []models.UserChangeHistory

	query := database.GetDB().Order("year DESC, month DESC")

	// 按年份筛选
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", year)
		}
	}

	// 按关键词筛选描述
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("description LIKE ?", "%"+keyword+"%")
	}

	if err := query.Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": records})
}

// CreateUserChangeHistory 上传用户变更记录
func CreateUserChangeHistory(c *gin.Context) {
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

	// 获取上传文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件"})
		return
	}

	// 检查文件类型，仅允许PDF
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持PDF格式文件"})
		return
	}

	// 构建按年份的上传路径: permission_user_changes/{year}/
	yearDir := filepath.Join(config.Cfg.Upload.PermissionUserChangePath, strconv.Itoa(year))
	os.MkdirAll(yearDir, 0755)

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	record := models.UserChangeHistory{
		Year:        year,
		Month:       month,
		Description: description,
		FileName:    file.Filename,
		FilePath:    filePath,
		FileSize:    file.Size,
		FileType:    file.Header.Get("Content-Type"),
	}

	if err := database.GetDB().Create(&record).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存记录失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Year", FieldLabel: "年份", NewValue: strconv.Itoa(year)},
		{FieldName: "Month", FieldLabel: "月份", NewValue: strconv.Itoa(month)},
		{FieldName: "Description", FieldLabel: "描述", NewValue: description},
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: record.FileName},
		{FieldName: "FileSize", FieldLabel: "文件大小", NewValue: fmt.Sprintf("%d", record.FileSize)},
	}
	services.LogOperation(username, displayName, "上传用户变更记录", "user_change_history", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": record})
}

// DeleteUserChangeHistory 删除用户变更记录
func DeleteUserChangeHistory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.UserChangeHistory
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	// 删除文件
	os.Remove(record.FilePath)

	if err := database.GetDB().Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("user_change_history")
	details := services.DiffStructs(record, models.UserChangeHistory{}, fieldLabels)
	services.LogOperation(username, displayName, "删除用户变更记录", "user_change_history", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// UpdateUserChangeHistory 更新用户变更记录
func UpdateUserChangeHistory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.UserChangeHistory
	if err := database.GetDB().First(&record, id).Error; err != nil {
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

	oldRecord := record
	oldYear := record.Year
	oldFilePath := record.FilePath

	// 如果年份变化，需要移动文件到新目录
	if year != oldYear {
		oldYearDir := filepath.Join(config.Cfg.Upload.PermissionUserChangePath, strconv.Itoa(oldYear))
		newYearDir := filepath.Join(config.Cfg.Upload.PermissionUserChangePath, strconv.Itoa(year))
		os.MkdirAll(newYearDir, 0755)

		// 移动文件
		fileName := filepath.Base(record.FilePath)
		newFilePath := filepath.Join(newYearDir, fileName)
		if err := os.Rename(oldFilePath, newFilePath); err != nil {
			// 如果rename失败（跨分区），尝试复制后删除
			if copyErr := services.CopyFile(oldFilePath, newFilePath); copyErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件移动失败"})
				return
			}
			os.Remove(oldFilePath)
			record.FilePath = newFilePath
		} else {
			record.FilePath = newFilePath
		}

		// 清理旧的空目录
		os.Remove(oldYearDir)
	}

	record.Year = year
	record.Month = month
	record.Description = description

	if err := database.GetDB().Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("user_change_history")
	details := services.DiffStructs(oldRecord, record, fieldLabels)
	services.LogOperation(username, displayName, "更新用户变更记录", "user_change_history", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DownloadUserChangeHistory 下载用户变更记录文件
func DownloadUserChangeHistory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.UserChangeHistory
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

// PreviewUserChangeHistory 预览用户变更记录文件
func PreviewUserChangeHistory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.UserChangeHistory
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
