"use client";

import { useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { useDeployPlugin } from "@/hooks/use-route-plugins";
import { useClusterNamespaces } from "@/hooks/use-namespaces";
import { showToast } from "@/components/ui/toast";
import type { Cluster, RoutePlugin, PluginDeployment } from "@/types";

interface DeployPluginDialogProps {
  open: boolean;
  onClose: () => void;
  plugin: RoutePlugin | null;
}

export function DeployPluginDialog({
  open,
  onClose,
  plugin,
}: DeployPluginDialogProps) {
  const [clusterId, setClusterId] = useState("");
  const [namespace, setNamespace] = useState("");
  const [result, setResult] = useState<PluginDeployment | null>(null);

  const { data: clustersRaw } = useQuery({
    queryKey: ["clusters"],
    queryFn: async () => {
      const res = await apiClient.get("/clusters");
      return res.data;
    },
    enabled: open,
  });
  const clusters: Cluster[] = Array.isArray(clustersRaw) ? clustersRaw : clustersRaw?.items ?? [];

  const selectedClusterId = Number(clusterId) || 0;
  const {
    data: namespaceItemsData,
    isLoading: namespaceLoading,
    isError: namespaceLoadError,
  } = useClusterNamespaces(selectedClusterId);
  const namespaceItems = namespaceItemsData ?? [];
  const namespaceOptions = namespaceItems.map((item) => item.namespace);
  const noNamespaceMapping =
    !!clusterId && !namespaceLoading && !namespaceLoadError && namespaceItems.length === 0;

  const deployPlugin = useDeployPlugin();

  useEffect(() => {
    if (open) {
      setResult(null);
      setNamespace("");
      setClusterId(clusters[0]?.id?.toString() || "");
    }
  }, [open, clusters]);

  useEffect(() => {
    if (!clusterId) {
      if (namespace !== "") setNamespace("");
      return;
    }
    if (namespaceLoading || namespaceLoadError) return;
    if (!namespaceOptions.includes(namespace)) {
      setNamespace(namespaceItems[0]?.namespace ?? "");
    }
  }, [clusterId, namespace, namespaceItems, namespaceLoading, namespaceLoadError, namespaceOptions]);

  const handleDeploy = () => {
    if (!plugin || !clusterId) return;
    deployPlugin.mutate(
      {
        id: plugin.id,
        cluster_id: Number(clusterId),
        namespace,
      },
      {
        onSuccess: (data) => {
          setResult(data);
          showToast("插件部署成功", "success");
        },
        onError: () => showToast("部署失败", "error"),
      }
    );
  };

  if (!open || !plugin) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative w-full max-w-lg rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">
            部署插件 - {plugin.name}
          </h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            ✕
          </button>
        </div>

        <div className="space-y-4 px-6 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                目标集群
              </label>
              <select
                value={clusterId}
                onChange={(e) => setClusterId(e.target.value)}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="">选择集群</option>
                {clusters.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.display_name || c.name} ({c.env})
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                命名空间
              </label>
              <select
                value={namespace}
                onChange={(e) => setNamespace(e.target.value)}
                disabled={!clusterId || namespaceLoading || namespaceLoadError || namespaceItems.length === 0}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-gray-50"
              >
                {!clusterId ? (
                  <option value="">请先选择集群</option>
                ) : namespaceLoading ? (
                  <option value="">加载命名空间中...</option>
                ) : namespaceLoadError ? (
                  <option value="">命名空间加载失败</option>
                ) : namespaceItems.length === 0 ? (
                  <option value="">暂无可用 namespace</option>
                ) : (
                  namespaceItems.map((ns) => (
                    <option key={ns.id} value={ns.namespace}>
                      {ns.namespace}
                      {ns.is_default ? "（默认）" : ""}
                    </option>
                  ))
                )}
              </select>
              {noNamespaceMapping && (
                <p className="mt-1 text-xs text-red-500">
                  该集群未配置 namespace 映射，请先在集群管理中配置
                </p>
              )}
            </div>
          </div>

          {result && (
            <div
              className={`rounded-md p-3 text-sm ${
                result.status === "success"
                  ? "bg-green-50 text-green-700"
                  : "bg-red-50 text-red-700"
              }`}
            >
              <p className="font-medium">
                状态: {result.status === "success" ? "部署成功" : "部署失败"}
              </p>
              {result.error_msg && (
                <p className="mt-1 text-xs">{result.error_msg}</p>
              )}
            </div>
          )}
        </div>

        <div className="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
          <button
            onClick={onClose}
            className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            {result ? "关闭" : "取消"}
          </button>
          {!result && (
            <button
              onClick={handleDeploy}
              disabled={
                !clusterId ||
                !namespace ||
                noNamespaceMapping ||
                namespaceLoading ||
                namespaceLoadError ||
                deployPlugin.isPending
              }
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {deployPlugin.isPending ? "部署中..." : "确认部署"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
