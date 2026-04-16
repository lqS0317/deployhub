package config

import (
	"bytes"
	"fmt"
	"text/template"
)

// safeTemplateFuncs 安全的模板函数集合，不包含任何危险操作
var safeTemplateFuncs = template.FuncMap{}

// RenderTemplate 使用 Go text/template 渲染配置模板
// 变量缺失时返回错误（missingkey=error）
func RenderTemplate(templateContent string, vars map[string]string) (string, error) {
	tmpl, err := template.New("config").
		Option("missingkey=error").
		Funcs(safeTemplateFuncs).
		Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("模板解析失败: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("模板渲染失败: %w", err)
	}

	return buf.String(), nil
}
