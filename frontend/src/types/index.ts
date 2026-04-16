// DeployHub 前端类型定义，与后端 API 响应结构对齐

// ==================== 用户 ====================

export interface User {
  id: number;
  username: string;
  email: string;
  role: "admin" | "member";
  status: "active" | "disabled";
  oauth_provider?: string;
  avatar?: string;
  nickname?: string;
  phone?: string;
  created_at: string;
  updated_at?: string;
}

// ==================== 基础设施 ====================

export interface Cluster {
  id: number;
  name: string;
  display_name?: string;
  env: "devnet" | "qanet" | "testnet" | "mainnet";
  api_server?: string;
  status: "active" | "inactive";
  helm_service_account?: string;
  created_at: string;
  updated_at?: string;
}

export interface GitRepo {
  id: number;
  name: string;
  url: string;
  provider: "github" | "gitlab" | "other";
  auth_type: "token" | "ssh";
  created_at: string;
  updated_at?: string;
}

export interface Registry {
  id: number;
  name: string;
  url: string;
  provider: string;
  is_default: boolean;
  created_at: string;
  updated_at?: string;
}

// ==================== 服务 ====================

export interface Service {
  id: number;
  name: string;
  display_name?: string;
  description?: string;
  git_repo_id: number;
  git_branch: string;
  dockerfile_path: string;
  registry_id: number;
  image_repo: string;
  cluster_id?: number;
  namespace?: string;
  replicas: number;
  cpu_request?: string;
  mem_request?: string;
  cpu_limit?: string;
  mem_limit?: string;
  port: number;
  health_check_path?: string;
  env_vars?: Record<string, string>;
  volumes?: Record<string, unknown>[];
  owner_id: number;
  service_type?: string;
  language?: string;
  language_version?: string;
  deploy_type: "direct" | "helm";
  workload_type: "deployment" | "statefulset";
  helm_repo_id?: number;
  helm_chart_path?: string;
  helm_values_path?: string;
  helm_release_name?: string;
  helm_chart_branch?: string;
  helm_env_file_path?: string;
  default_port?: number;
  default_replicas?: number;
  default_cpu_request?: string;
  default_mem_request?: string;
  default_cpu_limit?: string;
  default_mem_limit?: string;
  default_command?: string[];
  default_args?: string[];
  default_workload_type?: string;
  default_liveness_probe?: Record<string, unknown>;
  default_readiness_probe?: Record<string, unknown>;
  created_at: string;
  updated_at?: string;
  git_repo?: GitRepo;
  registry?: Registry;
  cluster?: Cluster;
  owner?: User;
  helm_repo?: GitRepo;
}

export interface ServiceMember {
  id: number;
  service_id: number;
  user_id: number;
  role: "owner" | "developer" | "viewer";
  created_at: string;
  service?: Service;
  user?: User;
}

// ==================== 构建 ====================

export type BuildStatus =
  | "pending"
  | "building"
  | "success"
  | "failed"
  | "cancelled";

export interface Build {
  id: number;
  service_id: number;
  trigger_user_id: number;
  git_branch: string;
  git_commit?: string;
  image_tag?: string;
  name?: string;
  dockerfile_path?: string;
  registry_id?: number;
  image_repo?: string;
  build_context?: string;
  status: BuildStatus;
  build_cluster_id: number;
  kaniko_job_name?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
  service?: Service;
  trigger_user?: User;
  build_cluster?: Cluster;
}

// ==================== 部署 ====================

export type DeployStatus =
  | "pending_approval"
  | "approved"
  | "previewing"
  | "previewed"
  | "deploying"
  | "pod_checking"
  | "pod_healthy"
  | "pod_unhealthy"
  | "success"
  | "failed"
  | "rolled_back"
  | "rejected"
  | "expired"
  | "cancelled";

export interface Deployment {
  id: number;
  service_id: number;
  build_id?: number;
  trigger_user_id: number;
  cluster_id: number;
  namespace: string;
  image_tag: string;
  replicas: number;
  status: DeployStatus;
  previous_image_tag?: string;
  is_rollback: boolean;
  rollback_from_id?: number;
  helm_revision?: number;
  image_source?: "build" | "external" | "env_file";
  external_image?: string;
  fail_reason?: string;
  preview_yaml?: string;
  preview_summary?: Record<string, unknown>;
  deploy_type?: "direct" | "helm";
  workload_type?: "deployment" | "statefulset";
  port?: number;
  cpu_request?: string;
  mem_request?: string;
  cpu_limit?: string;
  mem_limit?: string;
  health_check_path?: string;
  helm_repo_id?: number;
  helm_chart_path?: string;
  helm_release_name?: string;
  helm_chart_branch?: string;
  helm_service_account?: string;
  deploy_command?: string;
  pod_status?: string;
  pod_message?: string;
  started_at?: string;
  finished_at?: string;
  created_at: string;
  service?: Service;
  build?: Build;
  trigger_user?: User;
  cluster?: Cluster;
  rollback_from?: Deployment;
}

// ==================== 审批 ====================

export type ApprovalStatus = "pending" | "approved" | "rejected";

export interface Approval {
  id: number;
  deployment_id: number;
  requester_id: number;
  approver_id: number;
  status: ApprovalStatus;
  comment?: string;
  decided_at?: string;
  created_at: string;
  deployment?: Deployment;
  requester?: User;
  approver?: User;
}

// ==================== 配置管理 ====================

export interface ConfigTemplate {
  id: number;
  service_id: number;
  name: string;
  config_type: "configmap" | "secret";
  template_content: string;
  created_at: string;
  updated_at?: string;
  service?: Service;
}

export interface ConfigEnvValue {
  id: number;
  config_template_id: number;
  cluster_id: number;
  vars: Record<string, string>;
  created_at: string;
  updated_at?: string;
  config_template?: ConfigTemplate;
  cluster?: Cluster;
}

export interface ConfigVersion {
  id: number;
  config_template_id: number;
  cluster_id: number;
  version: number;
  rendered_content: string;
  created_by_id: number;
  created_at: string;
  config_template?: ConfigTemplate;
  cluster?: Cluster;
  created_by?: User;
}

export interface ConfigDeployment {
  id: number;
  config_version_id: number;
  cluster_id: number;
  namespace: string;
  resource_name: string;
  status: "pending" | "success" | "failed";
  deployed_by_id: number;
  deployed_at?: string;
  created_at: string;
  config_version?: ConfigVersion;
  cluster?: Cluster;
  deployed_by?: User;
}

// ==================== 通知 ====================

export interface Notification {
  id: number;
  user_id: number;
  type: string;
  title: string;
  content?: string;
  is_read: boolean;
  reference_type?: string;
  reference_id?: number;
  created_at: string;
  user?: User;
}

export interface NotificationChannel {
  id: number;
  name: string;
  type: "feishu" | "dingtalk" | "slack" | "generic";
  webhook_url: string;
  created_at: string;
  updated_at?: string;
}

// ==================== 审计 ====================

export interface AuditLog {
  id: number;
  user_id: number;
  action: string;
  resource_type?: string;
  resource_id?: number;
  detail?: Record<string, unknown>;
  ip_address?: string;
  created_at: string;
  user?: User;
}

// ==================== Namespace ====================

export interface ClusterNamespace {
  id: number;
  cluster_id: number;
  namespace: string;
  is_default: boolean;
  created_at: string;
}

// ==================== Helm ====================

export interface HelmValues {
  id: number;
  service_id: number;
  cluster_id: number;
  content: string;
  version: number;
  updated_by?: number;
  created_at: string;
  updated_at?: string;
  cluster?: Cluster;
}

// ==================== 组与权限 ====================

export interface Group {
  id: number;
  name: string;
  description: string;
  created_by: number;
  member_count?: number;
  permission_count?: number;
  creator?: User;
  created_at: string;
  updated_at?: string;
}

export interface GroupMember {
  id: number;
  group_id: number;
  user_id: number;
  created_at: string;
  user?: User;
}

export interface GroupServicePermission {
  id: number;
  group_id: number;
  service_id: number;
  role: "viewer" | "developer" | "owner";
  created_at: string;
  group?: Group;
  service?: Service;
}

export interface PermissionSource {
  type: "admin" | "personal" | "group";
  name: string;
  group_id?: number;
}

export interface EffectivePermission {
  service_id: number;
  service_name: string;
  role: string;
  sources: PermissionSource[];
}

// ==================== 配置中心 ====================

export interface ConfigEntry {
  id: number;
  service_id: number;
  cluster_id: number;
  name: string;
  config_type: "env" | "configmap" | "secret";
  format: "properties" | "yaml" | "json";
  mount_path?: string;
  draft_content?: string;
  created_at: string;
  updated_at: string;
}

export interface ConfigItem {
  id: number;
  config_entry_id: number;
  key: string;
  value: string;
  comment?: string;
  is_deleted: boolean;
  created_at: string;
  updated_at: string;
}

export interface ConfigRelease {
  id: number;
  config_entry_id: number;
  version: number;
  snapshot: unknown;
  status: "published" | "rolled_back";
  comment?: string;
  created_by_id: number;
  created_at: string;
  created_by?: User;
}

// ==================== 路由插件 ====================

export interface RoutePlugin {
  id: number;
  name: string;
  description: string;
  yaml_content: string;
  created_by_id: number;
  created_at: string;
  updated_at: string;
}

export interface PluginDeployment {
  id: number;
  plugin_id: number;
  cluster_id: number;
  namespace: string;
  status: string;
  yaml_snapshot: string;
  error_msg: string;
  deployed_at: string;
  cluster?: Cluster;
}

// ==================== 路由中心 ====================

export interface RouteEntry {
  id: number;
  name: string;
  resource_type: string;
  config: unknown;
  created_by_id: number;
  created_at: string;
  updated_at: string;
}

export interface RouteDeployment {
  id: number;
  route_entry_id: number;
  cluster_id: number;
  namespace: string;
  status: string;
  config_snapshot: unknown;
  rendered_yaml: string;
  error_msg: string;
  deployed_at: string;
  cluster?: Cluster;
}

// ==================== 通用 ====================

/** 分页响应包装 */
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

/** 类型别名，兼容旧代码引用 */
export type GitRepository = GitRepo;
export type ImageRegistry = Registry;

/** API 错误响应 */
export interface ApiError {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
}

/** 登录响应 */
export interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  user: Pick<User, "id" | "username" | "email">;
}

/** 注册响应 */
export interface RegisterResponse {
  user: Pick<User, "id" | "username" | "email" | "created_at">;
}
