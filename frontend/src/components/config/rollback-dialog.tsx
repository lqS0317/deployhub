"use client";

import { useState } from "react";
import { useConfigReleases, useRollbackConfig } from "@/hooks/use-config-center";
import { showToast } from "@/components/ui/toast";
import type { ConfigRelease } from "@/types";

interface RollbackDialogProps {
  open: boolean;
  onClose: () => void;
  entryId: number;
}

export function RollbackDialog({ open, onClose, entryId }: RollbackDialogProps) {
  const { data: releases = [] } = useConfigReleases(entryId);
  const rollback = useRollbackConfig();
  const [selectedRelease, setSelectedRelease] = useState<ConfigRelease | null>(null);
  const [comment, setComment] = useState("");
  const [showSnapshot, setShowSnapshot] = useState(false);

  const handleRollback = () => {
    if (!selectedRelease) return;
    rollback.mutate(
      {
        entryId,
        target_version: selectedRelease.version,
        comment: comment.trim(),
      },
      {
        onSuccess: () => {
          showToast("回滚成功", "success");
          setSelectedRelease(null);
          setComment("");
          onClose();
        },
        onError: () => showToast("回滚失败", "error"),
      }
    );
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative w-full max-w-2xl rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">回滚配置</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            ✕
          </button>
        </div>
        <div className="max-h-[60vh] overflow-y-auto px-6 py-4">
          {releases.length === 0 ? (
            <p className="text-center text-sm text-gray-400">暂无发布记录</p>
          ) : (
            <div className="space-y-2">
              {releases.map((r) => (
                <div
                  key={r.id}
                  onClick={() => setSelectedRelease(r)}
                  className={`cursor-pointer rounded-lg border p-3 transition-colors ${
                    selectedRelease?.id === r.id
                      ? "border-blue-400 bg-blue-50"
                      : "border-gray-200 hover:border-gray-300"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <span className="text-sm font-medium text-gray-900">v{r.version}</span>
                      <span
                        className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                          r.status === "published"
                            ? "bg-green-100 text-green-800"
                            : "bg-orange-100 text-orange-800"
                        }`}
                      >
                        {r.status === "published" ? "已发布" : "已回滚"}
                      </span>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-gray-500">
                      <span>{r.created_by?.username}</span>
                      <span>{new Date(r.created_at).toLocaleString("zh-CN")}</span>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedRelease(r);
                          setShowSnapshot(!showSnapshot || selectedRelease?.id !== r.id);
                        }}
                        className="text-blue-600 hover:text-blue-800"
                      >
                        查看快照
                      </button>
                    </div>
                  </div>
                  {r.comment && <p className="mt-1 text-xs text-gray-500">{r.comment}</p>}
                  {showSnapshot && selectedRelease?.id === r.id && (
                    <pre className="mt-2 max-h-40 overflow-auto rounded bg-gray-100 p-2 font-mono text-xs text-gray-700">
                      {typeof r.snapshot === "string"
                        ? r.snapshot
                        : JSON.stringify(r.snapshot, null, 2)}
                    </pre>
                  )}
                </div>
              ))}
            </div>
          )}
          {selectedRelease && (
            <div className="mt-4">
              <label className="mb-1 block text-sm font-medium text-gray-700">回滚备注</label>
              <textarea
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                rows={2}
                placeholder="可选：描述回滚原因"
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          )}
        </div>
        <div className="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
          <button
            onClick={onClose}
            className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
          >
            取消
          </button>
          <button
            onClick={handleRollback}
            disabled={!selectedRelease || rollback.isPending}
            className="rounded-md bg-orange-600 px-4 py-2 text-sm font-medium text-white hover:bg-orange-700 disabled:opacity-50"
          >
            {rollback.isPending
              ? "回滚中..."
              : selectedRelease
                ? `回滚到 v${selectedRelease.version}`
                : "请选择版本"}
          </button>
        </div>
      </div>
    </div>
  );
}
