package deploy

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"deployhub/internal/repository"
	"deployhub/internal/service/crypto"
)

// EnvImageFetcher 从 Git 仓库获取 app-env.yaml 并解析镜像信息
type EnvImageFetcher struct {
	gitRepoRepo repository.GitRepoRepository
	cryptoSvc   *crypto.CryptoService
}

func NewEnvImageFetcher(gitRepoRepo repository.GitRepoRepository, cryptoSvc *crypto.CryptoService) *EnvImageFetcher {
	return &EnvImageFetcher{gitRepoRepo: gitRepoRepo, cryptoSvc: cryptoSvc}
}

// FetchEnvImage 通过 GitHub/GitLab Raw API 获取文件并解析
func (f *EnvImageFetcher) FetchEnvImage(gitRepoID uint, branch, filePath string) (*EnvImageInfo, error) {
	gitRepo, err := f.gitRepoRepo.FindByID(gitRepoID)
	if err != nil {
		return nil, fmt.Errorf("Git 仓库不存在: %w", err)
	}

	credential, err := f.cryptoSvc.Decrypt(gitRepo.CredentialEncrypted)
	if err != nil {
		return nil, fmt.Errorf("解密 Git 凭证失败: %w", err)
	}

	if branch == "" {
		branch = "main"
	}

	// 构造 raw 文件 URL
	rawURL := buildRawFileURL(gitRepo.URL, gitRepo.Provider, branch, filePath, credential)
	if rawURL == "" {
		return nil, fmt.Errorf("不支持的 Git 提供商: %s", gitRepo.Provider)
	}

	// 请求文件内容
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if credential != "" {
		switch gitRepo.Provider {
		case "github":
			req.Header.Set("Authorization", "Bearer "+credential)
			req.Header.Set("Accept", "application/vnd.github.raw+json")
		case "gitlab":
			req.Header.Set("PRIVATE-TOKEN", credential)
		}
	}
	req.Header.Set("User-Agent", "DeployHub/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取文件失败: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}

	return ParseEnvImage(string(body))
}

// buildRawFileURL 构造 Git 提供商的 raw 文件 URL
func buildRawFileURL(repoURL, provider, branch, filePath, credential string) string {
	repoURL = strings.TrimSuffix(repoURL, ".git")
	repoURL = strings.TrimSuffix(repoURL, "/")

	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return ""
	}
	ownerRepo := parts[len(parts)-2] + "/" + parts[len(parts)-1]

	switch provider {
	case "github":
		return fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s", ownerRepo, filePath, branch)
	case "gitlab":
		host := "gitlab.com"
		if len(parts) > 4 {
			host = parts[2]
		}
		encoded := strings.ReplaceAll(ownerRepo, "/", "%2F")
		encodedPath := strings.ReplaceAll(filePath, "/", "%2F")
		return fmt.Sprintf("https://%s/api/v4/projects/%s/repository/files/%s/raw?ref=%s", host, encoded, encodedPath, branch)
	default:
		return ""
	}
}
