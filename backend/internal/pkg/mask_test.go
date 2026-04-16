package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"标准11位号码", "13812345678", "138****5678"},
		{"空字符串", "", ""},
		{"短于7位", "12345", "12345"},
		{"恰好7位", "1234567", "123****4567"},
		{"带区号", "+8613812345678", "+86****5678"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, MaskPhone(tt.input))
		})
	}
}
