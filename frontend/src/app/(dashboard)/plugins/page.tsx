"use client";

import { useState } from "react";
import {
  usePlugins,
  useDeletePlugin,
  usePluginDeployments,
} from "@/hooks/use-route-plugins";
import { CreatePluginDialog } from "@/components/plugin/create-plugin-dialog";
import { DeployPluginDialog } from "@/components/plugin/deploy-plugin-dialog";
import { showToast } from "@/components/ui/toast";
import type { RoutePlugin } from "@/types";

function DeployBadge({ pluginId }: { pluginId: number }) {
  const { data: deployments = [] } = usePluginDeployments(pluginId);
  const count = deployments.length;

  if (count === 0) {
    return (
      <span className="inline-flex rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600">
        未部署
      </span>
    );
  }

  return (
    <span className="inline-flex rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
      已部署 {count} 个环境
    </span>
  );
}

export default function PluginsPage() {
  const { data: plugins = [], isLoading } = usePlugins();
  const deletePlugin = useDeletePlugin();

  const [createOpen, setCreateOpen] = useState(false);
  const [editPlugin, setEditPlugin] = useState<RoutePlugin | null>(null);
  const [deployPlugin, setDeployPlugin] = useState<RoutePlugin | null>(null);

  const handleDelete = (plugin: RoutePlugin) => {
    if (!confirm(`确定删除插件 "${plugin.name}" 吗？`)) return;
    deletePlugin.mutate(plugin.id, {
      onSuccess: () => showToast("插件已删除", "success"),
      onError: () => showToast("删除失败", "error"),
    });
  };

  const formatTime = (t: string) => {
    if (!t) return "-";
    return new Date(t).toLocaleString("zh-CN");
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">插件中心</h1>
          <p className="mt-1 text-sm text-gray-500">
            管理路由插件（Traefik Middleware / APISIX Plugin 等）
          </p>
        </div>
        <button
          onClick={() => {
            setEditPlugin(null);
            setCreateOpen(true);
          }}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          新建插件
        </button>
      </div>

      {isLoading ? (
        <div className="py-12 text-center">
          <div className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
          <p className="mt-2 text-sm text-gray-500">加载中...</p>
        </div>
      ) : plugins.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center">
          <p className="text-gray-500">暂无插件，点击「新建插件」开始</p>
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  名称
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  描述
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  部署状态
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  更新时间
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {plugins.map((plugin) => (
                <tr key={plugin.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                    {plugin.name}
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-500 max-w-xs truncate">
                    {plugin.description || "-"}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm">
                    <DeployBadge pluginId={plugin.id} />
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-500">
                    {formatTime(plugin.updated_at)}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-right text-sm">
                    <button
                      onClick={() => {
                        setEditPlugin(plugin);
                        setCreateOpen(true);
                      }}
                      className="mr-3 text-blue-600 hover:text-blue-800"
                    >
                      编辑
                    </button>
                    <button
                      onClick={() => setDeployPlugin(plugin)}
                      className="mr-3 text-green-600 hover:text-green-800"
                    >
                      部署
                    </button>
                    <button
                      onClick={() => handleDelete(plugin)}
                      className="text-red-600 hover:text-red-800"
                    >
                      删除
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <CreatePluginDialog
        open={createOpen}
        onClose={() => {
          setCreateOpen(false);
          setEditPlugin(null);
        }}
        editPlugin={editPlugin}
      />

      <DeployPluginDialog
        open={!!deployPlugin}
        onClose={() => setDeployPlugin(null)}
        plugin={deployPlugin}
      />
    </div>
  );
}
