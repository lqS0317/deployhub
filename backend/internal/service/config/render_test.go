package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate_BasicVariableSubstitution(t *testing.T) {
	tpl := `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.APP_NAME}}
data:
  DB_HOST: "{{.DB_HOST}}"
  DB_PORT: "{{.DB_PORT}}"`

	vars := map[string]string{
		"APP_NAME": "my-app",
		"DB_HOST":  "postgres.svc",
		"DB_PORT":  "5432",
	}

	result, err := RenderTemplate(tpl, vars)
	assert.NoError(t, err)
	assert.Contains(t, result, "name: my-app")
	assert.Contains(t, result, `DB_HOST: "postgres.svc"`)
	assert.Contains(t, result, `DB_PORT: "5432"`)
}

func TestRenderTemplate_MissingVariable(t *testing.T) {
	tpl := `host: {{.DB_HOST}}
port: {{.DB_PORT}}`

	vars := map[string]string{
		"DB_HOST": "localhost",
	}

	_, err := RenderTemplate(tpl, vars)
	assert.Error(t, err, "缺少变量时应返回错误")
}

func TestRenderTemplate_ComplexTemplate(t *testing.T) {
	tpl := `{{range $key, $val := .}}
{{$key}}: "{{$val}}"
{{end}}`

	vars := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
	}

	result, err := RenderTemplate(tpl, vars)
	assert.NoError(t, err)
	assert.Contains(t, result, `KEY1: "value1"`)
	assert.Contains(t, result, `KEY2: "value2"`)
}

func TestRenderTemplate_ParseError(t *testing.T) {
	tpl := `{{.UNCLOSED`
	vars := map[string]string{"UNCLOSED": "val"}

	_, err := RenderTemplate(tpl, vars)
	assert.Error(t, err, "模板语法错误时应返回错误")
}

func TestRenderTemplate_EmptyVars(t *testing.T) {
	tpl := `static content only`
	vars := map[string]string{}

	result, err := RenderTemplate(tpl, vars)
	assert.NoError(t, err)
	assert.Equal(t, "static content only", result)
}
