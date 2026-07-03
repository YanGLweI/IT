package handlers

import (
	"encoding/json"
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
	"gorm.io/gorm"
)

// ListQuarterlyChecks 获取季度检查历史列表
func ListQuarterlyChecks(c *gin.Context) {
	var records []models.QuarterlyCheckHistory

	query := database.GetDB().Model(&models.QuarterlyCheckHistory{})

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

	if err := query.Order("year DESC, quarter DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 批量预加载关联软件信息（避免N+1查询）
	if len(records) > 0 {
		recordIDs := make([]uint, len(records))
		for i, r := range records {
			recordIDs[i] = r.ID
		}
		var allLinks []models.QuarterlyCheckSoftware
		database.GetDB().Preload("ApprovedSoftware").
			Where("quarterly_check_history_id IN ?", recordIDs).
			Find(&allLinks)
		linkMap := make(map[uint][]models.ApprovedSoftware)
		for _, l := range allLinks {
			linkMap[l.QuarterlyCheckHistoryID] = append(linkMap[l.QuarterlyCheckHistoryID], l.ApprovedSoftware)
		}
		for i := range records {
			records[i].SoftwareList = linkMap[records[i].ID]
			if records[i].SoftwareList == nil {
				records[i].SoftwareList = []models.ApprovedSoftware{}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": records, "total": total, "page_size": pageSize})
}

// CreateQuarterlyCheck 上传季度检查记录
func CreateQuarterlyCheck(c *gin.Context) {
	yearStr := c.PostForm("year")
	quarterStr := c.PostForm("quarter")
	description := c.PostForm("description")

	if yearStr == "" || quarterStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份和季度不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	quarter, err := strconv.Atoi(quarterStr)
	if err != nil || quarter < 1 || quarter > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "季度格式不正确，应为1-4"})
		return
	}

	// 解析关联软件ID（逗号分隔）
	softwareIDsStr := c.PostForm("software_ids")
	var softwareIDs []uint
	if softwareIDsStr != "" {
		for _, part := range strings.Split(softwareIDsStr, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if id, err := strconv.ParseUint(part, 10, 64); err == nil {
				softwareIDs = append(softwareIDs, uint(id))
			}
		}
	}

	// 验证软件ID有效性
	if len(softwareIDs) > 0 {
		var validCount int64
		database.GetDB().Model(&models.ApprovedSoftware{}).
			Where("id IN ? AND need_update = ?", softwareIDs, true).
			Count(&validCount)
		if int(validCount) != len(softwareIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部分软件ID无效或不需要更新"})
			return
		}
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

	// 构建按年份的上传路径: third_party_quarterly_checks/{year}/
	yearDir := filepath.Join(config.Cfg.Upload.ThirdPartyQuarterlyCheckPath, strconv.Itoa(year))
	os.MkdirAll(yearDir, 0755)

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	record := models.QuarterlyCheckHistory{
		Year:        year,
		Quarter:     quarter,
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

	// 保存软件关联并更新软件状态（使用事务）
	if len(softwareIDs) > 0 {
		err = database.GetDB().Transaction(func(tx *gorm.DB) error {
			// 先查询选中软件的当前版本，用于保存原始版本和回滚
			var softwareList []models.ApprovedSoftware
			if err := tx.Where("id IN ?", softwareIDs).Find(&softwareList).Error; err != nil {
				return err
			}
			versionMap := make(map[uint]string)
			for _, sw := range softwareList {
				versionMap[sw.ID] = sw.Version
			}
			// 创建关联记录（含原始版本）
			for _, swID := range softwareIDs {
				qs := models.QuarterlyCheckSoftware{
					QuarterlyCheckHistoryID: record.ID,
					ApprovedSoftwareID:      swID,
					OriginalVersion:         versionMap[swID],
				}
				if err := tx.Create(&qs).Error; err != nil {
					return err
				}
			}
			// 批量更新软件状态：version = latest_version, need_update = false
			if err := tx.Model(&models.ApprovedSoftware{}).
				Where("id IN ?", softwareIDs).
				Updates(map[string]interface{}{
					"version":     gorm.Expr("latest_version"),
					"need_update": false,
				}).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "软件关联更新失败"})
			return
		}
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Year", FieldLabel: "年份", NewValue: strconv.Itoa(year)},
		{FieldName: "Quarter", FieldLabel: "季度", NewValue: fmt.Sprintf("Q%d", quarter)},
		{FieldName: "Description", FieldLabel: "描述", NewValue: description},
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: record.FileName},
		{FieldName: "FileSize", FieldLabel: "文件大小", NewValue: fmt.Sprintf("%d", record.FileSize)},
	}
	if len(softwareIDs) > 0 {
		softwareIDsBytes, _ := json.Marshal(softwareIDs)
		details = append(details, services.LogDetail{
			FieldName:  "SoftwareIDs",
			FieldLabel: "关联软件",
			OldValue:   "[]",
			NewValue:   string(softwareIDsBytes),
		})
	}
	services.LogOperation(username, displayName, "上传季度检查记录", "quarterly_check_history", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": record})
}

// DeleteQuarterlyCheck 删除季度检查记录
func DeleteQuarterlyCheck(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.QuarterlyCheckHistory
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	// 查询关联的软件记录
	var links []models.QuarterlyCheckSoftware
	database.GetDB().Where("quarterly_check_history_id = ?", record.ID).Find(&links)

	// 使用事务：回滚软件状态 + 删除关联记录 + 删除检查记录
	err := database.GetDB().Transaction(func(tx *gorm.DB) error {
		// 恢复关联软件的状态
		for _, link := range links {
			if err := tx.Model(&models.ApprovedSoftware{}).
				Where("id = ?", link.ApprovedSoftwareID).
				Updates(map[string]interface{}{
					"version":     link.OriginalVersion,
					"need_update": true,
				}).Error; err != nil {
				return err
			}
		}
		// 硬删除关联记录
		if err := tx.Unscoped().Where("quarterly_check_history_id = ?", record.ID).
			Delete(&models.QuarterlyCheckSoftware{}).Error; err != nil {
			return err
		}
		// 硬删除检查记录
		if err := tx.Unscoped().Delete(&record).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 删除文件
	os.Remove(record.FilePath)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Year", FieldLabel: "年份", OldValue: strconv.Itoa(record.Year), NewValue: ""},
		{FieldName: "Quarter", FieldLabel: "季度", OldValue: fmt.Sprintf("Q%d", record.Quarter), NewValue: ""},
		{FieldName: "Description", FieldLabel: "描述", OldValue: record.Description, NewValue: ""},
		{FieldName: "FileName", FieldLabel: "文件名", OldValue: record.FileName, NewValue: ""},
	}
	if len(links) > 0 {
		var rolledBack []string
		for _, link := range links {
			var sw models.ApprovedSoftware
			database.GetDB().Select("name").First(&sw, link.ApprovedSoftwareID)
			rolledBack = append(rolledBack, sw.Name)
		}
		details = append(details, services.LogDetail{
			FieldName:  "SoftwareRollback",
			FieldLabel: "回滚软件",
			OldValue:   strings.Join(rolledBack, "、"),
			NewValue:   "已恢复原状态",
		})
	}
	services.LogOperation(username, displayName, "删除季度检查记录", "quarterly_check_history", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// UpdateQuarterlyCheck 更新季度检查记录
func UpdateQuarterlyCheck(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.QuarterlyCheckHistory
	if err := database.GetDB().First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	yearStr := c.PostForm("year")
	quarterStr := c.PostForm("quarter")
	description := c.PostForm("description")

	if yearStr == "" || quarterStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份和季度不能为空"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "年份格式不正确"})
		return
	}

	quarter, err := strconv.Atoi(quarterStr)
	if err != nil || quarter < 1 || quarter > 4 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "季度格式不正确，应为1-4"})
		return
	}

	// 解析关联软件ID（逗号分隔）
	softwareIDsStr := c.PostForm("software_ids")
	var newSoftwareIDs []uint
	hasSoftwareIDs := false
	if softwareIDsStr != "" {
		hasSoftwareIDs = true
		for _, part := range strings.Split(softwareIDsStr, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			if id, err := strconv.ParseUint(part, 10, 64); err == nil {
				newSoftwareIDs = append(newSoftwareIDs, uint(id))
			}
		}
	}

	oldRecord := record
	oldYear := record.Year
	oldFilePath := record.FilePath

	record.Year = year
	record.Quarter = quarter
	record.Description = description

	var addedIDs, removedIDs []uint

	if hasSoftwareIDs {
		// 查询当前已关联的软件记录
		var oldLinks []models.QuarterlyCheckSoftware
		database.GetDB().Where("quarterly_check_history_id = ?", record.ID).Find(&oldLinks)
		oldIDMap := make(map[uint]models.QuarterlyCheckSoftware)
		for _, l := range oldLinks {
			oldIDMap[l.ApprovedSoftwareID] = l
		}

		// 计算差量
		newIDSet := make(map[uint]bool)
		for _, id := range newSoftwareIDs {
			newIDSet[id] = true
		}
		oldIDSet := make(map[uint]bool)
		for id := range oldIDMap {
			oldIDSet[id] = true
		}

		// 新增关联 = newIDs - oldIDs
		for _, id := range newSoftwareIDs {
			if !oldIDSet[id] {
				addedIDs = append(addedIDs, id)
			}
		}
		// 取消关联 = oldIDs - newIDs
		for id := range oldIDMap {
			if !newIDSet[id] {
				removedIDs = append(removedIDs, id)
			}
		}

		// 验证新增关联软件的有效性
		if len(addedIDs) > 0 {
			var validCount int64
			database.GetDB().Model(&models.ApprovedSoftware{}).
				Where("id IN ? AND need_update = ?", addedIDs, true).
				Count(&validCount)
			if int(validCount) != len(addedIDs) {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "部分新增软件ID无效或不需要更新"})
				return
			}
		}

		// 使用事务执行基础记录更新 + 软件差量更新
		err = database.GetDB().Transaction(func(tx *gorm.DB) error {
			// 更新基础记录
			if err := tx.Save(&record).Error; err != nil {
				return err
			}

			// 恢复取消关联的软件
			for _, rmID := range removedIDs {
				link := oldIDMap[rmID]
				if err := tx.Model(&models.ApprovedSoftware{}).
					Where("id = ?", rmID).
					Updates(map[string]interface{}{
						"version":     link.OriginalVersion,
						"need_update": true,
					}).Error; err != nil {
					return err
				}
			}
			// 硬删除取消关联的记录
			if len(removedIDs) > 0 {
				if err := tx.Unscoped().
					Where("quarterly_check_history_id = ? AND approved_software_id IN ?", record.ID, removedIDs).
					Delete(&models.QuarterlyCheckSoftware{}).Error; err != nil {
					return err
				}
			}
			// 新增关联：先查询软件当前版本存入OriginalVersion
			if len(addedIDs) > 0 {
				var addedSoftware []models.ApprovedSoftware
				if err := tx.Where("id IN ?", addedIDs).Find(&addedSoftware).Error; err != nil {
					return err
				}
				addedVersionMap := make(map[uint]string)
				for _, sw := range addedSoftware {
					addedVersionMap[sw.ID] = sw.Version
				}
				// 创建关联记录
				for _, swID := range addedIDs {
					qs := models.QuarterlyCheckSoftware{
						QuarterlyCheckHistoryID: record.ID,
						ApprovedSoftwareID:      swID,
						OriginalVersion:         addedVersionMap[swID],
					}
					if err := tx.Create(&qs).Error; err != nil {
						return err
					}
				}
				// 更新新增关联的软件状态
				if err := tx.Model(&models.ApprovedSoftware{}).
					Where("id IN ?", addedIDs).
					Updates(map[string]interface{}{
						"version":     gorm.Expr("latest_version"),
						"need_update": false,
					}).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
			return
		}
	} else {
		// 未提供软件关联参数，仅更新基础记录
		if err := database.GetDB().Save(&record).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
			return
		}
	}

	// 事务成功后，如果年份变化，移动文件到新目录
	if year != oldYear {
		oldYearDir := filepath.Join(config.Cfg.Upload.ThirdPartyQuarterlyCheckPath, strconv.Itoa(oldYear))
		newYearDir := filepath.Join(config.Cfg.Upload.ThirdPartyQuarterlyCheckPath, strconv.Itoa(year))
		os.MkdirAll(newYearDir, 0755)

		fileName := filepath.Base(record.FilePath)
		newFilePath := filepath.Join(newYearDir, fileName)
		if err := os.Rename(oldFilePath, newFilePath); err != nil {
			if copyErr := services.CopyFile(oldFilePath, newFilePath); copyErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件移动失败"})
				return
			}
			os.Remove(oldFilePath)
			record.FilePath = newFilePath
		} else {
			record.FilePath = newFilePath
		}

		os.Remove(oldYearDir)
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("quarterly_check_history")
	details := services.DiffStructs(oldRecord, record, fieldLabels)
	// 记录软件变更日志
	if len(addedIDs) > 0 {
		addedIDsBytes, _ := json.Marshal(addedIDs)
		details = append(details, services.LogDetail{
			FieldName:  "SoftwareAdded",
			FieldLabel: "新增关联软件",
			OldValue:   "[]",
			NewValue:   string(addedIDsBytes),
		})
	}
	if len(removedIDs) > 0 {
		removedIDsBytes, _ := json.Marshal(removedIDs)
		details = append(details, services.LogDetail{
			FieldName:  "SoftwareRemoved",
			FieldLabel: "取消关联软件",
			OldValue:   string(removedIDsBytes),
			NewValue:   "[]",
		})
	}
	services.LogOperation(username, displayName, "更新季度检查记录", "quarterly_check_history", record.ID, record.FileName, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DownloadQuarterlyCheck 下载季度检查文件
func DownloadQuarterlyCheck(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.QuarterlyCheckHistory
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

// PreviewQuarterlyCheck 预览季度检查文件
func PreviewQuarterlyCheck(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var record models.QuarterlyCheckHistory
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
