"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTriggerBuild } from "@/hooks/use-builds";
import apiClient from "@/lib/api-client";
import type { Service } from "@/types";

interface TriggerBuildDialogProps {
  defaultServiceId?: string;
  onClose: () => void;
}

export function TriggerBuildDialog({ defaultServiceId, onClose }: TriggerBuildDialogProps) {
  const [serviceId, setServiceId] = useState(defaultServiceId || "");
  const [branch, setBranch] = useState("");
  const [commitSha, setCommitSha] = useState("");
  const [imageTag, setImageTag] = useState("");
  const [name, setName] = useState("");
  const [dockerfilePath, setDockerfilePath] = useState("Dockerfile");
  const [registryId, setRegistryId] = useState("");
  const [imageRepo, setImageRepo] = useState("");
  const [buildContext, setBuildContext] = useState(".");
  const [buildClusterId, setBuildClusterId] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});

  const triggerBuild = useTriggerBuild();

  const { data: services } = useQuery({
    queryKey: ["services"],
    queryFn: () => apiClient.get("/services").then((r) => r.data),
  });

  const { data: registries } = useQuery({
    queryKey: ["registries"],
    queryFn: () => apiClient.get("/registries").then((r) => r.data),
  });

  const { data: clustersData } = useQuery({
    queryKey: ["clusters"],
    queryFn: () => apiClient.get("/clusters").then((r) => r.data),
  });

  const selectedService: Service | undefined = (services?.items ?? []).find(
    (s: Service) => String(s.id) === serviceId
  );
  const gitRepoId = selectedService?.git_repo_id;

  const { data: branchesData, isLoading: branchesLoading } = useQuery({
    queryKey: ["git-repos", gitRepoId, "branches"],
    queryFn: () => apiClient.get(`/git-repos/${gitRepoId}/branches`).then((r) => r.data),
    enabled: !!gitRepoId,
  });

  const branches: { name: string; is_default?: boolean }[] = branchesData?.branches ?? [];

  const { data: commitsData, isLoading: commitsLoading } = useQuery({
    queryKey: ["git-repos", gitRepoId, "commits", branch],
    queryFn: () => apiClient.get(`/git-repos/${gitRepoId}/commits`, { params: { branch } }).then((r) => r.data),
    enabled: !!gitRepoId && !!branch,
  });
  const commits: { sha: string; message: string; author: string; date: string }[] = commitsData?.commits ?? [];

  // 选择服务后预填默认值
  const onServiceChange = (id: string) => {
    setServiceId(id);
    setBranch("");
    setCommitSha("");
    const svc = (services?.items ?? []).find((s: Service) => String(s.id) === id);
    if (svc) {
      setDockerfilePath(svc.dockerfile_path || "Dockerfile");
      setImageRepo(svc.image_repo || "");
      setRegistryId(svc.registry_id ? String(svc.registry_id) : "");
      setBuildClusterId(svc.cluster_id ? String(svc.cluster_id) : "");
    }
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};
    if (!serviceId) newErrors.serviceId = "请选择服务";
    if (!branch) newErrors.branch = "请选择分支";
    if (!buildClusterId) newErrors.buildClusterId = "请选择构建集群";
    if (!registryId) newErrors.registryId = "请选择镜像仓库";
    if (!imageRepo.trim()) newErrors.imageRepo = "请输入镜像路径";
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    triggerBuild.mutate(
      {
        service_id: Number(serviceId),
        git_branch: branch,
        git_commit: commitSha || undefined,
        image_tag: imageTag || undefined,
        name: name || undefined,
        dockerfile_path: dockerfilePath,
        registry_id: Number(registryId),
        image_repo: imageRepo,
        build_context: buildContext,
        build_cluster_id: Number(buildClusterId),
      },
      { onSuccess: () => onClose() }
    );
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />

      <div className="relative z-10 w-full max-w-lg max-h-[90vh] overflow-y-auto rounded-xl bg-white p-6 shadow-2xl">
        <div className="flex items-center justify-between border-b border-gray-200 pb-4">
          <h2 className="text-lg font-semibold text-gray-900">触发构建</h2>
          <button onClick={onClose} className="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600">
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          {/* 服务 */}
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              服务 <span className="text-red-500">*</span>
            </label>
            <select value={serviceId} onChange={(e) => onServiceChange(e.target.value)} disabled={!!defaultServiceId}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-gray-50">
              <option value="">请选择服务</option>
              {(services?.items ?? []).map((s: Service) => (
                <option key={s.id} value={s.id}>{s.display_name || s.name}</option>
              ))}
            </select>
            {errors.serviceId && <p className="mt-1 text-xs text-red-500">{errors.serviceId}</p>}
          </div>

          {/* 分支 */}
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              分支 <span className="text-red-500">*</span>
            </label>
            <select value={branch} onChange={(e) => setBranch(e.target.value)} disabled={!serviceId || branchesLoading}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-gray-50">
              <option value="">{branchesLoading ? "加载中..." : "请选择分支"}</option>
              {branches.map((b) => (
                <option key={b.name} value={b.name}>{b.name}{b.is_default ? " (默认)" : ""}</option>
              ))}
            </select>
            {errors.branch && <p className="mt-1 text-xs text-red-500">{errors.branch}</p>}
          </div>

          {/* Commit */}
          {branch && (
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">Commit</label>
              <select value={commitSha} onChange={(e) => setCommitSha(e.target.value)} disabled={commitsLoading}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-gray-50">
                <option value="">{commitsLoading ? "加载中..." : "最新提交（HEAD）"}</option>
                {commits.map((c) => (
                  <option key={c.sha} value={c.sha}>
                    {c.sha.slice(0, 7)} — {c.message.slice(0, 50)}{c.message.length > 50 ? "..." : ""} ({c.author})
                  </option>
                ))}
              </select>
              <p className="mt-1 text-xs text-gray-400">留空则使用该分支最新提交</p>
            </div>
          )}

          {/* 构建集群 */}
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              构建集群 <span className="text-red-500">*</span>
            </label>
            <select value={buildClusterId} onChange={(e) => setBuildClusterId(e.target.value)}
              className={`w-full rounded-lg border px-3 py-2 text-sm focus:outline-none focus:ring-1 ${errors.buildClusterId ? "border-red-300" : "border-gray-300"} focus:border-blue-500 focus:ring-blue-500`}>
              <option value="">请选择集群</option>
              {(clustersData?.items ?? []).map((c: { id: number; name: string; display_name?: string }) => (
                <option key={c.id} value={c.id}>{c.display_name || c.name}</option>
              ))}
            </select>
            {errors.buildClusterId && <p className="mt-1 text-xs text-red-500">{errors.buildClusterId}</p>}
          </div>

          {/* 构建配置 */}
          <div className="rounded-lg border border-gray-200 p-3 space-y-3">
            <h3 className="text-sm font-medium text-gray-700">构建配置</h3>

            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-xs text-gray-600">
                  镜像仓库 <span className="text-red-500">*</span>
                </label>
                <select value={registryId} onChange={(e) => setRegistryId(e.target.value)}
                  className={`w-full rounded border px-2 py-1.5 text-sm ${errors.registryId ? "border-red-300" : "border-gray-300"} focus:border-blue-500 focus:outline-none`}>
                  <option value="">请选择</option>
                  {(registries?.items ?? []).map((r: { id: number; name: string }) => (
                    <option key={r.id} value={r.id}>{r.name}</option>
                  ))}
                </select>
                {errors.registryId && <p className="mt-1 text-xs text-red-500">{errors.registryId}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs text-gray-600">
                  镜像路径 <span className="text-red-500">*</span>
                </label>
                <input type="text" value={imageRepo} onChange={(e) => setImageRepo(e.target.value)}
                  placeholder="lq0317/myapp"
                  className={`w-full rounded border px-2 py-1.5 text-sm ${errors.imageRepo ? "border-red-300" : "border-gray-300"} focus:border-blue-500 focus:outline-none`} />
                {errors.imageRepo && <p className="mt-1 text-xs text-red-500">{errors.imageRepo}</p>}
              </div>
              <div>
                <label className="mb-1 block text-xs text-gray-600">Dockerfile 路径</label>
                <input type="text" value={dockerfilePath} onChange={(e) => setDockerfilePath(e.target.value)}
                  placeholder="Dockerfile"
                  className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none" />
              </div>
              <div>
                <label className="mb-1 block text-xs text-gray-600">构建上下文</label>
                <input type="text" value={buildContext} onChange={(e) => setBuildContext(e.target.value)}
                  placeholder="."
                  className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none" />
              </div>
            </div>
          </div>

          {/* 可选项 */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">构建名称</label>
              <input type="text" value={name} onChange={(e) => setName(e.target.value)}
                placeholder="如 v1.0.0 构建"
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">镜像标签</label>
              <input type="text" value={imageTag} onChange={(e) => setImageTag(e.target.value)}
                placeholder="留空自动生成"
                className="w-full rounded-lg border border-gray-300 px-3 py-2 font-mono text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
            </div>
          </div>

          {/* 按钮 */}
          <div className="flex justify-end gap-3 border-t border-gray-200 pt-4">
            <button type="button" onClick={onClose}
              className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">
              取消
            </button>
            <button type="submit" disabled={triggerBuild.isPending}
              className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50">
              {triggerBuild.isPending ? "提交中..." : "开始构建"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
