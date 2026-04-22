"use client";

import { useState } from "react";
import {
  useRouteEntries,
  useDeleteRouteEntry,
  useEntryDeployments,
} from "@/hooks/use-route-entries";
import { CreateRouteDialog } from "@/components/route/create-route-dialog";
import { DeployRouteDialog } from "@/components/route/deploy-route-dialog";
import { showToast } from "@/components/ui/toast";
import type { RouteEntry } from "@/types";

const TABS = [
  { key: "service", label: "Service" },
  { key: "ingress", label: "Ingress" },
  { key: "ingressroute", label: "IngressRoute" },
  { key: "apisixroute", label: "ApisixRoute" },
  { key: "apisixupstream", label: "ApisixUpstream" },
] as const;

function DeployBadge({ entryId }: { entryId: number }) {
  const { data: deployments = [] } = useEntryDeployments(entryId);
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

export default function RoutesPage() {
  const [activeTab, setActiveTab] = useState<string>("service");
  const { data: entries = [], isLoading } = useRouteEntries(activeTab);
  const deleteEntry = useDeleteRouteEntry();

  const [createOpen, setCreateOpen] = useState(false);
  const [editEntry, setEditEntry] = useState<RouteEntry | null>(null);
  const [deployEntry, setDeployEntry] = useState<RouteEntry | null>(null);

  const handleDelete = (entry: RouteEntry) => {
    if (!confirm(`确定删除路由 "${entry.name}" 吗？`)) return;
    deleteEntry.mutate(entry.id, {
      onSuccess: () => showToast("路由已删除", "success"),
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
          <h1 className="text-2xl font-bold text-gray-900">路由中心</h1>
          <p className="mt-1 text-sm text-gray-500">
            管理 K8s Service、Ingress、IngressRoute、ApisixRoute
          </p>
        </div>
        <button
          onClick={() => {
            setEditEntry(null);
            setCreateOpen(true);
          }}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          新建
        </button>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-6">
          {TABS.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`whitespace-nowrap border-b-2 py-3 px-1 text-sm font-medium transition-colors ${
                activeTab === tab.key
                  ? "border-blue-600 text-blue-600"
                  : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {isLoading ? (
        <div className="py-12 text-center">
          <div className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
          <p className="mt-2 text-sm text-gray-500">加载中...</p>
        </div>
      ) : entries.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center">
          <p className="text-gray-500">
            暂无{TABS.find((t) => t.key === activeTab)?.label}路由，点击「新建」开始
          </p>
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
              {entries.map((entry: RouteEntry) => (
                <tr key={entry.id} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                    {entry.name}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm">
                    <DeployBadge entryId={entry.id} />
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-500">
                    {formatTime(entry.updated_at)}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-right text-sm">
                    <button
                      onClick={() => {
                        setEditEntry(entry);
                        setCreateOpen(true);
                      }}
                      className="mr-3 text-blue-600 hover:text-blue-800"
                    >
                      编辑
                    </button>
                    <button
                      onClick={() => setDeployEntry(entry)}
                      className="mr-3 text-green-600 hover:text-green-800"
                    >
                      部署
                    </button>
                    <button
                      onClick={() => handleDelete(entry)}
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

      <CreateRouteDialog
        key={editEntry?.id || "new"}
        open={createOpen}
        onClose={() => {
          setCreateOpen(false);
          setEditEntry(null);
        }}
        resourceType={activeTab}
        editEntry={editEntry}
      />

      <DeployRouteDialog
        open={!!deployEntry}
        onClose={() => setDeployEntry(null)}
        entry={deployEntry}
      />
    </div>
  );
}
