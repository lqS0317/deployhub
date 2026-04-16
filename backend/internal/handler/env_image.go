package handler

import (
	"net/http"
	"strconv"

	"deployhub/internal/pkg"
	"deployhub/internal/service/deploy"
	"deployhub/internal/service/svc"

	"github.com/gin-gonic/gin"
)

// EnvImageHandler 从 Git 仓库解析 app-env.yaml 镜像信息
type EnvImageHandler struct {
	fetcher    *deploy.EnvImageFetcher
	svcService *svc.ServiceService
}

func NewEnvImageHandler(fetcher *deploy.EnvImageFetcher, svcService *svc.ServiceService) *EnvImageHandler {
	return &EnvImageHandler{fetcher: fetcher, svcService: svcService}
}

// GetEnvImage 解析服务对应的 app-env.yaml 返回镜像信息
func (h *EnvImageHandler) GetEnvImage(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	service, err := h.svcService.GetByID(uint(serviceID))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, pkg.CodeNotFound, "服务不存在")
		return
	}

	if service.HelmRepoID == nil {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "未配置 Helm Chart Git 仓库")
		return
	}

	envFilePath := service.HelmEnvFilePath
	if envFilePath == "" {
		pkg.Error(c, http.StatusBadRequest, pkg.CodeBadRequest, "未配置 env 文件路径")
		return
	}

	branch := service.HelmChartBranch
	if branch == "" {
		branch = "main"
	}

	info, err := h.fetcher.FetchEnvImage(*service.HelmRepoID, branch, envFilePath)
	if err != nil {
		pkg.Error(c, http.StatusBadGateway, "UPSTREAM_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, info)
}
