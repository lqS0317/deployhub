"use client";

import { useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { useDeployPlugin } from "@/hooks/use-route-plugins";
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
  const [namespace, setNamespace] = useState("default");
  const [result, setResult] = useState<PluginDeployment | null>(null);

  const { data: clusters = [] } = useQuery({
    queryKey: ["clusters"],
    queryFn: async () => {
      const res = await apiClient.get("/clusters");
      const data = res.data;
      return Array.isArray(data) ? data : (data as { items?: Cluster[] })?.items ?? [];
    },
    enabled: open,
  });

  const deployPlugin = useDeployPlugin();

  useEffect(() => {
    if (open) {
      setResult(null);
      setNamespace("default");
      setClusterId(clusters[0]?.id?.toString() || "");
    }
  }, [open, clusters]);

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
              <input
                type="text"
                value={namespace}
                onChange={(e) => setNamespace(e.target.value)}
                placeholder="default"
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
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
              disabled={!clusterId || !namespace || deployPlugin.isPending}
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
