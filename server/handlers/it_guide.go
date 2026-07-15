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

// ============ 管理端接口 ============

// ListITGuides 获取指南列表
func ListITGuides(c *gin.Context) {
	var items []models.ITGuide

	query := database.GetDB().Model(&models.ITGuide{})

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if guideType := c.Query("guide_type"); guideType != "" {
		query = query.Where("guide_type = ?", guideType)
	}
	if published := c.Query("is_published"); published != "" {
		if published == "true" {
			query = query.Where("is_published = ?", true)
		} else if published == "false" {
			query = query.Where("is_published = ?", false)
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query.Count(&total)

	if err := query.Order("sort_order ASC, created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 获取所有已使用的分类列表
	var categories []string
	database.GetDB().Model(&models.ITGuide{}).
		Where("category != ''").
		Distinct().
		Pluck("category", &categories)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "total": total, "page_size": pageSize, "categories": categories})
}

// GetITGuide 获取单个指南详情（含步骤和媒体）
func GetITGuide(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在"})
		return
	}

	// 加载步骤
	var steps []models.ITGuideStep
	database.GetDB().Where("guide_id = ?", guide.ID).Order("sort_order ASC").Find(&steps)

	// 加载媒体
	var media []models.ITGuideMedia
	database.GetDB().Where("guide_id = ?", guide.ID).Order("sort_order ASC").Find(&media)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": guide, "steps": steps, "media": media})
}

// CreateITGuide 创建指南
func CreateITGuide(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	guideType := c.PostForm("guide_type")
	category := c.PostForm("category")

	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "指南标题不能为空"})
		return
	}
	if guideType != "step" && guideType != "video" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "指南类型必须为 step 或 video"})
		return
	}

	guide := models.ITGuide{
		Title:       title,
		Description: description,
		GuideType:   guideType,
		Category:    category,
	}

	if err := database.GetDB().Create(&guide).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Title", FieldLabel: "标题", NewValue: title},
		{FieldName: "GuideType", FieldLabel: "类型", NewValue: guideType},
		{FieldName: "Category", FieldLabel: "分类", NewValue: category},
	}
	services.LogOperation(username, displayName, "创建IT指南", "it_guide", guide.ID, guide.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": guide})
}

// UpdateITGuide 更新指南基本信息
func UpdateITGuide(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在"})
		return
	}

	oldGuide := guide

	title := c.PostForm("title")
	description := c.PostForm("description")
	category := c.PostForm("category")
	sortOrder := c.PostForm("sort_order")

	if title != "" {
		guide.Title = title
	}
	guide.Description = description
	guide.Category = category
	if sortOrder != "" {
		if v, err := strconv.Atoi(sortOrder); err == nil {
			guide.SortOrder = v
		}
	}

	if err := database.GetDB().Save(&guide).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := map[string]string{"Title": "标题", "Description": "描述", "Category": "分类", "SortOrder": "排序"}
	details := services.DiffStructs(oldGuide, guide, fieldLabels)
	services.LogOperation(username, displayName, "更新IT指南", "it_guide", guide.ID, guide.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": guide})
}

// DeleteITGuide 删除指南（级联删除步骤和媒体）
func DeleteITGuide(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在"})
		return
	}

	// 删除关联的媒体文件
	var mediaList []models.ITGuideMedia
	database.GetDB().Where("guide_id = ?", guide.ID).Find(&mediaList)
	for _, m := range mediaList {
		if m.FilePath != "" {
			os.Remove(m.FilePath)
		}
	}
	// 清理指南媒体目录
	guideMediaDir := filepath.Join(config.Cfg.Upload.ITGuideMediaPath, fmt.Sprintf("%d", guide.ID))
	os.RemoveAll(guideMediaDir)

	// 级联删除
	database.GetDB().Where("guide_id = ?", guide.ID).Delete(&models.ITGuideStep{})
	database.GetDB().Where("guide_id = ?", guide.ID).Delete(&models.ITGuideMedia{})
	database.GetDB().Delete(&guide)

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := map[string]string{"Title": "标题", "Description": "描述", "Category": "分类", "GuideType": "类型"}
	details := services.DiffStructs(guide, models.ITGuide{}, fieldLabels)
	services.LogOperation(username, displayName, "删除IT指南", "it_guide", guide.ID, guide.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// PublishITGuide 发布指南
func PublishITGuide(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在"})
		return
	}

	now := time.Now()
	if err := database.GetDB().Model(&guide).Updates(map[string]interface{}{
		"is_published": true,
		"published_at": &now,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "发布失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "IsPublished", FieldLabel: "发布状态", OldValue: "未发布", NewValue: "已发布"},
	}
	services.LogOperation(username, displayName, "发布IT指南", "it_guide", guide.ID, guide.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "发布成功"})
}

// UnpublishITGuide 取消发布
func UnpublishITGuide(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在"})
		return
	}

	if err := database.GetDB().Model(&guide).Updates(map[string]interface{}{
		"is_published": false,
		"published_at": nil,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "取消发布失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "IsPublished", FieldLabel: "发布状态", OldValue: "已发布", NewValue: "未发布"},
	}
	services.LogOperation(username, displayName, "取消发布IT指南", "it_guide", guide.ID, guide.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "取消发布成功"})
}

// ============ 步骤管理 ============

// ListITGuideSteps 获取指南的所有步骤
func ListITGuideSteps(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在"})
		return
	}

	var steps []models.ITGuideStep
	database.GetDB().Where("guide_id = ?", guide.ID).Order("sort_order ASC").Find(&steps)

	// 加载每个步骤的媒体
	type StepWithMedia struct {
		models.ITGuideStep
		Media []models.ITGuideMedia `json:"media"`
	}
	var result []StepWithMedia
	for _, s := range steps {
		var media []models.ITGuideMedia
		database.GetDB().Where("guide_id = ? AND step_id = ?", guide.ID, s.ID).Order("sort_order ASC").Find(&media)
		result = append(result, StepWithMedia{ITGuideStep: s, Media: media})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// CreateITGuideStep 添加步骤
func CreateITGuideStep(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在"})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	sortOrder, _ := strconv.Atoi(c.DefaultPostForm("sort_order", "0"))

	// 自动计算 step_number
	var maxStepNum int
	database.GetDB().Model(&models.ITGuideStep{}).
		Where("guide_id = ?", guide.ID).
		Select("COALESCE(MAX(step_number), 0)").
		Scan(&maxStepNum)

	step := models.ITGuideStep{
		GuideID:     uint(id),
		StepNumber:  maxStepNum + 1,
		Title:       title,
		Description: description,
		SortOrder:   sortOrder,
	}

	if err := database.GetDB().Create(&step).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建步骤失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Title", FieldLabel: "步骤标题", NewValue: title},
	}
	services.LogOperation(username, displayName, "添加指南步骤", "it_guide", guide.ID, guide.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": step})
}

// UpdateITGuideStep 更新步骤
func UpdateITGuideStep(c *gin.Context) {
	guideID, _ := strconv.Atoi(c.Param("id"))
	stepID, _ := strconv.Atoi(c.Param("stepId"))

	var step models.ITGuideStep
	if err := database.GetDB().Where("id = ? AND guide_id = ?", stepID, guideID).First(&step, stepID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "步骤不存在"})
		return
	}

	oldStep := step

	title := c.PostForm("title")
	description := c.PostForm("description")
	sortOrder := c.PostForm("sort_order")

	if title != "" {
		step.Title = title
	}
	step.Description = description
	if sortOrder != "" {
		if v, err := strconv.Atoi(sortOrder); err == nil {
			step.SortOrder = v
		}
	}

	if err := database.GetDB().Save(&step).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新步骤失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := map[string]string{"Title": "步骤标题", "Description": "步骤描述", "SortOrder": "排序"}
	details := services.DiffStructs(oldStep, step, fieldLabels)
	services.LogOperation(username, displayName, "更新指南步骤", "it_guide", step.GuideID, "", approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": step})
}

// DeleteITGuideStep 删除步骤
func DeleteITGuideStep(c *gin.Context) {
	stepID, _ := strconv.Atoi(c.Param("stepId"))
	var step models.ITGuideStep
	if err := database.GetDB().First(&step, stepID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "步骤不存在"})
		return
	}

	// 删除步骤关联的媒体文件
	var mediaList []models.ITGuideMedia
	database.GetDB().Where("guide_id = ? AND step_id = ?", step.GuideID, step.ID).Find(&mediaList)
	for _, m := range mediaList {
		if m.FilePath != "" {
			os.Remove(m.FilePath)
			dir := filepath.Dir(m.FilePath)
			entries, _ := os.ReadDir(dir)
			if len(entries) == 0 {
				os.Remove(dir)
			}
		}
	}
	database.GetDB().Where("guide_id = ? AND step_id = ?", step.GuideID, step.ID).Delete(&models.ITGuideMedia{})
	database.GetDB().Delete(&step)

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Title", FieldLabel: "步骤标题", OldValue: step.Title},
	}
	services.LogOperation(username, displayName, "删除指南步骤", "it_guide", step.GuideID, "", approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ReorderITGuideSteps 批量更新步骤排序
func ReorderITGuideSteps(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在"})
		return
	}

	var req struct {
		StepIDs []uint `json:"step_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	tx := database.GetDB().Begin()
	for i, stepID := range req.StepIDs {
		tx.Model(&models.ITGuideStep{}).Where("id = ? AND guide_id = ?", stepID, guide.ID).
			Updates(map[string]interface{}{"sort_order": i, "step_number": i + 1})
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "排序更新成功"})
}

// ============ 媒体上传 ============

// UploadITGuideMedia 上传媒体文件（支持文件上传和嵌入URL）
func UploadITGuideMedia(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在"})
		return
	}

	mediaType := c.PostForm("media_type")
	if mediaType != "image" && mediaType != "video" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "媒体类型必须为 image 或 video"})
		return
	}

	// 检查是否为嵌入视频（提供 embed_url 而非文件）
	embedURL := c.PostForm("embed_url")
	if embedURL != "" && mediaType == "video" {
		// 校验 URL 协议白名单
		if !strings.HasPrefix(embedURL, "http://") && !strings.HasPrefix(embedURL, "https://") && !strings.HasPrefix(embedURL, "//") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "嵌入URL必须以 http:// 或 https:// 开头"})
			return
		}
		stepID, _ := strconv.Atoi(c.PostForm("step_id"))
		sortOrder, _ := strconv.Atoi(c.DefaultPostForm("sort_order", "0"))

		media := models.ITGuideMedia{
			GuideID:   uint(id),
			StepID:    uint(stepID),
			MediaType: "video",
			FileName:  "embed_video",
			FilePath:  "",
			FileSize:  0,
			FileType:  "embed",
			EmbedURL:  embedURL,
			SortOrder: sortOrder,
		}

		if err := database.GetDB().Create(&media).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存嵌入视频记录失败"})
			return
		}

		username, displayName, approver := services.GetUserContext(c)
		details := []services.LogDetail{
			{FieldName: "EmbedURL", FieldLabel: "嵌入URL", NewValue: embedURL},
			{FieldName: "MediaType", FieldLabel: "媒体类型", NewValue: "video(embed)"},
		}
		services.LogOperation(username, displayName, "添加嵌入视频", "it_guide", guide.ID, guide.Title, approver, c.ClientIP(), details)

		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "嵌入视频添加成功", "data": media})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传文件或提供嵌入URL"})
		return
	}

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if mediaType == "image" {
		allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
		if !allowedExts[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片仅支持 JPG、PNG、GIF、WebP 格式"})
			return
		}
		if file.Size > 5*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过 5MB"})
			return
		}
	} else {
		allowedExts := map[string]bool{".mp4": true, ".webm": true}
		if !allowedExts[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "视频仅支持 MP4、WebM 格式"})
			return
		}
		if file.Size > 200*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "视频大小不能超过 200MB"})
			return
		}
	}

	// 按指南ID分目录存储
	yearDir := filepath.Join(config.Cfg.Upload.ITGuideMediaPath, fmt.Sprintf("%d", guide.ID))
	os.MkdirAll(yearDir, 0755)

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	stepID, _ := strconv.Atoi(c.PostForm("step_id"))
	sortOrder, _ := strconv.Atoi(c.DefaultPostForm("sort_order", "0"))

	media := models.ITGuideMedia{
		GuideID:   uint(id),
		StepID:    uint(stepID),
		MediaType: mediaType,
		FileName:  file.Filename,
		FilePath:  filePath,
		FileSize:  file.Size,
		FileType:  file.Header.Get("Content-Type"),
		SortOrder: sortOrder,
	}

	if err := database.GetDB().Create(&media).Error; err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存媒体记录失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "FileName", FieldLabel: "文件名", NewValue: file.Filename},
		{FieldName: "MediaType", FieldLabel: "媒体类型", NewValue: mediaType},
	}
	services.LogOperation(username, displayName, "上传指南媒体", "it_guide", guide.ID, guide.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": media})
}

// DeleteITGuideMedia 删除媒体文件
func DeleteITGuideMedia(c *gin.Context) {
	mediaID, _ := strconv.Atoi(c.Param("mediaId"))
	var media models.ITGuideMedia
	if err := database.GetDB().First(&media, mediaID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "媒体文件不存在"})
		return
	}

	if media.FilePath != "" {
		os.Remove(media.FilePath)
		dir := filepath.Dir(media.FilePath)
		entries, _ := os.ReadDir(dir)
		if len(entries) == 0 {
			os.Remove(dir)
		}
	}
	database.GetDB().Delete(&media)

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "FileName", FieldLabel: "文件名", OldValue: media.FileName},
	}
	services.LogOperation(username, displayName, "删除指南媒体", "it_guide", media.GuideID, "", approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ============ 公开接口 ============

// ListPublicITGuides 获取已发布指南列表（公开）
func ListPublicITGuides(c *gin.Context) {
	var items []models.ITGuide

	query := database.GetDB().Model(&models.ITGuide{}).Where("is_published = ?", true)

	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if guideType := c.Query("guide_type"); guideType != "" {
		query = query.Where("guide_type = ?", guideType)
	}

	if err := query.Order("sort_order ASC, published_at DESC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 获取分类列表
	var categories []string
	database.GetDB().Model(&models.ITGuide{}).
		Where("is_published = ? AND category != ''", true).
		Distinct().
		Pluck("category", &categories)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": items, "categories": categories})
}

// GetPublicITGuide 获取单个已发布指南详情（公开，含步骤和媒体）
func GetPublicITGuide(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var guide models.ITGuide
	if err := database.GetDB().Where("is_published = ?", true).First(&guide, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "指南不存在或未发布"})
		return
	}

	// 加载步骤（含媒体）
	type StepWithMedia struct {
		models.ITGuideStep
		Media []models.ITGuideMedia `json:"media"`
	}
	var steps []StepWithMedia
	var rawSteps []models.ITGuideStep
	database.GetDB().Where("guide_id = ?", guide.ID).Order("sort_order ASC").Find(&rawSteps)
	for _, s := range rawSteps {
		var media []models.ITGuideMedia
		database.GetDB().Where("guide_id = ? AND step_id = ?", guide.ID, s.ID).Order("sort_order ASC").Find(&media)
		steps = append(steps, StepWithMedia{ITGuideStep: s, Media: media})
	}

	// 加载指南本身的媒体（视频指南的视频）
	var guideMedia []models.ITGuideMedia
	database.GetDB().Where("guide_id = ? AND step_id = 0", guide.ID).Order("sort_order ASC").Find(&guideMedia)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": guide, "steps": steps, "media": guideMedia})
}
