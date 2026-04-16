"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useBuilds, useRetryBuild, useDeleteBuild } from "@/hooks/use-builds";
import { TriggerBuildDialog } from "@/components/build/trigger-build-dialog";
import { BuildLogViewer } from "@/components/build/build-log-viewer";
import apiClient from "@/lib/api-client";
import type { Build } from "@/types";

// 构建状态颜色映射
const STATUS_COLORS: Record<string, string> = {
  success: "bg-green-100 text-green-700",
  succeeded: "bg-green-100 text-green-700",
  running: "bg-blue-100 text-blue-700",
  building: "bg-blue-100 text-blue-700",
  queued: "bg-yellow-100 text-yellow-700",
  pending: "bg-yellow-100 text-yellow-700",
  failed: "bg-red-100 text-red-700",
  cancelled: "bg-gray-100 text-gray-600",
};

const STATUS_LABELS: Record<string, string> = {
  success: "成功",
  succeeded: "成功",
  running: "构建中",
  building: "构建中",
  queued: "排队中",
  pending: "等待中",
  failed: "失败",
  cancelled: "已取消",
};

export default function BuildsPage() {
  const [serviceFilter, setServiceFilter] = useState("");
  const [page, setPage] = useState(1);
  const [showTrigger, setShowTrigger] = useState(false);
  const [viewingLogId, setViewingLogId] = useState<number | null>(null);
  const pageSize = 20;

  const queryClient = useQueryClient();
  const { data, isLoading } = useBuilds({
    page,
    page_size: pageSize,
  });
  const cancelBuild = useMutation({
    mutationFn: async (buildId: string | number) => {
      await apiClient.post(`/builds/${buildId}/cancel`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["builds"] });
    },
  });

  const allBuilds: Build[] = data?.items ?? [];
  const builds = serviceFilter
    ? allBuilds.filter((b) =>
        (b.service?.name || "").toLowerCase().includes(serviceFilter.toLowerCase())
      )
    : allBuilds;
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / pageSize);

  // 取消构建
  const handleCancel = (e: React.MouseEvent, buildId: number) => {
    e.stopPropagation();
    if (window.confirm("确定要取消该构建任务吗？")) {
      cancelBuild.mutate(buildId);
    }
  };

  const retryBuild = useRetryBuild();
  const deleteBuild = useDeleteBuild();

  const isCancellable = (status: string) =>
    ["pending", "queued", "running", "building"].includes(status);

  const isRetryable = (status: string) =>
    ["failed", "cancelled"].includes(status);

  const handleRetry = (e: React.MouseEvent, buildId: number) => {
    e.stopPropagation();
    retryBuild.mutate(buildId);
  };

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">构建中心</h1>
          <p className="mt-1 text-sm text-gray-500">管理镜像构建任务和日志</p>
        </div>
        <button
          onClick={() => setShowTrigger(true)}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
        >
          触发构建
        </button>
      </div>

      {/* 过滤器 */}
      <div className="flex items-center gap-4">
        <div className="relative max-w-xs">
          <input
            type="text"
            value={serviceFilter}
            onChange={(e) => {
              setServiceFilter(e.target.value);
              setPage(1);
            }}
            placeholder="搜索服务名称..."
            className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
        </div>
      </div>

      {/* 构建列表 */}
      <div className="overflow-x-auto rounded-lg border border-gray-200 bg-white">
        <table className="w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">服务</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">分支</th>
              <th className="hidden px-4 py-3 text-left text-xs font-medium text-gray-500 lg:table-cell">提交</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">镜像标签</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">状态</th>
              <th className="hidden px-4 py-3 text-left text-xs font-medium text-gray-500 md:table-cell">触发人</th>
              <th className="hidden px-4 py-3 text-left text-xs font-medium text-gray-500 xl:table-cell">创建时间</th>
              <th className="sticky right-0 bg-gray-50 px-4 py-3 text-right text-xs font-medium text-gray-500">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {isLoading ? (
              <tr>
                <td colSpan={8} className="px-4 py-12 text-center">
                  <div className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
                  <p className="mt-2 text-sm text-gray-500">加载中...</p>
                </td>
              </tr>
            ) : builds.length === 0 ? (
              <tr>
                <td colSpan={8} className="px-4 py-12 text-center text-sm text-gray-500">
                  暂无构建记录
                </td>
              </tr>
            ) : (
              builds.map((b) => (
                <tr key={b.id} className="transition-colors hover:bg-gray-50">
                  <td className="whitespace-nowrap px-4 py-3 text-sm font-medium text-gray-900">
                    {b.service?.name || String(b.service_id)}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-sm text-gray-600">
                    {b.git_branch || "-"}
                  </td>
                  <td className="hidden whitespace-nowrap px-4 py-3 font-mono text-sm text-gray-600 lg:table-cell">
                    {b.git_commit ? b.git_commit.slice(0, 7) : "-"}
                  </td>
                  <td className="max-w-[200px] truncate px-4 py-3 font-mono text-sm text-gray-600" title={b.image_tag || ""}>
                    {b.image_tag || "-"}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3">
                    <span
                      className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        STATUS_COLORS[b.status] || "bg-gray-100 text-gray-600"
                      }`}
                    >
                      {STATUS_LABELS[b.status] || b.status}
                    </span>
                  </td>
                  <td className="hidden whitespace-nowrap px-4 py-3 text-sm text-gray-600 md:table-cell">
                    {b.trigger_user?.username || "-"}
                  </td>
                  <td className="hidden whitespace-nowrap px-4 py-3 text-sm text-gray-500 xl:table-cell">
                    {b.created_at ? new Date(b.created_at).toLocaleString("zh-CN") : "-"}
                  </td>
                  <td className="sticky right-0 whitespace-nowrap bg-white px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => setViewingLogId(b.id)}
                        className="text-sm text-blue-600 transition-colors hover:text-blue-800"
                      >
                        日志
                      </button>
                      {isCancellable(b.status) && (
                        <button
                          onClick={(e) => handleCancel(e, b.id)}
                          className="text-sm text-red-600 transition-colors hover:text-red-800"
                        >
                          取消
                        </button>
                      )}
                      {isRetryable(b.status) && (
                        <button
                          onClick={(e) => handleRetry(e, b.id)}
                          disabled={retryBuild.isPending}
                          className="text-sm text-green-600 transition-colors hover:text-green-800 disabled:opacity-50"
                        >
                          重建
                        </button>
                      )}
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          if (window.confirm("确定删除此构建记录？")) {
                            deleteBuild.mutate(b.id);
                          }
                        }}
                        className="text-sm text-red-500 transition-colors hover:text-red-700"
                      >
                        删除
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-gray-500">
            共 {total} 条记录，第 {page}/{totalPages} 页
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              上一页
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              下一页
            </button>
          </div>
        </div>
      )}

      {/* 弹窗 */}
      {showTrigger && <TriggerBuildDialog onClose={() => setShowTrigger(false)} />}
      {viewingLogId && (
        <BuildLogViewer buildId={viewingLogId} onClose={() => setViewingLogId(null)} />
      )}
    </div>
  );
}
