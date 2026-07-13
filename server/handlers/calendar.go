package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateCalendarRequest 创建日程请求
type CreateCalendarRequest struct {
	Title          string                        `json:"title" binding:"required"`
	Description    string                        `json:"description"`
	StartTime      time.Time                     `json:"start_time" binding:"required"`
	EndTime        time.Time                     `json:"end_time" binding:"required"`
	IsAllDay       bool                          `json:"is_all_day"`
	RepeatRuleJSON string                        `json:"repeat_rule_json"`
	Participants   []CreateCalendarParticipantReq `json:"participants"`
}

// CreateCalendarParticipantReq 创建参与者请求
type CreateCalendarParticipantReq struct {
	UserDN      string `json:"user_dn" binding:"required"`
	DisplayName string `json:"display_name"`
}

// CheckConflictRequest 冲突检测请求
type CheckConflictRequest struct {
	StartTime    time.Time                      `json:"start_time" binding:"required"`
	EndTime      time.Time                      `json:"end_time" binding:"required"`
	RepeatRuleJSON string                       `json:"repeat_rule_json"`
	Participants []CreateCalendarParticipantReq  `json:"participants"`
	ExcludeID    *uint                          `json:"exclude_id"` // 编辑时排除自身
}

// ConflictItem 冲突项
type ConflictItem struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	UserName  string `json:"user_name"`
}

// ListCalendars 查询日程列表（仅返回当前用户创建或参与的日程）
func ListCalendars(c *gin.Context) {
	username, _, _ := services.GetUserContext(c)
	db := database.GetDB()

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	keyword := c.Query("keyword")

	// 基础查询：当前用户创建或参与的日程
	query := db.Model(&models.Calendar{}).
		Joins("LEFT JOIN calendar_participants ON calendar_participants.calendar_id = calendars.id").
		Where("calendars.deleted_at IS NULL AND (calendars.created_by = ? OR calendar_participants.user_dn = ?)", username, username).
		Preload("Participants").
		Distinct()

	// 日期范围筛选
	if startDate != "" && endDate != "" {
		rangeStart, err1 := time.Parse("2006-01-02", startDate)
		rangeEnd, err2 := time.Parse("2006-01-02", endDate)
		if err1 == nil && err2 == nil {
			// 查询与范围有交集的日程（含重复日程展开）
			// 先查基础日程
			query = query.Where(
				"(calendars.start_time <= ? AND calendars.end_time >= ?) OR (calendars.repeat_rule_json IS NOT NULL AND calendars.repeat_rule_json != '' AND calendars.repeat_rule_json != 'null')",
				rangeEnd, rangeStart,
			)
		}
	}

	// 关键词搜索
	if keyword != "" {
		query = query.Where("calendars.title LIKE ?", "%"+keyword+"%")
	}

	var calendars []models.Calendar
	if err := query.Order("calendars.start_time ASC").Find(&calendars).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询日程失败: " + err.Error()})
		return
	}

	// 如果有日期范围，对重复日程进行展开过滤
	if startDate != "" && endDate != "" {
		rangeStart, _ := time.Parse("2006-01-02", startDate)
		rangeEndTmp, _ := time.Parse("2006-01-02", endDate)
		rangeEnd := rangeEndTmp.AddDate(0, 0, 1) // endDate的23:59:59

		var filtered []models.Calendar
		for _, cal := range calendars {
			rule, _ := services.ParseRepeatRule(cal.RepeatRuleJSON)
			if rule != nil {
				instances := services.ExpandRecurringEvents(rule, cal.StartTime, rangeStart, rangeEnd)
				if len(instances) > 0 {
					filtered = append(filtered, cal)
				}
			} else {
				// 非重复日程，检查时间重叠
				if services.CheckTimeOverlap(cal.StartTime, cal.EndTime, rangeStart, rangeEnd) {
					filtered = append(filtered, cal)
				}
			}
		}
		calendars = filtered
	}

	if calendars == nil {
		calendars = []models.Calendar{}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": calendars,
	})
}

// GetCalendar 获取单个日程详情
func GetCalendar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的日程ID"})
		return
	}

	username, _, _ := services.GetUserContext(c)
	db := database.GetDB()

	var calendar models.Calendar
	if err := db.Preload("Participants").First(&calendar, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "日程不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 检查可见性：只有创建者或参与者可查看
	if calendar.CreatedBy != username {
		isParticipant := false
		for _, p := range calendar.Participants {
			if p.UserDN == username {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "无权查看此日程"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": calendar,
	})
}

// CreateCalendar 创建日程
func CreateCalendar(c *gin.Context) {
	var req CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	username, displayName, _ := services.GetUserContext(c)
	db := database.GetDB()

	// 校验开始时间不能是过去时间
	if req.StartTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不能创建过去时间的日程"})
		return
	}

	// 处理空repeat_rule_json（MySQL JSON列不接受空字符串）
	repeatRule := req.RepeatRuleJSON
	if repeatRule == "" {
		repeatRule = "null"
	}

	calendar := models.Calendar{
		Title:          req.Title,
		Description:    req.Description,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		IsAllDay:       req.IsAllDay,
		RepeatRuleJSON: repeatRule,
		CreatedBy:      username,
	}

	// 事务创建日程和参与者
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&calendar).Error; err != nil {
			return fmt.Errorf("创建日程失败: %v", err)
		}

		// 批量插入参与者（使用短用户名确保可见性查询匹配）
		for _, p := range req.Participants {
			participant := models.CalendarParticipant{
				CalendarID:  calendar.ID,
				UserDN:      extractShortUsername(p.UserDN),
				DisplayName: p.DisplayName,
			}
			if err := tx.Create(&participant).Error; err != nil {
				return fmt.Errorf("创建参与者失败: %v", err)
			}
		}

		// 生成通知记录
		generateNotificationRecords(tx, &calendar, req.Participants)

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 记录审计日志（同步写入，确保即时可查）
	details := []services.LogDetail{
		{FieldName: "title", FieldLabel: "标题", OldValue: "", NewValue: req.Title},
		{FieldName: "start_time", FieldLabel: "开始时间", OldValue: "", NewValue: req.StartTime.Format("2006-01-02 15:04")},
		{FieldName: "end_time", FieldLabel: "结束时间", OldValue: "", NewValue: req.EndTime.Format("2006-01-02 15:04")},
	}
	if req.Description != "" {
		details = append(details, services.LogDetail{FieldName: "description", FieldLabel: "描述", OldValue: "", NewValue: req.Description})
	}
	participantNames := make([]string, len(req.Participants))
	for i, p := range req.Participants {
		participantNames[i] = p.DisplayName
	}
	if len(participantNames) > 0 {
		details = append(details, services.LogDetail{FieldName: "participants", FieldLabel: "参与者", OldValue: "", NewValue: strings.Join(participantNames, ", ")})
	}
	services.LogOperation(username, displayName, "create", "calendar", calendar.ID, req.Title, "", c.ClientIP(), details)

	// 重新查询带参与者的日程
	db.Preload("Participants").First(&calendar, calendar.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    calendar,
	})
}

// UpdateCalendar 更新日程
func UpdateCalendar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的日程ID"})
		return
	}

	username, displayName, _ := services.GetUserContext(c)
	db := database.GetDB()

	var calendar models.Calendar
	if err := db.Preload("Participants").First(&calendar, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "日程不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 只有创建者可编辑
	if calendar.CreatedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "只有创建者可编辑此日程"})
		return
	}

	var req CreateCalendarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 记录旧值用于审计
	oldCalendar := calendar
	oldParticipants := make([]string, len(calendar.Participants))
	for i, p := range calendar.Participants {
		oldParticipants[i] = p.DisplayName
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新主表
		calendar.Title = req.Title
		calendar.Description = req.Description
		calendar.StartTime = req.StartTime
		calendar.EndTime = req.EndTime
		calendar.IsAllDay = req.IsAllDay
		calendar.RepeatRuleJSON = req.RepeatRuleJSON
		if calendar.RepeatRuleJSON == "" {
			calendar.RepeatRuleJSON = "null"
		}

		if err := tx.Save(&calendar).Error; err != nil {
			return fmt.Errorf("更新日程失败: %v", err)
		}

		// 删除旧参与者
		if err := tx.Where("calendar_id = ?", calendar.ID).Delete(&models.CalendarParticipant{}).Error; err != nil {
			return fmt.Errorf("删除旧参与者失败: %v", err)
		}

		// 插入新参与者（使用短用户名）
		for _, p := range req.Participants {
			participant := models.CalendarParticipant{
				CalendarID:  calendar.ID,
				UserDN:      extractShortUsername(p.UserDN),
				DisplayName: p.DisplayName,
			}
			if err := tx.Create(&participant).Error; err != nil {
				return fmt.Errorf("创建参与者失败: %v", err)
			}
		}

		// 重新生成通知记录
		tx.Where("calendar_id = ?", calendar.ID).Delete(&models.CalendarNotification{})
		generateNotificationRecords(tx, &calendar, req.Participants)

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 记录审计日志（同步写入，确保即时可查）
	fieldLabels := services.GetFieldLabels("calendar")
	details := services.DiffStructs(oldCalendar, calendar, fieldLabels)
	// 参与者变更
	newParticipants := make([]string, len(req.Participants))
	for i, p := range req.Participants {
		newParticipants[i] = p.DisplayName
	}
	oldStr := strings.Join(oldParticipants, ", ")
	newStr := strings.Join(newParticipants, ", ")
	if oldStr != newStr {
		details = append(details, services.LogDetail{
			FieldName: "participants", FieldLabel: "参与者",
			OldValue: oldStr, NewValue: newStr,
		})
	}
	if len(details) > 0 {
		services.LogOperation(username, displayName, "update", "calendar", calendar.ID, req.Title, "", c.ClientIP(), details)
	}

	db.Preload("Participants").First(&calendar, calendar.ID)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    calendar,
	})
}

// DeleteCalendar 删除日程
func DeleteCalendar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的日程ID"})
		return
	}

	username, displayName, _ := services.GetUserContext(c)
	db := database.GetDB()

	var calendar models.Calendar
	if err := db.Preload("Participants").First(&calendar, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "日程不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 只有创建者可删除
	if calendar.CreatedBy != username {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "只有创建者可删除此日程"})
		return
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除参与者
		if err := tx.Where("calendar_id = ?", calendar.ID).Delete(&models.CalendarParticipant{}).Error; err != nil {
			return err
		}
		// 删除通知记录
		if err := tx.Where("calendar_id = ?", calendar.ID).Delete(&models.CalendarNotification{}).Error; err != nil {
			return err
		}
		// 软删除日程
		if err := tx.Delete(&calendar).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败: " + err.Error()})
		return
	}

	// 记录审计日志（同步写入，确保即时可查）
	participantNames := make([]string, len(calendar.Participants))
	for i, p := range calendar.Participants {
		participantNames[i] = p.DisplayName
	}
	delDetails := []services.LogDetail{
		{FieldName: "title", FieldLabel: "标题", OldValue: calendar.Title, NewValue: ""},
		{FieldName: "start_time", FieldLabel: "开始时间", OldValue: calendar.StartTime.Format("2006-01-02 15:04"), NewValue: ""},
		{FieldName: "end_time", FieldLabel: "结束时间", OldValue: calendar.EndTime.Format("2006-01-02 15:04"), NewValue: ""},
	}
	if len(participantNames) > 0 {
		delDetails = append(delDetails, services.LogDetail{
			FieldName: "participants", FieldLabel: "参与者",
			OldValue: strings.Join(participantNames, ", "), NewValue: "",
		})
	}
	services.LogOperation(username, displayName, "delete", "calendar", calendar.ID, calendar.Title, "", c.ClientIP(), delDetails)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
	})
}

// GetTodayNotifications 获取当前用户所有未读通知（含即时通知和当天通知）
func GetTodayNotifications(c *gin.Context) {
	username, _, _ := services.GetUserContext(c)
	db := database.GetDB()

	var notifications []models.CalendarNotification
	db.Where("user_dn = ? AND read_at IS NULL",
		username).
		Order("notify_time ASC").
		Find(&notifications)

	// 填充日程信息
	for i := range notifications {
		var calendar models.Calendar
		if err := db.First(&calendar, notifications[i].CalendarID).Error; err == nil {
			notifications[i].CalendarTitle = calendar.Title
			notifications[i].StartTime = &calendar.StartTime
			notifications[i].EndTime = &calendar.EndTime
			notifications[i].IsAllDay = calendar.IsAllDay
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": notifications,
	})
}

// GetUnreadCount 获取未读通知数量
func GetUnreadCount(c *gin.Context) {
	username, _, _ := services.GetUserContext(c)
	db := database.GetDB()

	var count int64
	db.Model(&models.CalendarNotification{}).
		Where("user_dn = ? AND read_at IS NULL", username).
		Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"count": count},
	})
}

// MarkNotificationRead 标记通知为已读
func MarkNotificationRead(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的通知ID"})
		return
	}

	username, _, _ := services.GetUserContext(c)
	db := database.GetDB()

	now := time.Now()
	result := db.Model(&models.CalendarNotification{}).
		Where("id = ? AND user_dn = ?", id, username).
		Update("read_at", now)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "标记失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "已标记为已读",
	})
}

// GetPendingNotifications 获取未来24小时内待通知的日程（用于前端定时轮询）
func GetPendingNotifications(c *gin.Context) {
	username, _, _ := services.GetUserContext(c)
	db := database.GetDB()

	now := time.Now()
	past24h := now.Add(-24 * time.Hour)

	var notifications []models.CalendarNotification
	db.Where("user_dn = ? AND notify_time <= ? AND notify_time >= ? AND popup_shown = false AND read_at IS NULL",
		username, now, past24h).
		Order("notify_time ASC").
		Find(&notifications)

	// 填充日程信息
	for i := range notifications {
		var calendar models.Calendar
		if err := db.First(&calendar, notifications[i].CalendarID).Error; err == nil {
			notifications[i].CalendarTitle = calendar.Title
			notifications[i].StartTime = &calendar.StartTime
			notifications[i].EndTime = &calendar.EndTime
			notifications[i].IsAllDay = calendar.IsAllDay
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": notifications,
	})
}

// MarkNotificationPopupShown 标记通知已弹出
func MarkNotificationPopupShown(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的通知ID"})
		return
	}

	username, _, _ := services.GetUserContext(c)
	db := database.GetDB()

	db.Model(&models.CalendarNotification{}).
		Where("id = ? AND user_dn = ?", id, username).
		Update("popup_shown", true)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "ok",
	})
}

// CheckConflict 检测日程时间冲突
func CheckConflict(c *gin.Context) {
	var req CheckConflictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	db := database.GetDB()
	var conflicts []ConflictItem

	// 对每个参与者检查冲突
	for _, p := range req.Participants {
		// 查询该参与者的所有日程
		var calendars []models.Calendar
		query := db.Joins("JOIN calendar_participants ON calendar_participants.calendar_id = calendars.id").
			Where("calendar_participants.user_dn = ? AND calendars.deleted_at IS NULL", p.UserDN).
			Preload("Participants")

		if req.ExcludeID != nil {
			query = query.Where("calendars.id != ?", *req.ExcludeID)
		}

		query.Find(&calendars)

		// 计算新日程的实例时间
		newRule, _ := services.ParseRepeatRule(req.RepeatRuleJSON)
		var newInstances []time.Time
		if newRule != nil {
			// 展开未来30天
			rangeEnd := req.StartTime.AddDate(0, 1, 0)
			newInstances = services.ExpandRecurringEvents(newRule, req.StartTime, req.StartTime, rangeEnd)
		} else {
			newInstances = []time.Time{req.StartTime}
		}

		duration := req.EndTime.Sub(req.StartTime)

		for _, cal := range calendars {
			existRule, _ := services.ParseRepeatRule(cal.RepeatRuleJSON)
			var existInstances []time.Time
			if existRule != nil {
				rangeEnd := req.StartTime.AddDate(0, 1, 0)
				existInstances = services.ExpandRecurringEvents(existRule, cal.StartTime, req.StartTime.AddDate(0, -1, 0), rangeEnd)
			} else {
				existInstances = []time.Time{cal.StartTime}
			}

			existDuration := cal.EndTime.Sub(cal.StartTime)

			for _, newInst := range newInstances {
				newEnd := newInst.Add(duration)
				for _, existInst := range existInstances {
					existEnd := existInst.Add(existDuration)
					if services.CheckTimeOverlap(newInst, newEnd, existInst, existEnd) {
						conflicts = append(conflicts, ConflictItem{
							ID:        cal.ID,
							Title:     cal.Title,
							StartTime: existInst,
							EndTime:   existEnd,
							UserName:  p.DisplayName,
						})
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": conflicts,
	})
}

// extractShortUsername 从LDAP DN中提取短用户名（CN值）
// 例如: "CN=zbj,OU=IT,DC=example,DC=com" -> "zbj"
func extractShortUsername(dn string) string {
	dn = strings.TrimSpace(dn)
	upper := strings.ToUpper(dn)
	if idx := strings.Index(upper, "CN="); idx >= 0 {
		rest := dn[idx+3:]
		if commaIdx := strings.Index(rest, ","); commaIdx >= 0 {
			return rest[:commaIdx]
		}
		return rest
	}
	return dn
}

// generateNotificationRecords 为日程生成通知记录
func generateNotificationRecords(tx *gorm.DB, calendar *models.Calendar, participants []CreateCalendarParticipantReq) {
	now := time.Now()

	// 收集所有需要通知的用户（去重）
	userSet := make(map[string]string) // shortUsername -> displayName
	for _, p := range participants {
		shortName := extractShortUsername(p.UserDN)
		userSet[shortName] = p.DisplayName
	}
	// 创建者也加入通知列表
	if _, exists := userSet[calendar.CreatedBy]; !exists {
		userSet[calendar.CreatedBy] = ""
	}

	for shortName := range userSet {
		// 仅在日程时间到达时通知（创建时不弹框，到达日程时间才弹框提醒）
		dayNotif := models.CalendarNotification{
			CalendarID: calendar.ID,
			UserDN:     shortName,
			NotifyType: "login",
			NotifyTime: calendar.StartTime,
			SentAt:     now,
		}
		tx.Create(&dayNotif)
	}
}
