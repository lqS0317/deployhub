package config

import (
	"fmt"
	"strings"
)

// DiffVersions 对比两个版本的配置内容，返回逐行 diff 结果
// 使用简单的逐行比较算法，输出类 unified diff 格式
func DiffVersions(old, new string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	lcs := longestCommonSubsequence(oldLines, newLines)

	var result []string
	oi, ni, li := 0, 0, 0

	for li < len(lcs) {
		for oi < len(oldLines) && oldLines[oi] != lcs[li] {
			result = append(result, fmt.Sprintf("- %s", oldLines[oi]))
			oi++
		}
		for ni < len(newLines) && newLines[ni] != lcs[li] {
			result = append(result, fmt.Sprintf("+ %s", newLines[ni]))
			ni++
		}
		result = append(result, fmt.Sprintf("  %s", lcs[li]))
		oi++
		ni++
		li++
	}

	for oi < len(oldLines) {
		result = append(result, fmt.Sprintf("- %s", oldLines[oi]))
		oi++
	}
	for ni < len(newLines) {
		result = append(result, fmt.Sprintf("+ %s", newLines[ni]))
		ni++
	}

	return strings.Join(result, "\n")
}

// longestCommonSubsequence 计算两个字符串切片的最长公共子序列
func longestCommonSubsequence(a, b []string) []string {
	m, n := len(a), len(b)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	lcs := make([]string, 0, dp[m][n])
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append([]string{a[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}
