"use client";

import { useState } from "react";
import type { Cluster } from "@/types";

interface DeployConfigDialogProps {
  open: boolean;
  onClose: () => void;
  onConfirm: (clusterId: string, namespace: string) => void;
  clusters: Cluster[];
  renderedPreview: string;
  templateName: string;
  deploying?: boolean;
}

// 配置部署确认对话框：选择集群与命名空间，预览渲染结果后确认部署
export function DeployConfigDialog({
  open,
  onClose,
  onConfirm,
  clusters,
  renderedPreview,
  templateName,
  deploying = false,
}: DeployConfigDialogProps) {
  const [clusterId, setClusterId] = useState<string>(
    clusters[0]?.id?.toString() || ""
  );
  const [namespace, setNamespace] = useState("default");

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* 遮罩层 */}
      <div
        className="absolute inset-0 bg-black/50"
        onClick={onClose}
      />
      {/* 对话框内容 */}
      <div className="relative w-full max-w-2xl rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">
            部署配置 - {templateName}
          </h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 transition-colors"
          >
            ✕
          </button>
        </div>

        <div className="px-6 py-4 space-y-4">
          {/* 集群选择 */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                目标集群
              </label>
              <select
                value={clusterId}
                onChange={(e) => setClusterId(e.target.value)}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                {clusters.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.display_name || c.name} ({c.env})
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
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

          {/* 渲染预览 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              渲染预览
            </label>
            <pre className="max-h-[300px] overflow-auto rounded-lg bg-gray-900 p-4 text-sm text-green-400 font-mono">
              {renderedPreview || "暂无预览内容"}
            </pre>
          </div>
        </div>

        <div className="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
          <button
            onClick={onClose}
            className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
          >
            取消
          </button>
          <button
            onClick={() => onConfirm(clusterId, namespace)}
            disabled={deploying || !clusterId || !namespace}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            {deploying ? "部署中..." : "确认部署"}
          </button>
        </div>
      </div>
    </div>
  );
}
