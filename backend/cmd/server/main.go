package main

import (
	"log"
	"net/http"

	"deployhub/internal/config"
	"deployhub/internal/handler"
	"deployhub/internal/middleware"
	"deployhub/internal/model"
	"deployhub/internal/repository"
	"deployhub/internal/service/approval"
	"deployhub/internal/service/audit"
	"deployhub/internal/service/auth"
	"deployhub/internal/service/build"
	"deployhub/internal/service/cluster"
	configSvc "deployhub/internal/service/config"
	"deployhub/internal/service/configcenter"
	"deployhub/internal/service/crypto"
	"deployhub/internal/service/deploy"
	groupSvc "deployhub/internal/service/group"
	"deployhub/internal/service/routing"
	"deployhub/internal/service/storage"
	"deployhub/internal/service/gitrepo"
	"deployhub/internal/service/notification"
	"deployhub/internal/service/registry"
	"deployhub/internal/service/setting"
	"deployhub/internal/service/svc"
	"deployhub/internal/ws"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := model.InitDB(cfg.DatabaseURL, cfg.LogLevel)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 初始化加密服务
	cryptoSvc, err := crypto.NewCryptoService(cfg.AESKey)
	if err != nil {
		log.Fatalf("加密服务初始化失败: %v", err)
	}

	// 初始化对象存储
	storageSvc := storage.NewStorageService(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3Region)

	// 初始化 Repository
	userRepo := repository.NewUserRepository(db)
	clusterRepo := repository.NewClusterRepository(db)
	gitRepoRepo := repository.NewGitRepoRepository(db)
	registryRepo := repository.NewRegistryRepository(db)
	serviceRepo := repository.NewServiceRepository(db)
	memberRepo := repository.NewServiceMemberRepository(db)
	buildRepo := repository.NewBuildRepository(db)
	deployRepo := repository.NewDeploymentRepository(db)
	approvalRepo := repository.NewApprovalRepository(db)
	configTemplateRepo := repository.NewConfigTemplateRepository(db)
	configEnvValueRepo := repository.NewConfigEnvValueRepository(db)
	configVersionRepo := repository.NewConfigVersionRepository(db)
	configDeployRepo := repository.NewConfigDeploymentRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	notifChannelRepo := repository.NewNotificationChannelRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	helmValuesRepo := repository.NewHelmValuesRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	groupMemberRepo := repository.NewGroupMemberRepository(db)
	groupPermRepo := repository.NewGroupPermissionRepository(db)
	systemSettingRepo := repository.NewSystemSettingRepository(db)
	notifRuleRepo := repository.NewNotificationRuleRepository(db)
	svcNotifRuleRepo := repository.NewServiceNotificationRuleRepository(db)
	notifLogRepo := repository.NewNotificationLogRepository(db)
	configEntryRepo := repository.NewConfigEntryRepository(db)
	configItemRepo := repository.NewConfigItemRepository(db)
	configReleaseRepo := repository.NewConfigReleaseRepository(db)
	configPermRepo := repository.NewConfigPermissionRepository(db)
	routeEntryRepo := repository.NewRouteEntryRepository(db)
	routeDeploymentRepo := repository.NewRouteDeploymentRepository(db)
	routePermissionRepo := repository.NewRoutePermissionRepository(db)
	routePluginRepo := repository.NewRoutePluginRepository(db)
	pluginDeploymentRepo := repository.NewPluginDeploymentRepository(db)
	// 初始化 Service
	jwtSvc := auth.NewJWTService(cfg.JWTSecret)
	authSvc := auth.NewAuthService(userRepo, jwtSvc, cryptoSvc)
	clusterSvc := cluster.NewClusterService(clusterRepo, cryptoSvc)
	clientPool := cluster.NewClientsetPool(clusterSvc)
	gitRepoSvc := gitrepo.NewGitRepoService(gitRepoRepo, cryptoSvc)
	registrySvc := registry.NewRegistryService(registryRepo, cryptoSvc)
	serviceSvc := svc.NewServiceService(serviceRepo, memberRepo)
	memberSvc := svc.NewMemberService(memberRepo)
	effectiveRoleSvc := svc.NewEffectiveRoleService(memberRepo, groupPermRepo, userRepo, serviceRepo)
	rbacSvc := svc.NewRBACService(effectiveRoleSvc)
	grpSvc := groupSvc.NewGroupService(groupRepo, groupMemberRepo, groupPermRepo)
	settingSvc := setting.NewSettingService(systemSettingRepo)
	// 初始化 WebSocket Hub（需在 buildExecutor 之前）
	wsHub := ws.NewHub()
	go wsHub.Run()

	// 初始化通知调度器（需在 buildExecutor / approvalSvc 等之前）
	webhookSender := notification.NewWebhookSender()
	notifDispatcher := notification.NewDispatcher(notifRuleRepo, svcNotifRuleRepo, notifLogRepo, webhookSender)

	buildSvc := build.NewBuildService(buildRepo, serviceRepo)
	buildExecutor := build.NewBuildExecutor(clientPool, buildRepo, serviceRepo, registryRepo, gitRepoRepo, clusterRepo, cryptoSvc, wsHub, notifDispatcher)
	deploySvc := deploy.NewDeployService(deployRepo, serviceRepo, buildRepo)
	approvalSvc := approval.NewApprovalService(approvalRepo, deployRepo, userRepo, notifDispatcher)
	cfgSvc := configSvc.NewConfigService(configTemplateRepo, configEnvValueRepo, configVersionRepo, configDeployRepo, cryptoSvc, clientPool)
	notifSvc := notification.NewNotificationService(notifRepo)
	auditSvc := audit.NewAuditService(auditLogRepo)
	configCenterSvc := configcenter.NewConfigService(configEntryRepo, configItemRepo, configReleaseRepo, cryptoSvc)
	configPermSvc := configcenter.NewConfigPermissionService(configPermRepo, userRepo, serviceRepo)

	// 初始化路由中心
	k8sRouteDeployer := routing.NewK8sRouteDeployer(clientPool)
	routeSvc := routing.NewRouteService(routeEntryRepo, routeDeploymentRepo, k8sRouteDeployer)
	pluginSvc := routing.NewPluginService(routePluginRepo, pluginDeploymentRepo, k8sRouteDeployer)
	routePermSvc := routing.NewPermissionService(routePermissionRepo)

	// 初始化 Deploy 组件
	executor := deploy.NewDirectExecutor(clientPool)
	// 配置中心部署助手：部署时自动下发 ConfigMap/Secret
	configDeployHelper := deploy.NewConfigDeployHelper(configCenterSvc, cryptoSvc)
	if executor != nil {
		executor.SetConfigDeployHelper(configDeployHelper)
	}
	helmExecutor := deploy.NewHelmExecutor(clientPool, deployRepo, gitRepoRepo, helmValuesRepo, cryptoSvc, wsHub, clusterRepo, settingSvc, notifDispatcher)
	rolloutWatcher := deploy.NewRolloutWatcher(clientPool, deploySvc, wsHub, notifDispatcher)
	rolloutWatcher.RecoverStuckDeployments()

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authSvc, jwtSvc, storageSvc)
	clusterHandler := handler.NewClusterHandler(clusterSvc, clientPool)
	gitRepoHandler := handler.NewGitRepoHandler(gitRepoSvc)
	registryHandler := handler.NewRegistryHandler(registrySvc)
	serviceHandler := handler.NewServiceHandler(serviceSvc, memberSvc, effectiveRoleSvc)
	buildHandler := handler.NewBuildHandler(buildSvc, buildExecutor, effectiveRoleSvc)
	deployHandler := handler.NewDeployHandler(deploySvc, executor, helmExecutor, rolloutWatcher, effectiveRoleSvc, approvalSvc, userRepo, notifDispatcher)
	approvalHandler := handler.NewApprovalHandler(approvalSvc)
	configHandler := handler.NewConfigHandler(cfgSvc)
	notifHandler := handler.NewNotificationHandler(notifSvc)
	notifChannelHandler := handler.NewNotificationChannelHandler(notifChannelRepo, webhookSender)
	auditHandler := handler.NewAuditHandler(auditSvc)
	userAdminHandler := handler.NewUserAdminHandler(userRepo, authSvc, groupMemberRepo)
	clusterNsRepo := repository.NewClusterNamespaceRepository(db)
	envImageFetcher := deploy.NewEnvImageFetcher(gitRepoRepo, cryptoSvc)
	envImageHandler := handler.NewEnvImageHandler(envImageFetcher, serviceSvc)
	namespaceHandler := handler.NewNamespaceHandler(clusterNsRepo, clientPool)
	groupHandler := handler.NewGroupHandler(grpSvc)
	permissionHandler := handler.NewPermissionHandler(effectiveRoleSvc)
	helmValuesHandler := handler.NewHelmValuesHandler(helmValuesRepo)
	systemSettingHandler := handler.NewSystemSettingHandler(settingSvc)
	notifRuleHandler := handler.NewNotificationRuleHandler(notifRuleRepo, svcNotifRuleRepo, notifLogRepo)
	configCenterHandler := handler.NewConfigCenterHandler(configCenterSvc, configPermSvc)
	routeEntryHandler := handler.NewRouteEntryHandler(routeSvc)
	routePluginHandler := handler.NewRoutePluginHandler(pluginSvc)
	routePermHandler := handler.NewRoutePermissionHandler(routePermSvc)

	r := gin.Default()

	// 注册全局中间件：CORS + 结构化日志
	r.Use(middleware.CORS(), middleware.RequestLogger())

	// 健康检查
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "deployhub",
		})
	})

	// 认证路由（公开）
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/me", middleware.JWTAuth(jwtSvc), authHandler.GetMe)
		authGroup.PUT("/profile", middleware.JWTAuth(jwtSvc), authHandler.UpdateProfile)
		authGroup.PUT("/password", middleware.JWTAuth(jwtSvc), authHandler.ChangePassword)
		authGroup.POST("/avatar", middleware.JWTAuth(jwtSvc), authHandler.UploadAvatar)
		authGroup.POST("/logout", middleware.JWTAuth(jwtSvc), authHandler.Logout)
	}

	// 需认证的 API 路由
	api := r.Group("/api/v1", middleware.JWTAuth(jwtSvc), middleware.AuditLog(auditSvc))
	{
		handler.RegisterClusterRoutes(api, clusterHandler)
		handler.RegisterGitRepoRoutes(api, gitRepoHandler)
		handler.RegisterRegistryRoutes(api, registryHandler)
		handler.RegisterServiceRoutes(api, serviceHandler, rbacSvc)
		handler.RegisterBuildRoutes(api, buildHandler)
		handler.RegisterDeployRoutes(api, deployHandler)
		handler.RegisterApprovalRoutes(api, approvalHandler)
		handler.RegisterConfigRoutes(api, configHandler)
		handler.RegisterNotificationRoutes(api, notifHandler)
		handler.RegisterAuditRoutes(api, auditHandler)
		handler.RegisterUserAdminRoutes(api, userAdminHandler)
		handler.RegisterNotificationChannelRoutes(api, notifChannelHandler)
		handler.RegisterGroupRoutes(api, groupHandler)
		handler.RegisterPermissionRoutes(api, permissionHandler)
		handler.RegisterHelmValuesRoutes(api, helmValuesHandler)
		handler.RegisterNamespaceRoutes(api, namespaceHandler)
		handler.RegisterSystemSettingRoutes(api, systemSettingHandler)
		handler.RegisterNotificationRuleRoutes(api, notifRuleHandler)
		handler.RegisterConfigCenterRoutes(api, configCenterHandler)
		handler.RegisterRouteRoutes(api, routeEntryHandler, routePermHandler)
		handler.RegisterPluginRoutes(api, routePluginHandler)
		api.GET("/services/:id/env-image", envImageHandler.GetEnvImage)
	}

	// 认证路由中添加权限总览
	authGroup.GET("/my-permissions", middleware.JWTAuth(jwtSvc), permissionHandler.GetMyPermissions)

	// WebSocket 路由
	r.GET("/ws/builds/:id/log", handler.HandleBuildLogWS(wsHub, jwtSvc))
	r.GET("/ws/deployments/:id/progress", handler.HandleDeployProgressWS(wsHub, jwtSvc))
	r.GET("/ws/deployments/:id/pod-logs", handler.HandlePodLogWS(clientPool, deployRepo, jwtSvc))

	// Pod 列表（需要 JWT + 集群连接）
	api.GET("/deployments/:id/pods", handler.HandleDeployPodList(clientPool, deployRepo))

	// 启动服务
	addr := ":" + cfg.ServerPort
	log.Printf("DeployHub 服务启动，监听 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
