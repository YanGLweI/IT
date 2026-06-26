package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ListSftpAccounts 获取SFTP账号列表
func ListSftpAccounts(c *gin.Context) {
	serverIDStr := c.Query("server_id")
	if serverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请指定服务器"})
		return
	}

	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的服务器ID"})
		return
	}

	var accounts []models.SftpAccount
	if err := database.GetDB().Where("server_id = ?", serverID).Order("created_at ASC").Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": accounts})
}

// CreateSftpAccount 创建SFTP账号
func CreateSftpAccount(c *gin.Context) {
	var account models.SftpAccount
	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 检查服务器是否存在
	var server models.SftpServer
	if err := database.GetDB().First(&server, account.ServerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务器不存在"})
		return
	}

	// 检查同服务器下账号名是否重复
	var count int64
	database.GetDB().Model(&models.SftpAccount{}).Where("server_id = ? AND account_name = ?", account.ServerID, account.AccountName).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该服务器下已存在同名账号"})
		return
	}

	if err := database.GetDB().Create(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": account})
}

// UpdateSftpAccount 更新SFTP账号
func UpdateSftpAccount(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var account models.SftpAccount
	if err := database.GetDB().First(&account, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "账号不存在"})
		return
	}

	var input struct {
		AccountName     string `json:"account_name" binding:"required"`
		CreatedTime     string `json:"created_time"`
		Validity        string `json:"validity"`
		PermissionsJSON string `json:"permissions_json"`
		ContactPerson   string `json:"contact_person"`
		DepartmentID    uint   `json:"department_id"`
		WhitelistJSON   string `json:"whitelist_json"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 检查同服务器下账号名是否重复（排除自己）
	var count int64
	database.GetDB().Model(&models.SftpAccount{}).Where("server_id = ? AND account_name = ? AND id != ?", account.ServerID, input.AccountName, id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该服务器下已存在同名账号"})
		return
	}

	account.AccountName = input.AccountName
	account.CreatedTime = input.CreatedTime
	account.Validity = input.Validity
	account.PermissionsJSON = input.PermissionsJSON
	account.ContactPerson = input.ContactPerson
	account.DepartmentID = input.DepartmentID
	account.WhitelistJSON = input.WhitelistJSON

	if err := database.GetDB().Save(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": account})
}

// DeleteSftpAccount 删除SFTP账号
func DeleteSftpAccount(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var account models.SftpAccount
	if err := database.GetDB().First(&account, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "账号不存在"})
		return
	}

	if err := database.GetDB().Delete(&account).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ExportSftpConfirmation 导出SFTP账号月度确认表Excel
func ExportSftpConfirmation(c *gin.Context) {
	serverIDStr := c.Query("server_id")
	if serverIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请指定服务器"})
		return
	}

	serverID, err := strconv.ParseUint(serverIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的服务器ID"})
		return
	}

	// 查询服务器
	var server models.SftpServer
	if err := database.GetDB().First(&server, serverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务器不存在"})
		return
	}

	// 查询该服务器下所有账号
	var accounts []models.SftpAccount
	if err := database.GetDB().Where("server_id = ?", serverID).Order("created_at ASC").Find(&accounts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询账号失败"})
		return
	}

	// 查询部门列表（用于显示部门名称）
	var departments []models.Department
	database.GetDB().Find(&departments)
	deptMap := make(map[uint]string)
	for _, dept := range departments {
		deptMap[dept.ID] = dept.Name
	}

	// 当前年月
	now := time.Now()
	yearMonth := fmt.Sprintf("%d年%d月份", now.Year(), now.Month())

	// 创建Excel
	f := excelize.NewFile()
	sheetName := server.Name
	f.SetSheetName("Sheet1", sheetName)

	// ---- 样式定义 ----
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	headerInfoStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical: "center",
			WrapText: true,
		},
	})

	tableHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#D9E1F2"},
		},
	})

	dataCellStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   10,
			Family: "微软雅黑",
		},
		Alignment: &excelize.Alignment{
			Vertical:   "center",
			WrapText:   true,
			Horizontal: "left",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "#000000", Style: 1},
			{Type: "top", Color: "#000000", Style: 1},
			{Type: "right", Color: "#000000", Style: 1},
			{Type: "bottom", Color: "#000000", Style: 1},
		},
	})

	// 列数: 序号/账号名/创建时间/有效期/权限/所属对接人/所属部门/白名单IP/确认人/确认时间 = 10列
	lastColIndex := 9 // J列 (0-indexed)
	lastCol := string(rune('A' + lastColIndex))

	// ---- Row 1: 标题 ----
	f.SetCellValue(sheetName, "A1", "SFTP账号月度确认表")
	f.MergeCell(sheetName, "A1", lastCol+"1")
	f.SetCellStyle(sheetName, "A1", lastCol+"1", titleStyle)
	f.SetRowHeight(sheetName, 1, 36)

	// ---- Row 2: 服务器信息 + 确认月份 ----
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("服务器：%s", server.Name))
	f.MergeCell(sheetName, "A2", "E2")
	f.SetCellStyle(sheetName, "A2", "E2", headerInfoStyle)

	f.SetCellValue(sheetName, "F2", fmt.Sprintf("确认月份：%s", yearMonth))
	f.MergeCell(sheetName, "F2", lastCol+"2")
	f.SetCellStyle(sheetName, "F2", lastCol+"2", headerInfoStyle)
	f.SetRowHeight(sheetName, 2, 30)

	// ---- Row 3: 表头 ----
	headers := []string{"序号", "账号名", "创建时间", "有效期", "权限", "所属对接人", "所属部门", "白名单IP", "确认人", "确认时间"}
	for i, h := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, col+"3", h)
		f.SetCellStyle(sheetName, col+"3", col+"3", tableHeaderStyle)
	}
	f.SetRowHeight(sheetName, 3, 28)

	// ---- Row 4+: 数据行 ----
	rowNum := 4
	for i, account := range accounts {
		// 解析权限
		permissions := "-"
		if account.PermissionsJSON != "" && account.PermissionsJSON != "[]" {
			var perms []string
			if err := json.Unmarshal([]byte(account.PermissionsJSON), &perms); err == nil {
				permLabels := []string{}
				for _, p := range perms {
					if p == "read" {
						permLabels = append(permLabels, "读")
					} else if p == "write" {
						permLabels = append(permLabels, "写")
					}
				}
				if len(permLabels) > 0 {
					permissions = joinWithSep(permLabels, "、")
				}
			}
		}

		// 解析白名单
		whitelist := "-"
		if account.WhitelistJSON != "" && account.WhitelistJSON != "[]" {
			var ips []string
			if err := json.Unmarshal([]byte(account.WhitelistJSON), &ips); err == nil {
				if len(ips) > 0 {
					whitelist = joinWithSep(ips, "\n")
				}
			}
		}

		// 部门名称
		deptName := "-"
		if account.DepartmentID > 0 {
			if name, ok := deptMap[account.DepartmentID]; ok {
				deptName = name
			}
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), i+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", rowNum), account.AccountName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", rowNum), account.CreatedTime)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", rowNum), account.Validity)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", rowNum), permissions)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", rowNum), account.ContactPerson)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", rowNum), deptName)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", rowNum), whitelist)
		// I列: 确认人 - 留空
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", rowNum), "")
		// J列: 确认时间 - 留空
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", rowNum), "")

		// 设置数据行样式
		for j := 0; j < 10; j++ {
			col := string(rune('A' + j))
			f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", col, rowNum), fmt.Sprintf("%s%d", col, rowNum), dataCellStyle)
		}

		rowNum++
	}

	// ---- 空行 ----
	rowNum++

	// ---- 复核行 ----
	footerContent := "复核部门：IT          复核人：                              复核日期："
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", rowNum), footerContent)
	f.MergeCell(sheetName, fmt.Sprintf("A%d", rowNum), lastCol+fmt.Sprintf("%d", rowNum))
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", rowNum), lastCol+fmt.Sprintf("%d", rowNum), headerInfoStyle)
	f.SetRowHeight(sheetName, rowNum, 30)

	// ---- 设置列宽 ----
	colWidths := map[string]float64{
		"A": 6,   // 序号
		"B": 18,  // 账号名
		"C": 14,  // 创建时间
		"D": 14,  // 有效期
		"E": 12,  // 权限
		"F": 14,  // 所属对接人
		"G": 14,  // 所属部门
		"H": 25,  // 白名单IP
		"I": 14,  // 确认人
		"J": 14,  // 确认时间
	}
	for col, width := range colWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	// 输出
	fileName := fmt.Sprintf("SFTP账号确认表(%s)-%s.xlsx", yearMonth, server.Name)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", fileName))
	f.Write(c.Writer)
}

// joinWithSep 用指定分隔符拼接字符串
func joinWithSep(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
