"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useDeployment, useExecuteDeploy, useCancelDeploy, useDeleteDeploy } from "@/hooks/use-deployments";
import { RollbackDialog } from "@/components/deploy/rollback-dialog";
import { DeployPreviewPanel } from "@/components/deploy/deploy-preview-panel";
import { DeployLogPanel } from "@/components/deploy/deploy-log-panel";
import { PodLogPanel } from "@/components/deploy/pod-log-panel";
import { showToast } from "@/components/ui/toast";

const STATUS_COLORS: Record<string, string> = {
  success: "bg-green-100 text-green-700",
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
};

const STATUS_LABELS: Record<string, string> = {
  success: "成功",
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
};

const PHASES = ["pending_approval", "approved", "previewed", "deploying", "pod_checking", "pod_healthy"] as const;
const PHASE_LABELS: Record<string, string> = {
  pending_approval: "待审批",
  approved: "已审批",
  previewed: "已预览",
  deploying: "部署中",
  pod_checking: "Pod 检查",
  pod_healthy: "完成",
};

export default function DeploymentDetailPage() {
  const params = useParams();
  const router = useRouter();
  const deploymentId = Number(params.id);

  const [showRollback, setShowRollback] = useState(false);
  const [showPreview, setShowPreview] = useState(false);

  const { data: deployment, isLoading, refetch } = useDeployment(deploymentId);
  const executeDeploy = useExecuteDeploy();
  const cancelDeploy = useCancelDeploy();
  const deleteDeploy = useDeleteDeploy();

  const getPhaseIndex = (): number => {
    if (!deployment) return 0;
    const s = deployment.status;
    if (s === "pod_healthy" || s === "success") return PHASES.length;
    if (s === "pod_unhealthy") return PHASES.indexOf("pod_checking");
    if (s === "failed" || s === "cancelled") return -1;
    if (s === "previewing") return 2;
    const idx = PHASES.indexOf(s as (typeof PHASES)[number]);
    return idx >= 0 ? idx : 0;
  };

  if (isLoading) {
    return <div className="flex items-center justify-center py-20"><div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" /></div>;
  }

  if (!deployment) {
    return <div className="py-20 text-center text-gray-500">部署记录不存在或已被删除</div>;
  }

  const currentPhase = getPhaseIndex();
  const canExecute = deployment.status === "approved";
  const canCancel = ["pending_approval", "approved", "previewed", "previewing"].includes(deployment.status);
  const showPreviewResult = ["previewed", "pending_approval", "approved"].includes(deployment.status);

  return (
    <div className="space-y-6">
      {/* 头部 */}
      <div className="flex items-center gap-4">
        <button onClick={() => router.push("/deployments")} className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600">
          <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900">部署详情</h1>
          <p className="mt-1 font-mono text-sm text-gray-500">#{deploymentId}</p>
        </div>
        <div className="flex items-center gap-2">
          {deployment.status === "previewing" && (
            <span className="rounded-lg bg-indigo-100 px-4 py-2 text-sm font-medium text-indigo-700">
              正在预览中...
            </span>
          )}
          {deployment.status === "pending_approval" && (
            <span className="rounded-lg bg-yellow-100 px-4 py-2 text-sm font-medium text-yellow-700">
              等待管理员审批
            </span>
          )}
          {showPreviewResult && (
            <button onClick={() => setShowPreview(true)}
              className="rounded-lg border border-blue-600 px-4 py-2 text-sm font-medium text-blue-600 hover:bg-blue-50">
              查看预览
            </button>
          )}
          {canExecute && (
            <button
              onClick={() => executeDeploy.mutate(deploymentId, {
                onSuccess: () => { showToast("部署已执行", "success"); refetch(); },
              })}
              disabled={executeDeploy.isPending}
              className="rounded-lg bg-green-600 px-4 py-2 text-sm font-medium text-white hover:bg-green-700 disabled:opacity-50"
            >
              确认执行
            </button>
          )}
          {canCancel && (
            <button
              onClick={() => cancelDeploy.mutate(deploymentId, {
                onSuccess: () => { showToast("已取消", "success"); refetch(); },
              })}
              className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-600 hover:bg-gray-50"
            >
              取消
            </button>
          )}
          {["success", "pod_healthy", "pod_unhealthy"].includes(deployment.status) && !deployment.is_rollback && (
            <button onClick={() => setShowRollback(true)}
              className="rounded-lg border border-orange-600 px-4 py-2 text-sm font-medium text-orange-600 hover:bg-orange-50">
              回滚
            </button>
          )}
        </div>
      </div>

      {/* 信息卡片 */}
      <div className="rounded-lg border border-gray-200 bg-white p-6">
        <div className="grid grid-cols-2 gap-6 md:grid-cols-4">
          <div>
            <p className="text-xs text-gray-500">服务</p>
            <p className="mt-1 text-sm font-medium text-gray-900">{deployment.service?.name || String(deployment.service_id)}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500">镜像</p>
            <p className="mt-1 font-mono text-sm font-medium text-gray-900">
              {deployment.external_image || deployment.image_tag || "-"}
            </p>
          </div>
          <div>
            <p className="text-xs text-gray-500">集群</p>
            <p className="mt-1 text-sm font-medium text-gray-900">{deployment.cluster?.display_name || deployment.cluster?.name || "-"}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500">状态</p>
            <span className={`mt-1 inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${STATUS_COLORS[deployment.status] || "bg-gray-100 text-gray-600"}`}>
              {STATUS_LABELS[deployment.status] || deployment.status}
            </span>
          </div>
          <div>
            <p className="text-xs text-gray-500">命名空间</p>
            <p className="mt-1 text-sm font-medium text-gray-900">{deployment.namespace || "-"}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500">部署类型</p>
            <span className={`mt-1 inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
              deployment.deploy_type === "helm" ? "bg-purple-100 text-purple-700" : "bg-blue-100 text-blue-700"
            }`}>
              {deployment.deploy_type === "helm" ? "Helm" : "Direct"}
            </span>
          </div>
          <div>
            <p className="text-xs text-gray-500">副本数</p>
            <p className="mt-1 text-sm font-medium text-gray-900">{deployment.replicas ?? "-"}</p>
          </div>
          <div>
            <p className="text-xs text-gray-500">镜像来源</p>
            <p className="mt-1 text-sm font-medium text-gray-900">
              {deployment.image_source === "external" ? "外部镜像" : deployment.image_source === "env_file" ? "Env 文件" : "系统构建"}
            </p>
          </div>
          {deployment.port && (
            <div>
              <p className="text-xs text-gray-500">端口</p>
              <p className="mt-1 text-sm font-medium text-gray-900">{deployment.port}</p>
            </div>
          )}
          {deployment.helm_chart_path && (
            <div>
              <p className="text-xs text-gray-500">Chart 路径</p>
              <p className="mt-1 text-sm font-medium text-gray-900">{deployment.helm_chart_path}</p>
            </div>
          )}
          <div>
            <p className="text-xs text-gray-500">创建时间</p>
            <p className="mt-1 text-sm font-medium text-gray-900">
              {deployment.created_at ? new Date(deployment.created_at).toLocaleString("zh-CN") : "-"}
            </p>
          </div>
          <div>
            <p className="text-xs text-gray-500">完成时间</p>
            <p className="mt-1 text-sm font-medium text-gray-900">
              {deployment.finished_at ? new Date(deployment.finished_at).toLocaleString("zh-CN") : "-"}
            </p>
          </div>
        </div>
      </div>

      {/* 执行命令 */}
      {deployment.deploy_command && (
        <div className="rounded-lg border border-gray-200 bg-white overflow-hidden">
          <div className="flex items-center justify-between border-b border-gray-200 px-4 py-3">
            <h3 className="text-sm font-semibold text-gray-900">Helm 执行命令</h3>
            <button
              onClick={() => {
                navigator.clipboard.writeText(deployment.deploy_command || "");
                showToast("已复制到剪贴板", "success");
              }}
              className="text-xs text-blue-600 hover:text-blue-800"
            >
              复制
            </button>
          </div>
          <pre className="overflow-x-auto bg-gray-900 p-4 text-xs leading-relaxed text-gray-300 max-h-80 overflow-y-auto whitespace-pre-wrap">
            {deployment.deploy_command}
          </pre>
        </div>
      )}

      {/* 失败原因 */}
      {deployment.status === "failed" && deployment.fail_reason && (
        <div className="rounded-lg border border-red-200 bg-red-50 p-4">
          <div className="flex items-start gap-3">
            <svg className="h-5 w-5 text-red-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
            </svg>
            <div>
              <h4 className="text-sm font-medium text-red-800">部署失败</h4>
              <p className="mt-1 text-sm text-red-700 whitespace-pre-wrap">{deployment.fail_reason}</p>
            </div>
          </div>
        </div>
      )}

      {/* Pod 健康状态卡片 */}
      {deployment.pod_status && (
        <div className={`rounded-lg border p-4 ${
          deployment.pod_status === "healthy" ? "border-green-200 bg-green-50" :
          deployment.pod_status === "unhealthy" ? "border-orange-200 bg-orange-50" :
          "border-blue-200 bg-blue-50"
        }`}>
          <div className="flex items-start gap-3">
            {deployment.pod_status === "healthy" && (
              <svg className="h-5 w-5 text-green-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            )}
            {deployment.pod_status === "unhealthy" && (
              <svg className="h-5 w-5 text-orange-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
            )}
            {deployment.pod_status === "checking" && (
              <div className="h-5 w-5 mt-0.5 flex-shrink-0 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
            )}
            <div className="flex-1">
              <h4 className={`text-sm font-medium ${
                deployment.pod_status === "healthy" ? "text-green-800" :
                deployment.pod_status === "unhealthy" ? "text-orange-800" :
                "text-blue-800"
              }`}>
                {deployment.pod_status === "healthy" ? "Pod 运行正常" :
                 deployment.pod_status === "unhealthy" ? "Pod 异常" :
                 "Pod 健康检查中..."}
              </h4>
              {deployment.pod_status === "checking" && (
                <p className="mt-1 text-sm text-blue-600">60 秒观察期，检测 Pod 是否稳定运行</p>
              )}
              {deployment.pod_status === "healthy" && (
                <p className="mt-1 text-sm text-green-600">所有 Pod 在观察期内稳定运行，无重启或异常</p>
              )}
              {deployment.pod_status === "unhealthy" && deployment.pod_message && (
                <pre className="mt-2 rounded bg-orange-100 p-3 text-xs text-orange-900 whitespace-pre-wrap overflow-x-auto max-h-64 overflow-y-auto">
                  {deployment.pod_message}
                </pre>
              )}
            </div>
          </div>
        </div>
      )}

      {/* 删除按钮 */}
      {!["deploying", "previewing", "pod_checking"].includes(deployment.status) && (
        <div className="flex justify-end">
          <button
            onClick={() => {
              if (window.confirm("确定删除此部署记录？删除后不可恢复。")) {
                deleteDeploy.mutate(deploymentId, {
                  onSuccess: () => router.push("/deployments"),
                });
              }
            }}
            className="rounded-lg border border-red-300 px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-50"
          >
            删除部署
          </button>
        </div>
      )}

      {/* 状态时间线 */}
      <div className="rounded-lg border border-gray-200 bg-white p-6">
        <h3 className="mb-6 text-sm font-semibold text-gray-900">状态时间线</h3>
        <div className="flex items-center justify-between">
          {PHASES.map((phase, idx) => {
            const isCompleted = idx < currentPhase;
            const isCurrent = idx === currentPhase;
            const isFailed = (deployment.status === "failed" || deployment.status === "cancelled") && isCurrent;

            return (
              <div key={phase} className="flex flex-1 items-center">
                <div className="flex flex-col items-center">
                  <div className={`flex h-8 w-8 items-center justify-center rounded-full text-xs font-medium ${
                    isFailed ? "bg-red-100 text-red-600"
                      : isCompleted ? "bg-green-100 text-green-600"
                      : isCurrent ? "bg-blue-100 text-blue-600"
                      : "bg-gray-100 text-gray-400"
                  }`}>
                    {isCompleted ? (
                      <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    ) : isFailed ? "✕" : idx + 1}
                  </div>
                  <span className="mt-2 text-xs text-gray-500">{PHASE_LABELS[phase] || phase}</span>
                </div>
                {idx < PHASES.length - 1 && (
                  <div className={`mx-2 h-0.5 flex-1 ${idx < currentPhase ? "bg-green-400" : "bg-gray-200"}`} />
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* 日志面板 */}
      {["deploying", "success", "failed", "pod_checking", "pod_healthy", "pod_unhealthy"].includes(deployment.status) && (
        <>
          <DeployLogPanel deploymentId={deploymentId} />
          <PodLogPanel deploymentId={deploymentId} />
        </>
      )}

      {showRollback && <RollbackDialog deployment={deployment} onClose={() => setShowRollback(false)} />}
      {showPreview && <DeployPreviewPanel deploymentId={deploymentId} onClose={() => setShowPreview(false)} />}
    </div>
  );
}
