"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useDeployments, useExecuteDeploy, useCancelDeploy, useDeleteDeploy } from "@/hooks/use-deployments";
import { DeployDialog } from "@/components/deploy/deploy-dialog";
import { RollbackDialog } from "@/components/deploy/rollback-dialog";
import { DeployPreviewPanel } from "@/components/deploy/deploy-preview-panel";
import { showToast } from "@/components/ui/toast";
import type { Deployment } from "@/types";

const STATUS_COLORS: Record<string, string> = {
  success: "bg-green-100 text-green-700",
  succeeded: "bg-green-100 text-green-700",
  deploying: "bg-blue-100 text-blue-700",
  approved: "bg-blue-100 text-blue-700",
  pending_approval: "bg-yellow-100 text-yellow-700",
  previewing: "bg-indigo-100 text-indigo-700",
  previewed: "bg-cyan-100 text-cyan-700",
  pod_checking: "bg-blue-100 text-blue-700",
  pod_healthy: "bg-green-100 text-green-700",
  pod_unhealthy: "bg-orange-100 text-orange-700",
  cancelled: "bg-gray-100 text-gray-500",
  failed: "bg-red-100 text-red-700",
  rejected: "bg-red-100 text-red-700",
  rolled_back: "bg-orange-100 text-orange-700",
  expired: "bg-gray-100 text-gray-600",
};

const STATUS_LABELS: Record<string, string> = {
  success: "成功",
  succeeded: "成功",
  deploying: "部署中",
  approved: "已审批",
  pending_approval: "待审批",
  previewing: "预览中",
  previewed: "待确认",
  pod_checking: "Pod 检查中",
  pod_healthy: "运行正常",
  pod_unhealthy: "Pod 异常",
  cancelled: "已取消",
  failed: "失败",
  rejected: "已拒绝",
  rolled_back: "已回滚",
  expired: "已过期",
};

export default function DeploymentsPage() {
  const router = useRouter();
  const [serviceFilter, setServiceFilter] = useState("");
  const [page, setPage] = useState(1);
  const [showDeploy, setShowDeploy] = useState(false);
  const [rollbackTarget, setRollbackTarget] = useState<Deployment | null>(null);
  const [previewingId, setPreviewingId] = useState<number | null>(null);
  const executeDeploy = useExecuteDeploy();
  const cancelDeploy = useCancelDeploy();
  const deleteDeploy = useDeleteDeploy();
  const pageSize = 20;

  const { data, isLoading } = useDeployments({
    service_id: serviceFilter || undefined,
    page,
    page_size: pageSize,
  });

  const deployments: Deployment[] = data?.items ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / pageSize);

  return (
    <div className="space-y-6">
      {/* 标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">发布管理</h1>
          <p className="mt-1 text-sm text-gray-500">管理服务发布和回滚操作</p>
        </div>
        <button
          onClick={() => setShowDeploy(true)}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
        >
          发起部署
        </button>
      </div>

      {/* 过滤器 */}
      <div className="flex items-center gap-4">
        <input
          type="text"
          value={serviceFilter}
          onChange={(e) => {
            setServiceFilter(e.target.value);
            setPage(1);
          }}
          placeholder="按服务 ID 过滤..."
          className="max-w-xs rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </div>

      {/* 部署列表 */}
      <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">服务</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">镜像版本</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">集群</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">副本数</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">状态</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">回滚</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">触发人</th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">创建时间</th>
              <th className="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {isLoading ? (
              <tr>
                <td colSpan={9} className="px-6 py-12 text-center">
                  <div className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
                  <p className="mt-2 text-sm text-gray-500">加载中...</p>
                </td>
              </tr>
            ) : deployments.length === 0 ? (
              <tr>
                <td colSpan={9} className="px-6 py-12 text-center text-sm text-gray-500">
                  暂无部署记录
                </td>
              </tr>
            ) : (
              deployments.map((d) => (
                <tr
                  key={d.id}
                  onClick={() => router.push(`/deployments/${d.id}`)}
                  className="cursor-pointer transition-colors hover:bg-gray-50"
                >
                  <td className="whitespace-nowrap px-6 py-4 text-sm font-medium text-gray-900">
                    {d.service?.name || String(d.service_id)}
                  </td>
                  <td className="max-w-[300px] truncate whitespace-nowrap px-6 py-4 font-mono text-sm text-gray-600" title={d.external_image || d.image_tag || ""}>
                    {d.external_image || d.image_tag || "-"}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {d.cluster?.display_name || d.cluster?.name || "-"}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {d.replicas ?? "-"}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4">
                    <span
                      className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        STATUS_COLORS[d.status] || "bg-gray-100 text-gray-600"
                      } ${d.status === "pod_checking" ? "animate-pulse" : ""}`}
                      title={d.status === "failed" && d.fail_reason ? d.fail_reason : d.status === "pod_unhealthy" && d.pod_message ? d.pod_message : undefined}
                    >
                      {STATUS_LABELS[d.status] || d.status}
                    </span>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {d.is_rollback ? (
                      <span className="inline-flex rounded-full bg-orange-50 px-2 py-0.5 text-xs font-medium text-orange-600">
                        是
                      </span>
                    ) : (
                      <span className="text-gray-400">否</span>
                    )}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {d.trigger_user?.username || "-"}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-500">
                    {d.created_at ? new Date(d.created_at).toLocaleString("zh-CN") : "-"}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-2">
                      {d.status === "previewing" && (
                        <span className="text-xs text-indigo-600">预览中...</span>
                      )}
                      {d.status === "previewed" && (
                        <button
                          onClick={(e) => { e.stopPropagation(); setPreviewingId(d.id); }}
                          className="text-sm text-blue-600 hover:text-blue-800"
                        >
                          查看预览
                        </button>
                      )}
                      {d.status === "approved" && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            executeDeploy.mutate(d.id, {
                              onSuccess: () => showToast("部署已执行", "success"),
                            });
                          }}
                          className="text-sm text-green-600 hover:text-green-800"
                        >
                          执行
                        </button>
                      )}
                      {d.status === "pending_approval" && (
                        <span className="text-xs text-yellow-600">等待审批</span>
                      )}
                      {["pending_approval", "approved", "previewed", "previewing"].includes(d.status) && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            cancelDeploy.mutate(d.id, {
                              onSuccess: () => showToast("已取消", "success"),
                            });
                          }}
                          className="text-sm text-gray-500 hover:text-gray-700"
                        >
                          取消
                        </button>
                      )}
                      {["success", "pod_healthy", "pod_unhealthy"].includes(d.status) && (
                        <button
                          onClick={(e) => { e.stopPropagation(); setRollbackTarget(d); }}
                          className="text-sm text-orange-600 hover:text-orange-800"
                        >
                          回滚
                        </button>
                      )}
                      {!["deploying", "previewing", "pod_checking"].includes(d.status) && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            if (window.confirm("确定删除此部署记录？")) {
                              deleteDeploy.mutate(d.id);
                            }
                          }}
                          className="text-sm text-red-500 hover:text-red-700"
                        >
                          删除
                        </button>
                      )}
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
      {showDeploy && <DeployDialog onClose={() => setShowDeploy(false)} />}
      {rollbackTarget && (
        <RollbackDialog
          deployment={rollbackTarget}
          onClose={() => setRollbackTarget(null)}
        />
      )}
      {previewingId && (
        <DeployPreviewPanel
          deploymentId={previewingId}
          onClose={() => setPreviewingId(null)}
        />
      )}
    </div>
  );
}
