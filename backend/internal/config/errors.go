package config

import "errors"

var (
	ErrMissingJWTSecret    = errors.New("JWT_SECRET 环境变量未设置")
	ErrMissingAESKey       = errors.New("AES_KEY 环境变量未设置")
	ErrInvalidAESKeyLength = errors.New("AES_KEY 必须为 64 位十六进制字符串（32 字节）")
)
