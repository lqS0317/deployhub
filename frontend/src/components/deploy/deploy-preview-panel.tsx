"use client";

import { useState } from "react";
import { useDeployPreview, useExecuteDeploy, useCancelDeploy } from "@/hooks/use-deployments";
import { showToast } from "@/components/ui/toast";

interface Props {
  deploymentId: number;
  onClose: () => void;
}

export function DeployPreviewPanel({ deploymentId, onClose }: Props) {
  const { data, isLoading } = useDeployPreview(deploymentId);
  const executeDeploy = useExecuteDeploy();
  const cancelDeploy = useCancelDeploy();
  const [showYaml, setShowYaml] = useState(false);

  const handleExecute = () => {
    if (window.confirm("确认执行此部署？")) {
      executeDeploy.mutate(deploymentId, {
        onSuccess: () => {
          showToast("部署已触发执行", "success");
          onClose();
        },
      });
    }
  };

  const handleCancel = () => {
    cancelDeploy.mutate(deploymentId, {
      onSuccess: () => {
        showToast("部署已取消", "success");
        onClose();
      },
    });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative z-10 w-full max-w-4xl max-h-[85vh] overflow-y-auto rounded-xl bg-white shadow-2xl">
        <div className="sticky top-0 flex items-center justify-between border-b border-gray-200 bg-white px-6 py-4">
          <h2 className="text-lg font-semibold">部署预览</h2>
          <button onClick={onClose} className="rounded p-1 text-gray-400 hover:bg-gray-100">
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div className="p-6 space-y-4">
          {isLoading ? (
            <p className="text-center text-gray-500 py-8">加载预览中...</p>
          ) : !data ? (
            <p className="text-center text-gray-500 py-8">暂无预览结果</p>
          ) : (
            <>
              {/* 状态 */}
              <div className="flex items-center gap-2">
                <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                  data.status === "previewed" ? "bg-blue-100 text-blue-700" :
                  data.status === "failed" ? "bg-red-100 text-red-700" :
                  "bg-yellow-100 text-yellow-700"
                }`}>
                  {data.status === "previewed" ? "预览完成" : data.status === "failed" ? "预览失败" : data.status}
                </span>
              </div>

              {/* 执行命令 */}
              {data.deploy_command && (
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <h4 className="text-sm font-medium text-gray-900">执行命令</h4>
                    <button
                      onClick={() => {
                        navigator.clipboard.writeText(data.deploy_command || "");
                        showToast("已复制", "success");
                      }}
                      className="text-xs text-blue-600 hover:text-blue-800"
                    >
                      复制
                    </button>
                  </div>
                  <pre className="max-h-[200px] overflow-auto rounded-lg bg-gray-900 p-4 text-xs text-green-300 font-mono whitespace-pre-wrap">
                    {data.deploy_command}
                  </pre>
                </div>
              )}

              {/* YAML 展开/收起 */}
              {data.preview_yaml && (
                <div>
                  <button
                    onClick={() => setShowYaml(!showYaml)}
                    className="text-sm text-blue-600 hover:text-blue-800"
                  >
                    {showYaml ? "收起渲染 YAML" : "查看渲染 YAML"}
                  </button>
                  {showYaml && (
                    <pre className="mt-2 max-h-[400px] overflow-auto rounded-lg bg-gray-900 p-4 text-xs text-gray-300 font-mono">
                      {data.preview_yaml}
                    </pre>
                  )}
                </div>
              )}

              {/* 操作按钮 */}
              {data.status === "previewed" && (
                <div className="flex justify-end gap-3 border-t border-gray-200 pt-4">
                  <button
                    onClick={handleCancel}
                    disabled={cancelDeploy.isPending}
                    className="rounded-lg border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                  >
                    取消部署
                  </button>
                  <button
                    onClick={handleExecute}
                    disabled={executeDeploy.isPending}
                    className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
                  >
                    {executeDeploy.isPending ? "执行中..." : "确认执行"}
                  </button>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
