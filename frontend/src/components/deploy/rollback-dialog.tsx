"use client";

import { useRollbackDeployment } from "@/hooks/use-deployments";
import type { Deployment } from "@/types";

interface RollbackDialogProps {
  deployment: Deployment;
  onClose: () => void;
}

export function RollbackDialog({ deployment, onClose }: RollbackDialogProps) {
  const rollback = useRollbackDeployment();

  const handleConfirm = () => {
    rollback.mutate(
      { deploymentId: deployment.id },
      { onSuccess: () => onClose() }
    );
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />

      <div className="relative z-10 w-full max-w-md rounded-xl bg-white p-6 shadow-2xl">
        <div className="flex items-center justify-between border-b border-gray-200 pb-4">
          <h2 className="text-lg font-semibold text-gray-900">确认回滚</h2>
          <button
            onClick={onClose}
            className="rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="mt-4 space-y-4">
          {/* 警告提示 */}
          <div className="rounded-lg bg-orange-50 p-4">
            <div className="flex items-start gap-3">
              <svg className="mt-0.5 h-5 w-5 flex-shrink-0 text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z"
                />
              </svg>
              <div>
                <p className="text-sm font-medium text-orange-800">此操作将回滚到上一个部署版本</p>
                <p className="mt-1 text-xs text-orange-600">回滚会创建一次新的部署记录，将服务恢复到之前的镜像版本。</p>
              </div>
            </div>
          </div>

          {/* 当前部署信息 */}
          <div className="rounded-lg border border-gray-200 p-4">
            <h4 className="mb-3 text-sm font-medium text-gray-900">当前部署信息</h4>
            <dl className="space-y-2 text-sm">
              <div className="flex justify-between">
                <dt className="text-gray-500">服务</dt>
                <dd className="font-medium text-gray-900">{deployment.service?.name || deployment.service_id}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">当前镜像</dt>
                <dd className="font-mono font-medium text-gray-900">{deployment.image_tag || "-"}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">集群</dt>
                <dd className="font-medium text-gray-900">{deployment.cluster?.name || deployment.cluster_id || "-"}</dd>
              </div>
              <div className="flex justify-between">
                <dt className="text-gray-500">部署时间</dt>
                <dd className="font-medium text-gray-900">
                  {deployment.created_at ? new Date(deployment.created_at).toLocaleString("zh-CN") : "-"}
                </dd>
              </div>
            </dl>
          </div>

          {/* 按钮 */}
          <div className="flex justify-end gap-3 border-t border-gray-200 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50"
            >
              取消
            </button>
            <button
              onClick={handleConfirm}
              disabled={rollback.isPending}
              className="rounded-lg bg-orange-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-orange-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {rollback.isPending ? "回滚中..." : "确认回滚"}
            </button>
          </div>

          {rollback.isError && (
            <p className="text-sm text-red-500">
              回滚失败: {(rollback.error as Error)?.message || "未知错误"}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
