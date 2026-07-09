package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// DynamicGenerator 动态文件生成函数签名
// 接收 gin.Context，直接写入响应（复用现有 handler 的输出逻辑）
type DynamicGenerator func(c *gin.Context)

// dynamicRegistry 动态生成器注册表
var dynamicRegistry = map[string]DynamicGenerator{}

// RegisterDynamicGenerator 注册动态生成器
func RegisterDynamicGenerator(name string, generator DynamicGenerator) {
	dynamicRegistry[name] = generator
}

// GetDynamicGenerator 获取已注册的生成器
func GetDynamicGenerator(name string) (DynamicGenerator, error) {
	gen, ok := dynamicRegistry[name]
	if !ok {
		return nil, fmt.Errorf("未找到动态生成器: %s", name)
	}
	return gen, nil
}

func init() {
	// 注册: 用户变更记录表（复用 ExportChangeRecord 的 Excel 生成逻辑）
	// 注意: ExportChangeRecord 直接写入 c.Writer，所以可以直接作为生成器使用
	RegisterDynamicGenerator("export_user_change_record", ExportChangeRecord)
}
