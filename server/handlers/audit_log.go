package handlers

import (
	"net/http"
	"strconv"
	"time"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// ListLoginLogs 获取登录日志列表
func ListLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.GetDB().Model(&models.LoginLog{})

	// 用户名搜索
	username := c.Query("username")
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	// 日期范围
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var total int64
	query.Count(&total)

	var logs []models.LoginLog
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListOperationLogs 获取操作日志列表
func ListOperationLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	query := database.GetDB().Model(&models.OperationLog{})

	// 用户名搜索
	username := c.Query("username")
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}

	// 操作类型筛选（前缀匹配：如"创建"匹配"创建用户权限"、"创建区域"等）
	action := c.Query("action")
	if action != "" {
		query = query.Where("action LIKE ?", action+"%")
	}

	// 资源类型筛选
	resourceType := c.Query("resource_type")
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}

	// 日期范围
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at <= ?", t.Add(24*time.Hour))
		}
	}

	var total int64
	query.Count(&total)

	var logs []models.OperationLog
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":      200,
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetOperationLogDetails 获取操作日志明细
func GetOperationLogDetails(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var opLog models.OperationLog
	if err := database.GetDB().First(&opLog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "操作日志不存在"})
		return
	}

	var details []models.OperationLogDetail
	if err := database.GetDB().Where("operation_log_id = ?", id).Find(&details).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"log":     opLog,
			"details": details,
		},
	})
}

// Logout 用户登出
func Logout(c *gin.Context) {
	username, _ := c.Get("username")
	displayName, _ := c.Get("display_name")
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	usernameStr, _ := username.(string)
	displayNameStr, _ := displayName.(string)

	// 异步记录登出日志
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 忽略错误
			}
		}()
		logoutLog := models.LoginLog{
			Username:    usernameStr,
			DisplayName: displayNameStr,
			Action:      "logout",
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Detail:      "用户登出",
			CreatedAt:   time.Now(),
		}
		database.GetDB().Create(&logoutLog)
	}()

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "登出成功"})
}
