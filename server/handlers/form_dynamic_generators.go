package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// DynamicGenerator 动态文件生成函数签名
// 接收 gin.Context，直接写入响应（复用现有 handler 的输出逻辑）
type DynamicGenerator func(c *gin.Context)

// GeneratorParam 描述动态生成器所需参数
type GeneratorParam struct {
	Name     string `json:"name"`              // 参数名，如 department_id
	Label    string `json:"label"`             // 显示标签，如 "选择部门"
	Type     string `json:"type"`              // 参数类型：select, input, date等
	Source   string `json:"source,omitempty"`  // 数据来源API路径（仅select类型）
	Required bool   `json:"required"`          // 是否必填
}

// DynamicGeneratorInfo 动态生成器完整信息
type DynamicGeneratorInfo struct {
	Name     string           `json:"name"`      // 生成器名称
	Handler  DynamicGenerator `json:"-"`         // 处理函数
	Params   []GeneratorParam `json:"params"`    // 所需参数列表
	FileName string           `json:"file_name"` // 默认文件名模板
}

// dynamicRegistry 动态生成器注册表
var dynamicRegistry = map[string]*DynamicGeneratorInfo{}

// RegisterDynamicGenerator 注册动态生成器
func RegisterDynamicGenerator(name string, handler DynamicGenerator, params []GeneratorParam, fileName string) {
	dynamicRegistry[name] = &DynamicGeneratorInfo{
		Name:     name,
		Handler:  handler,
		Params:   params,
		FileName: fileName,
	}
}

// GetDynamicGenerator 获取已注册的生成器
func GetDynamicGenerator(name string) (*DynamicGeneratorInfo, error) {
	info, ok := dynamicRegistry[name]
	if !ok {
		return nil, fmt.Errorf("未找到动态生成器: %s", name)
	}
	return info, nil
}

// ListDynamicGenerators 列出所有注册的动态生成器（供前端查询）
func ListDynamicGenerators() []*DynamicGeneratorInfo {
	list := make([]*DynamicGeneratorInfo, 0, len(dynamicRegistry))
	for _, info := range dynamicRegistry {
		list = append(list, info)
	}
	return list
}

func init() {
	// 注册: 用户变更记录表（无参数）
	RegisterDynamicGenerator("export_user_change_record", ExportChangeRecord, nil, "用户变更记录表.xlsx")

	// 注册: 部门用户确认表（需要部门ID参数）
	RegisterDynamicGenerator("export_department_confirmation", ExportDepartmentConfirmation,
		[]GeneratorParam{
			{Name: "department_id", Label: "选择部门", Type: "select", Source: "/api/departments", Required: true},
		},
		"部门用户确认表.xlsx")
}
