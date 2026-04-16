"use client";

import { useState, useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import {
  useDeployRouteEntry,
  usePreviewRouteEntry,
} from "@/hooks/use-route-entries";
import { YamlPreview } from "./yaml-preview";
import { showToast } from "@/components/ui/toast";
import type { Cluster, RouteEntry } from "@/types";

interface DeployRouteDialogProps {
  open: boolean;
  onClose: () => void;
  entry: RouteEntry | null;
}

export function DeployRouteDialog({
  open,
  onClose,
  entry,
}: DeployRouteDialogProps) {
  const [clusterId, setClusterId] = useState("");
  const [namespace, setNamespace] = useState("default");

  const { data: clusters = [] } = useQuery({
    queryKey: ["clusters"],
    queryFn: async () => {
      const res = await apiClient.get("/clusters");
      const data = res.data;
      return Array.isArray(data) ? data : (data as { items?: Cluster[] })?.items ?? [];
    },
    enabled: open,
  });

  const {
    data: preview,
    isLoading: previewLoading,
  } = usePreviewRouteEntry(entry?.id ?? 0, namespace);

  const deployEntry = useDeployRouteEntry();

  useEffect(() => {
    if (open) {
      setNamespace("default");
      setClusterId(clusters[0]?.id?.toString() || "");
    }
  }, [open, clusters]);

  const handleDeploy = () => {
    if (!entry || !clusterId) return;
    deployEntry.mutate(
      {
        id: entry.id,
        cluster_id: Number(clusterId),
        namespace,
      },
      {
        onSuccess: () => {
          showToast("路由部署成功", "success");
          onClose();
        },
        onError: () => showToast("部署失败", "error"),
      }
    );
  };

  if (!open || !entry) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative w-full max-w-2xl rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">
            部署路由 - {entry.name}
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

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              YAML 预览
            </label>
            <YamlPreview
              yaml={preview?.yaml || ""}
              loading={previewLoading}
            />
          </div>
        </div>

        <div className="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
          <button
            onClick={onClose}
            className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            取消
          </button>
          <button
            onClick={handleDeploy}
            disabled={!clusterId || !namespace || deployEntry.isPending}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {deployEntry.isPending ? "部署中..." : "确认部署"}
          </button>
        </div>
      </div>
    </div>
  );
}
