package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"deployhub/internal/pkg"
	"deployhub/internal/service/gitrepo"

	"github.com/gin-gonic/gin"
)

// GitRepoHandler Git 仓库处理器
type GitRepoHandler struct {
	svc *gitrepo.GitRepoService
}

// NewGitRepoHandler 创建 Git 仓库处理器
func NewGitRepoHandler(svc *gitrepo.GitRepoService) *GitRepoHandler {
	return &GitRepoHandler{svc: svc}
}

type createGitRepoRequest struct {
	Name       string `json:"name" binding:"required"`
	URL        string `json:"url" binding:"required"`
	Provider   string `json:"provider" binding:"required"`
	AuthType   string `json:"auth_type" binding:"required"`
	Credential string `json:"credential" binding:"required"`
}

// Create 创建 Git 仓库
func (h *GitRepoHandler) Create(c *gin.Context) {
	var req createGitRepoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	result, err := h.svc.Create(req.Name, req.URL, req.Provider, req.AuthType, req.Credential)
	if err != nil {
		pkg.Error(c, http.StatusConflict, pkg.CodeConflict, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": result.ID, "name": result.Name, "url": result.URL,
		"provider": result.Provider, "auth_type": result.AuthType, "created_at": result.CreatedAt,
	})
}

// List 列出 Git 仓库
func (h *GitRepoHandler) List(c *gin.Context) {
	page, pageSize := pkg.GetPagination(c)
	items, total, err := h.svc.List(page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "查询失败")
		return
	}
	pkg.Paginated(c, items, total, page, pageSize)
}

// Get 获取单个 Git 仓库
func (h *GitRepoHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	result, err := h.svc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "Git 仓库不存在")
		return
	}
	c.JSON(http.StatusOK, result)
}

// Update 更新 Git 仓库
func (h *GitRepoHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		Name       string  `json:"name"`
		URL        string  `json:"url"`
		Credential *string `json:"credential"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "参数校验失败")
		return
	}
	result, err := h.svc.Update(uint(id), req.Name, req.URL, req.Credential)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, result)
}

// Delete 删除 Git 仓库
func (h *GitRepoHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.svc.Delete(uint(id)); err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// TestConnection 测试 Git 仓库连通性
func (h *GitRepoHandler) TestConnection(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	repo, err := h.svc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "Git 仓库不存在")
		return
	}

	decrypted, err := h.svc.GetDecryptedCredential(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "解密凭证失败"})
		return
	}

	ok, msg := testGitConnection(repo.URL, repo.Provider, repo.AuthType, decrypted)
	c.JSON(http.StatusOK, gin.H{"success": ok, "message": msg})
}

// ListBranches 获取 Git 仓库分支列表
func (h *GitRepoHandler) ListBranches(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	repo, err := h.svc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "Git 仓库不存在")
		return
	}

	decrypted, err := h.svc.GetDecryptedCredential(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "解密凭证失败")
		return
	}

	branches, err := fetchBranches(repo.URL, repo.Provider, repo.AuthType, decrypted)
	if err != nil {
		pkg.Error(c, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"branches": branches})
}

// ListCommits 获取 Git 仓库指定分支的 commit 列表
func (h *GitRepoHandler) ListCommits(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	branch := c.Query("branch")
	if branch == "" {
		branch = "main"
	}

	repo, err := h.svc.GetByID(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "Git 仓库不存在")
		return
	}

	decrypted, err := h.svc.GetDecryptedCredential(uint(id))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, pkg.CodeInternalError, "解密凭证失败")
		return
	}

	commits, err := fetchCommits(repo.URL, repo.Provider, repo.AuthType, decrypted, branch)
	if err != nil {
		pkg.Error(c, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"commits": commits})
}

// RegisterGitRepoRoutes 注册 Git 仓库路由
func RegisterGitRepoRoutes(r *gin.RouterGroup, h *GitRepoHandler) {
	repos := r.Group("/git-repos")
	{
		repos.GET("", h.List)
		repos.POST("", h.Create)
		repos.GET("/:id", h.Get)
		repos.PUT("/:id", h.Update)
		repos.DELETE("/:id", h.Delete)
		repos.POST("/:id/test-connection", h.TestConnection)
		repos.GET("/:id/branches", h.ListBranches)
		repos.GET("/:id/commits", h.ListCommits)
	}
}

type branchInfo struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// fetchBranches 通过 Git 提供商 API 获取分支列表
func fetchBranches(repoURL, provider, authType, credential string) ([]branchInfo, error) {
	ownerRepo := extractOwnerRepo(repoURL)
	if ownerRepo == "" {
		return nil, fmt.Errorf("无法解析仓库 owner/repo")
	}

	client := &http.Client{Timeout: 15 * time.Second}

	switch provider {
	case "github":
		return fetchGitHubBranches(client, ownerRepo, authType, credential)
	case "gitlab":
		return fetchGitLabBranches(client, repoURL, ownerRepo, authType, credential)
	default:
		return []branchInfo{{Name: "main", IsDefault: true}}, nil
	}
}

func fetchGitHubBranches(client *http.Client, ownerRepo, authType, credential string) ([]branchInfo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/branches?per_page=100", ownerRepo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "DeployHub/1.0")
	req.Header.Set("Accept", "application/vnd.github+json")
	if credential != "" && authType == "token" {
		req.Header.Set("Authorization", "Bearer "+credential)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回 %d", resp.StatusCode)
	}

	var ghBranches []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghBranches); err != nil {
		return nil, fmt.Errorf("解析 GitHub 分支数据失败: %w", err)
	}

	// 获取默认分支
	defaultBranch := getGitHubDefaultBranch(client, ownerRepo, authType, credential)

	result := make([]branchInfo, 0, len(ghBranches))
	for _, b := range ghBranches {
		result = append(result, branchInfo{
			Name:      b.Name,
			IsDefault: b.Name == defaultBranch,
		})
	}
	return result, nil
}

func getGitHubDefaultBranch(client *http.Client, ownerRepo, authType, credential string) string {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s", ownerRepo)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "DeployHub/1.0")
	if credential != "" && authType == "token" {
		req.Header.Set("Authorization", "Bearer "+credential)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "main"
	}
	defer resp.Body.Close()

	var repo struct {
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil || repo.DefaultBranch == "" {
		return "main"
	}
	return repo.DefaultBranch
}

func fetchGitLabBranches(client *http.Client, repoURL, ownerRepo, authType, credential string) ([]branchInfo, error) {
	// GitLab 项目 ID 用 URL 编码的 owner/repo
	encodedPath := strings.ReplaceAll(ownerRepo, "/", "%2F")
	host := "gitlab.com"
	if parts := strings.Split(repoURL, "/"); len(parts) > 2 {
		host = parts[2]
	}
	apiURL := fmt.Sprintf("https://%s/api/v4/projects/%s/repository/branches?per_page=100", host, encodedPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	if credential != "" && authType == "token" {
		req.Header.Set("PRIVATE-TOKEN", credential)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitLab API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API 返回 %d", resp.StatusCode)
	}

	var glBranches []struct {
		Name    string `json:"name"`
		Default bool   `json:"default"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&glBranches); err != nil {
		return nil, fmt.Errorf("解析 GitLab 分支数据失败: %w", err)
	}

	result := make([]branchInfo, 0, len(glBranches))
	for _, b := range glBranches {
		result = append(result, branchInfo{Name: b.Name, IsDefault: b.Default})
	}
	return result, nil
}

// testGitConnection 通过 HTTP HEAD/GET 测试 Git 仓库 API 连通性
func testGitConnection(repoURL, provider, authType, credential string) (bool, string) {
	apiURL := ""
	switch provider {
	case "github":
		parts := extractOwnerRepo(repoURL)
		if parts == "" {
			return false, "无法解析仓库 owner/repo"
		}
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s", parts)
	case "gitlab":
		apiURL = strings.TrimSuffix(repoURL, ".git") + "/-/raw/HEAD/README.md"
	default:
		apiURL = repoURL
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, fmt.Sprintf("构建请求失败: %v", err)
	}

	if credential != "" {
		switch authType {
		case "token":
			req.Header.Set("Authorization", "Bearer "+credential)
		}
	}
	req.Header.Set("User-Agent", "DeployHub/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("连接失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, "连接成功"
	}
	return false, fmt.Sprintf("返回状态码 %d", resp.StatusCode)
}

type commitInfo struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Date    string `json:"date"`
}

// fetchCommits 获取指定分支最近 20 条 commit
func fetchCommits(repoURL, provider, authType, credential, branch string) ([]commitInfo, error) {
	ownerRepo := extractOwnerRepo(repoURL)
	if ownerRepo == "" {
		return nil, fmt.Errorf("无法解析仓库 owner/repo")
	}
	client := &http.Client{Timeout: 15 * time.Second}

	switch provider {
	case "github":
		return fetchGitHubCommits(client, ownerRepo, authType, credential, branch)
	case "gitlab":
		return fetchGitLabCommits(client, repoURL, ownerRepo, authType, credential, branch)
	default:
		return []commitInfo{}, nil
	}
}

func fetchGitHubCommits(client *http.Client, ownerRepo, authType, credential, branch string) ([]commitInfo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/commits?sha=%s&per_page=20", ownerRepo, branch)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("User-Agent", "DeployHub/1.0")
	req.Header.Set("Accept", "application/vnd.github+json")
	if credential != "" && authType == "token" {
		req.Header.Set("Authorization", "Bearer "+credential)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回 %d", resp.StatusCode)
	}

	var ghCommits []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Message string `json:"message"`
			Author  struct {
				Name string `json:"name"`
				Date string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghCommits); err != nil {
		return nil, fmt.Errorf("解析 commit 数据失败: %w", err)
	}

	result := make([]commitInfo, 0, len(ghCommits))
	for _, c := range ghCommits {
		msg := c.Commit.Message
		if idx := strings.Index(msg, "\n"); idx > 0 {
			msg = msg[:idx]
		}
		result = append(result, commitInfo{
			SHA:     c.SHA,
			Message: msg,
			Author:  c.Commit.Author.Name,
			Date:    c.Commit.Author.Date,
		})
	}
	return result, nil
}

func fetchGitLabCommits(client *http.Client, repoURL, ownerRepo, authType, credential, branch string) ([]commitInfo, error) {
	encodedPath := strings.ReplaceAll(ownerRepo, "/", "%2F")
	host := "gitlab.com"
	if parts := strings.Split(repoURL, "/"); len(parts) > 2 {
		host = parts[2]
	}
	apiURL := fmt.Sprintf("https://%s/api/v4/projects/%s/repository/commits?ref_name=%s&per_page=20", host, encodedPath, branch)

	req, _ := http.NewRequest("GET", apiURL, nil)
	if credential != "" && authType == "token" {
		req.Header.Set("PRIVATE-TOKEN", credential)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitLab API 失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API 返回 %d", resp.StatusCode)
	}

	var glCommits []struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		AuthorName string `json:"author_name"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&glCommits); err != nil {
		return nil, fmt.Errorf("解析 commit 数据失败: %w", err)
	}

	result := make([]commitInfo, 0, len(glCommits))
	for _, c := range glCommits {
		msg := c.Message
		if idx := strings.Index(msg, "\n"); idx > 0 {
			msg = msg[:idx]
		}
		result = append(result, commitInfo{SHA: c.ID, Message: msg, Author: c.AuthorName, Date: c.CreatedAt})
	}
	return result, nil
}

// extractOwnerRepo 从 GitHub URL 提取 owner/repo
func extractOwnerRepo(rawURL string) string {
	rawURL = strings.TrimSuffix(rawURL, ".git")
	rawURL = strings.TrimSuffix(rawURL, "/")
	parts := strings.Split(rawURL, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	return ""
}
