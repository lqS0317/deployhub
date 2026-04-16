"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useCreateDeployment } from "@/hooks/use-deployments";
import { useEnvImage } from "@/hooks/use-env-image";
import apiClient from "@/lib/api-client";
import type { Build, Service } from "@/types";

interface DeployDialogProps {
  defaultServiceId?: string;
  onClose: () => void;
}

export function DeployDialog({ defaultServiceId, onClose }: DeployDialogProps) {
  const [serviceId, setServiceId] = useState(defaultServiceId || "");
  const [clusterId, setClusterId] = useState("");
  const [namespace, setNamespace] = useState("default");
  const [buildId, setBuildId] = useState("");
  const [imageSource, setImageSource] = useState("build");
  const [externalImage, setExternalImage] = useState("");
  const [deployType, setDeployType] = useState("direct");
  // workload_type/port 从服务默认值读取，不在部署时选择
  const [helmRepoId, setHelmRepoId] = useState("");
  const [helmChartPath, setHelmChartPath] = useState("charts/general-chart");
  const [helmReleaseName, setHelmReleaseName] = useState("");
  const [helmChartBranch, setHelmChartBranch] = useState("main");
  const [helmServiceAccount, setHelmServiceAccount] = useState("");
  const [directMode, setDirectMode] = useState("config");
  const [rawYaml, setRawYaml] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  const createDeployment = useCreateDeployment();

  const { data: services } = useQuery({
    queryKey: ["services"],
    queryFn: () => apiClient.get("/services").then((r) => r.data),
  });

  const { data: gitRepos } = useQuery({
    queryKey: ["git-repos"],
    queryFn: () => apiClient.get("/git-repos").then((r) => r.data),
  });

  const { data: clustersData } = useQuery({
    queryKey: ["clusters"],
    queryFn: () => apiClient.get("/clusters").then((r) => r.data),
  });

  const { data: buildsData, isLoading: buildsLoading } = useQuery({
    queryKey: ["builds", serviceId, "succeeded"],
    queryFn: () =>
      apiClient.get("/builds", { params: { service_id: serviceId } }).then((r) => r.data),
    enabled: !!serviceId,
  });

  const successBuilds: Build[] = (buildsData?.items ?? []).filter(
    (b: Build) => b.status === "success"
  );

  const selectedService: Service | undefined = (services?.items ?? []).find(
    (s: Service) => String(s.id) === serviceId
  );
  const selectedBuild = successBuilds.find((b: Build) => String(b.id) === buildId);
  const isHelm = deployType === "helm";

  const { data: envImageData, isLoading: envImageLoading } = useEnvImage(
    selectedService?.id ?? 0,
    imageSource === "env_file" && !!selectedService?.helm_env_file_path
  );

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!serviceId) newErrors.serviceId = "请选择服务";
    if (!clusterId) newErrors.clusterId = "请选择集群";
    if (!namespace.trim()) newErrors.namespace = "请输入命名空间";
    if (imageSource === "build" && !buildId && !isHelm) newErrors.buildId = "请选择构建版本";
    if (imageSource === "external" && !externalImage.trim()) newErrors.externalImage = "请输入镜像地址";
    if (imageSource === "env_file" && !selectedService?.helm_env_file_path) newErrors.imageSource = "该服务未配置 env 文件路径";
    if (deployType === "direct" && directMode === "yaml" && !rawYaml.trim()) newErrors.rawYaml = "请输入 K8s YAML";
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    let imageTag = "";
    if (imageSource === "build" && selectedBuild) {
      imageTag = selectedBuild.image_tag || "";
    } else if (imageSource === "env_file" && envImageData) {
      imageTag = envImageData.tag || "";
    }

    createDeployment.mutate(
      {
        service_id: Number(serviceId),
        cluster_id: Number(clusterId),
        namespace,
        build_id: buildId ? Number(buildId) : undefined,
        image_tag: imageTag,
        image_source: imageSource,
        external_image: imageSource === "external" ? externalImage : undefined,
        deploy_type: deployType,
        workload_type: selectedService?.default_workload_type || "deployment",
        helm_repo_id: helmRepoId ? Number(helmRepoId) : undefined,
        helm_chart_path: helmChartPath || undefined,
        helm_release_name: helmReleaseName || undefined,
        helm_chart_branch: helmChartBranch || undefined,
        helm_service_account: helmServiceAccount || undefined,
        direct_mode: deployType === "direct" ? directMode : undefined,
        raw_yaml: deployType === "direct" && directMode === "yaml" ? rawYaml : undefined,
      },
      { onSuccess: () => onClose() }
    );
  };

  const imageSources = [
    { value: "build", label: "系统构建" },
    { value: "external", label: "外部镜像" },
    ...(isHelm && selectedService?.helm_env_file_path
      ? [{ value: "env_file", label: "从 Env 文件读取" }]
      : []),
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />

      <div className="relative z-10 w-full max-w-lg max-h-[90vh] overflow-y-auto rounded-xl bg-white p-6 shadow-2xl">
        <div className="flex items-center justify-between border-b border-gray-200 pb-4">
          <h2 className="text-lg font-semibold text-gray-900">发起部署</h2>
          <button onClick={onClose} className="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600">
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          {/* 部署方式 */}
          <div>
            <label className="mb-2 block text-sm font-medium text-gray-700">部署方式</label>
            <div className="flex gap-2">
              {(() => {
                const activeMode = deployType === "helm" ? "helm" : directMode;
                const modes = [
                  { key: "config", label: "配置部署", active: "border-blue-500 bg-blue-50 text-blue-700" },
                  { key: "yaml", label: "YAML 部署", active: "border-green-500 bg-green-50 text-green-700" },
                  { key: "helm", label: "Helm 部署", active: "border-purple-500 bg-purple-50 text-purple-700" },
                ];
                return modes.map((m) => (
                  <label key={m.key} className={`flex-1 flex items-center justify-center cursor-pointer rounded-lg border-2 px-3 py-2.5 text-sm font-medium transition-colors ${activeMode === m.key ? m.active : "border-gray-200 text-gray-600 hover:border-gray-300"}`}>
                    <input type="radio" name="deploy-mode" checked={activeMode === m.key}
                      onChange={() => {
                        if (m.key === "helm") { setDeployType("helm"); setDirectMode(""); }
                        else { setDeployType("direct"); setDirectMode(m.key); }
                        setServiceId(""); setErrors({});
                      }}
                      className="sr-only" />
                    <span>{m.label}</span>
                  </label>
                ));
              })()}
            </div>
          </div>

          {/* 选择服务 */}
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              服务 <span className="text-red-500">*</span>
            </label>
            <select
              value={serviceId}
              onChange={(e) => {
                setServiceId(e.target.value);
                setBuildId("");
                setImageSource("build");
                setExternalImage("");
                setErrors({});
              }}
              disabled={!!defaultServiceId}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-gray-50"
            >
              <option value="">请选择服务</option>
              {(services?.items ?? []).filter((s: Service) => !s.deploy_type || s.deploy_type === deployType).map((s: Service) => (
                <option key={s.id} value={s.id}>
                  {s.display_name || s.name}
                </option>
              ))}
            </select>
            {errors.serviceId && <p className="mt-1 text-xs text-red-500">{errors.serviceId}</p>}
          </div>

          {/* 目标集群 + 命名空间 */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                目标集群 <span className="text-red-500">*</span>
              </label>
              <select
                value={clusterId}
                onChange={(e) => setClusterId(e.target.value)}
                className={`w-full rounded-lg border px-3 py-2 text-sm focus:outline-none focus:ring-1 ${errors.clusterId ? "border-red-300" : "border-gray-300"} focus:border-blue-500 focus:ring-blue-500`}
              >
                <option value="">请选择</option>
                {(clustersData?.items ?? []).map((c: { id: number; name: string; display_name?: string }) => (
                  <option key={c.id} value={c.id}>{c.display_name || c.name}</option>
                ))}
              </select>
              {errors.clusterId && <p className="mt-1 text-xs text-red-500">{errors.clusterId}</p>}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                服务命名空间 <span className="text-red-500">*</span>
              </label>
              <input
                type="text" value={namespace} onChange={(e) => setNamespace(e.target.value)}
                placeholder="default"
                className={`w-full rounded-lg border px-3 py-2 text-sm focus:outline-none focus:ring-1 ${errors.namespace ? "border-red-300" : "border-gray-300"} focus:border-blue-500 focus:ring-blue-500`}
              />
              {errors.namespace && <p className="mt-1 text-xs text-red-500">{errors.namespace}</p>}
            </div>
          </div>

          {/* 镜像来源 */}
          {serviceId && (
            <div>
              <label className="mb-2 block text-sm font-medium text-gray-700">镜像来源</label>
              <div className="flex gap-3">
                {imageSources.map((src) => (
                  <label key={src.value} className="flex items-center gap-1.5 cursor-pointer">
                    <input
                      type="radio" value={src.value}
                      checked={imageSource === src.value}
                      onChange={(e) => setImageSource(e.target.value)}
                      className="text-blue-600"
                    />
                    <span className="text-sm">{src.label}</span>
                  </label>
                ))}
              </div>
              {errors.imageSource && <p className="mt-1 text-xs text-red-500">{errors.imageSource}</p>}
            </div>
          )}

          {/* build 模式 */}
          {imageSource === "build" && serviceId && (
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">构建版本</label>
              <select
                value={buildId} onChange={(e) => setBuildId(e.target.value)}
                disabled={!serviceId || buildsLoading}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-gray-50"
              >
                <option value="">{buildsLoading ? "加载中..." : "选择成功的构建（可选）"}</option>
                {successBuilds.map((b: Build) => (
                  <option key={b.id} value={b.id}>
                    {b.git_branch}:{b.git_commit?.slice(0, 7)} — {b.image_tag || `#${b.id}`}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* external 模式 */}
          {imageSource === "external" && (
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                完整镜像地址 <span className="text-red-500">*</span>
              </label>
              <input
                type="text" value={externalImage} onChange={(e) => setExternalImage(e.target.value)}
                placeholder="docker.io/myorg/myapp:v1.2.3"
                className={`w-full rounded-lg border px-3 py-2 font-mono text-sm focus:outline-none focus:ring-1 ${errors.externalImage ? "border-red-300" : "border-gray-300"} focus:border-blue-500 focus:ring-blue-500`}
              />
              {errors.externalImage && <p className="mt-1 text-xs text-red-500">{errors.externalImage}</p>}
            </div>
          )}

          {/* env_file 模式 */}
          {imageSource === "env_file" && (
            <div className="rounded-lg border border-blue-200 bg-blue-50/50 p-3">
              <h4 className="text-sm font-medium text-blue-900 mb-2">
                Env 文件: {selectedService?.helm_env_file_path}
              </h4>
              {envImageLoading ? (
                <p className="text-sm text-blue-600">读取中...</p>
              ) : envImageData ? (
                <dl className="grid grid-cols-2 gap-2 text-xs">
                  <div>
                    <dt className="text-blue-600">镜像仓库</dt>
                    <dd className="font-mono font-medium text-blue-900">{envImageData.repository}</dd>
                  </div>
                  <div>
                    <dt className="text-blue-600">标签</dt>
                    <dd className="font-mono font-medium text-blue-900">{envImageData.tag}</dd>
                  </div>
                  <div className="col-span-2">
                    <dt className="text-blue-600">完整地址</dt>
                    <dd className="font-mono font-medium text-blue-900">{envImageData.full_image}</dd>
                  </div>
                </dl>
              ) : (
                <p className="text-sm text-red-500">无法读取 env 文件镜像信息</p>
              )}
            </div>
          )}

          {/* 部署配置 */}
          {serviceId && (
            <div className="space-y-3">
              {/* Direct 模式 */}
              {deployType === "direct" && (
                <div className="space-y-3">
                  {directMode === "yaml" && (
                    <div>
                      <label className="mb-1 block text-xs text-gray-500">K8s YAML（支持多资源 --- 分隔）</label>
                      <textarea value={rawYaml} onChange={(e) => setRawYaml(e.target.value)}
                        rows={10} placeholder={"apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: my-app\n..."}
                        className="w-full rounded border border-gray-300 px-2 py-1.5 font-mono text-xs focus:border-blue-500 focus:outline-none" />
                    </div>
                  )}

                  {directMode === "config" && (
                    <ConfigPreview serviceId={serviceId} clusterId={clusterId} />
                  )}
                </div>
              )}

              {/* Helm 配置 */}
              {deployType === "helm" && (
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="mb-1 block text-xs text-gray-500">Chart Git 仓库</label>
                    <select value={helmRepoId} onChange={(e) => setHelmRepoId(e.target.value)}
                      className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm">
                      <option value="">请选择</option>
                      {(gitRepos?.items ?? []).map((r: { id: number; name: string }) => (
                        <option key={r.id} value={r.id}>{r.name}</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="mb-1 block text-xs text-gray-500">Chart 分支</label>
                    <input type="text" value={helmChartBranch} onChange={(e) => setHelmChartBranch(e.target.value)}
                      placeholder="main" className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm" />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs text-gray-500">Chart 路径</label>
                    <input type="text" value={helmChartPath} onChange={(e) => setHelmChartPath(e.target.value)}
                      placeholder="charts/general-chart" className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm" />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs text-gray-500">Release 名称</label>
                    <input type="text" value={helmReleaseName} onChange={(e) => setHelmReleaseName(e.target.value)}
                      placeholder="默认使用服务名" className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm" />
                  </div>
                  <div>
                    <label className="mb-1 block text-xs text-gray-500">ServiceAccount（可选）</label>
                    <input type="text" value={helmServiceAccount} onChange={(e) => setHelmServiceAccount(e.target.value)}
                      placeholder="覆盖集群默认 SA" className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm" />
                  </div>
                </div>
              )}
            </div>
          )}

          {/* 按钮 */}
          <div className="flex justify-end gap-3 border-t border-gray-200 pt-4">
            <button type="button" onClick={onClose}
              className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">
              取消
            </button>
            <button type="submit" disabled={createDeployment.isPending}
              className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50">
              {createDeployment.isPending ? "提交中..." : "确认部署"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

// 部署配置预览：显示该服务+集群下已发布的配置条目
function ConfigPreview({ serviceId, clusterId }: { serviceId: string; clusterId: string }) {
  const { data, isLoading } = useQuery({
    queryKey: ["config-entries-published", serviceId, clusterId],
    queryFn: () => apiClient.get(`/services/${serviceId}/config-entries/published`, { params: { cluster_id: clusterId } }).then(r => r.data),
    enabled: !!serviceId && !!clusterId,
  });

  const entries: { name: string; config_type: string; mount_path: string; version: number }[] = Array.isArray(data) ? data : [];

  const typeBadge = (t: string) => {
    const colors: Record<string, string> = { env: "bg-green-100 text-green-700", configmap: "bg-blue-100 text-blue-700", secret: "bg-orange-100 text-orange-700", serviceaccount: "bg-purple-100 text-purple-700" };
    return <span className={`rounded px-1.5 py-0.5 text-xs font-medium ${colors[t] || "bg-gray-100 text-gray-600"}`}>{t}</span>;
  };

  return (
    <div className="rounded-lg border border-gray-200 p-3 space-y-2">
      <h4 className="text-xs font-medium text-gray-700">部署时加载的配置</h4>
      {isLoading ? (
        <p className="text-xs text-gray-400">加载中...</p>
      ) : entries.length === 0 ? (
        <p className="text-xs text-gray-400">该环境暂无已发布的配置条目</p>
      ) : (
        <div className="space-y-1">
          {entries.map((e) => (
            <div key={e.name} className="flex items-center justify-between rounded bg-gray-50 px-2 py-1.5 text-xs">
              <div className="flex items-center gap-2">
                {typeBadge(e.config_type)}
                <span className="font-medium text-gray-900">{e.name}</span>
                <span className="text-gray-400">v{e.version}</span>
              </div>
              <span className="font-mono text-gray-500">
                {e.config_type === "env" ? "envFrom" : e.config_type === "serviceaccount" ? "serviceAccountName" : e.mount_path || `/etc/config/${e.name}`}
              </span>
            </div>
          ))}
        </div>
      )}
      <p className="text-xs text-gray-400">配置在配置中心管理，部署时自动注入</p>
    </div>
  );
}
