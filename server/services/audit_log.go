package services

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserContext 从Gin上下文中安全地获取用户信息
func GetUserContext(c *gin.Context) (username, displayName, approver string) {
	if u, ok := c.Get("username"); ok {
		if s, ok := u.(string); ok {
			username = s
		}
	}
	if d, ok := c.Get("display_name"); ok {
		if s, ok := d.(string); ok {
			displayName = s
		}
	}
	if a, ok := c.Get("dual_control_verified_by"); ok {
		if s, ok := a.(string); ok {
			approver = s
		}
	}
	return username, displayName, approver
}

// LogDetail 日志明细结构
type LogDetail struct {
	FieldName  string `json:"field_name"`
	FieldLabel string `json:"field_label"`
	OldValue   string `json:"old_value"`
	NewValue   string `json:"new_value"`
}

// LogLogin 记录登录日志（异步）
func LogLogin(username, displayName, action, ipAddress, userAgent, detail string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("记录登录日志失败: %v", r)
			}
		}()
		loginLog := models.LoginLog{
			Username:    username,
			DisplayName: displayName,
			Action:      action,
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Detail:      detail,
			CreatedAt:   time.Now(),
		}
		database.GetDB().Create(&loginLog)
	}()
}

// LogOperation 记录操作日志（同步事务）
func LogOperation(username, displayName, action, resourceType string, resourceID uint, resourceName, approver, ipAddress string, details []LogDetail) error {
	db := database.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		opLog := models.OperationLog{
			Username:     username,
			DisplayName:  displayName,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			ResourceName: resourceName,
			Approver:     approver,
			IPAddress:    ipAddress,
			CreatedAt:    time.Now(),
		}
		if err := tx.Create(&opLog).Error; err != nil {
			return err
		}

		// 批量创建明细
		for _, d := range details {
			detail := models.OperationLogDetail{
				OperationLogID: opLog.ID,
				FieldName:      d.FieldName,
				FieldLabel:     d.FieldLabel,
				OldValue:       d.OldValue,
				NewValue:       d.NewValue,
				CreatedAt:      time.Now(),
			}
			if err := tx.Create(&detail).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DiffStructs 比较两个struct生成变更明细
func DiffStructs(old interface{}, new interface{}, fieldLabels map[string]string) []LogDetail {
	var details []LogDetail
	oldVal := reflect.ValueOf(old)
	newVal := reflect.ValueOf(new)

	if oldVal.Kind() == reflect.Ptr {
		oldVal = oldVal.Elem()
	}
	if newVal.Kind() == reflect.Ptr {
		newVal = newVal.Elem()
	}

	skipFields := map[string]bool{
		"ID":        true,
		"CreatedAt": true,
		"UpdatedAt": true,
		"DeletedAt": true,
	}

	for i := 0; i < oldVal.NumField(); i++ {
		field := oldVal.Type().Field(i)
		fieldName := field.Name

		// 跳过特定字段
		if skipFields[fieldName] {
			continue
		}

		oldFieldValue := oldVal.Field(i)
		newFieldValue := newVal.Field(i)

		// 转换为字符串进行比较
		oldStr := fmt.Sprintf("%v", oldFieldValue.Interface())
		newStr := fmt.Sprintf("%v", newFieldValue.Interface())

		if oldStr != newStr {
			label := fieldName
			if lbl, ok := fieldLabels[fieldName]; ok {
				label = lbl
			}
			details = append(details, LogDetail{
				FieldName:  fieldName,
				FieldLabel: label,
				OldValue:   oldStr,
				NewValue:   newStr,
			})
		}
	}

	return details
}

// GetFieldLabels 获取各模型的字段中文标签映射
func GetFieldLabels(resourceType string) map[string]string {
	switch resourceType {
	case "asset":
		return map[string]string{
			"ComputerName": "计算机名",
			"RegionID":     "区域ID",
			"IPAddress":    "IP地址",
			"OSType":       "操作系统",
			"Purpose":      "用途",
			"AssetLevel":   "资产等级",
			"Status":       "状态",
			"Remark":       "备注",
		}
	case "region":
		return map[string]string{
			"Name":        "名称",
			"Description": "描述",
		}
	case "policy":
		return map[string]string{
			"Title":       "标题",
			"Description": "描述",
			"FileName":    "文件名",
			"FilePath":    "文件路径",
			"FileSize":    "文件大小",
			"FileType":    "文件类型",
		}
	case "topology":
		return map[string]string{
			"Name":        "名称",
			"Description": "描述",
			"FileName":    "文件名",
			"FilePath":    "文件路径",
			"FileSize":    "文件大小",
		}
	case "os_type":
		return map[string]string{
			"Name": "名称",
		}
	case "department":
		return map[string]string{
			"Name":      "名称",
			"SortOrder": "排序",
		}
	case "department_position":
		return map[string]string{
			"DepartmentID": "部门ID",
			"PositionName": "岗位名称",
		}
	case "permission_rule":
		return map[string]string{
			"PositionName": "岗位名称",
			"SortOrder":    "排序",
			"RulesJSON":    "权限规则",
		}
	case "user_permission":
		return map[string]string{
			"Name":            "姓名",
			"DepartmentID":    "部门ID",
			"PositionName":    "岗位名称",
			"SystemRolesJSON": "系统角色",
		}
	case "sftp_server":
		return map[string]string{
			"Name":      "名称",
			"SortOrder": "排序",
		}
	case "sftp_account":
		return map[string]string{
			"ServerID":        "服务器ID",
			"AccountName":     "账号名称",
			"CreatedTime":     "创建时间",
			"Validity":        "有效期",
			"PermissionsJSON": "权限",
			"ContactPerson":   "联系人",
			"Department":      "部门",
			"WhitelistJSON":   "白名单",
		}
	case "approved_software":
		return map[string]string{
			"Name":          "名称",
			"Version":       "版本",
			"LatestVersion": "最新版本",
			"NeedUpdate":    "需要更新",
			"UpdateReason":  "更新原因",
			"Vendor":        "厂商",
			"VendorWebsite": "厂商网站",
			"LicenseType":   "许可证类型",
			"Purpose":       "用途",
		}
	case "asset_software":
		return map[string]string{
			"AssetID":            "资产ID",
			"ApprovedSoftwareID": "核准软件ID",
		}
	case "monthly_check_history":
		return map[string]string{
			"Year":        "年份",
			"Month":       "月份",
			"Description": "描述",
			"FileName":    "文件名",
			"FilePath":    "文件路径",
			"FileSize":    "文件大小",
			"FileType":    "文件类型",
		}
	case "quarterly_check_history":
		return map[string]string{
			"Year":        "年份",
			"Quarter":     "季度",
			"Description": "描述",
			"FileName":    "文件名",
			"FilePath":    "文件路径",
			"FileSize":    "文件大小",
			"FileType":    "文件类型",
		}
	default:
		return make(map[string]string)
	}
}
