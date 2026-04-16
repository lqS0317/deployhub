"use client";

import { useState, useMemo } from "react";
import { usePublish, useConfigItems, useConfigDraft, useConfigReleases } from "@/hooks/use-config-center";
import { showToast } from "@/components/ui/toast";
import type { ConfigItem, ConfigRelease } from "@/types";

interface PublishDialogProps {
  open: boolean;
  onClose: () => void;
  entryId: number;
  format: string;
}

interface DiffEntry {
  key: string;
  type: "added" | "modified" | "deleted";
  oldValue?: string;
  newValue?: string;
}

export function PublishDialog({ open, onClose, entryId, format }: PublishDialogProps) {
  const [comment, setComment] = useState("");
  const publish = usePublish();

  const { data: items = [], isLoading: itemsLoading } = useConfigItems(entryId);
  const { data: draft = "" } = useConfigDraft(entryId);
  const { data: releases = [], isLoading: releasesLoading } = useConfigReleases(entryId);

  const lastPublished = releases.find((r: ConfigRelease) => r.status === "published");
  const isLoading = itemsLoading || releasesLoading;

  const diff = useMemo(() => {
    if (format === "properties") {
      return computeKVDiff(items, lastPublished?.snapshot);
    }
    return computeTextDiff(draft, lastPublished?.snapshot);
  }, [items, draft, lastPublished, format]);

  const handlePublish = () => {
    publish.mutate(
      { entryId, comment: comment.trim() },
      {
        onSuccess: () => {
          showToast("发布成功", "success");
          setComment("");
          onClose();
        },
        onError: () => showToast("发布失败", "error"),
      }
    );
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative max-h-[80vh] w-full max-w-lg overflow-y-auto rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">发布配置</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            ✕
          </button>
        </div>

        <div className="space-y-4 px-6 py-4">
          {isLoading && (
            <p className="py-4 text-center text-sm text-gray-500">加载变更内容...</p>
          )}

          {!isLoading &&
            (format === "properties" ? (
              <div>
                <h4 className="mb-2 text-sm font-medium text-gray-700">
                  待发布变更 <span className="text-blue-600">({diff.length})</span>
                </h4>
                {diff.length === 0 ? (
                  <p className="text-sm text-gray-500">无变更</p>
                ) : (
                  <div className="max-h-60 overflow-y-auto rounded border border-gray-200">
                    <table className="w-full text-xs">
                      <thead>
                        <tr className="border-b bg-gray-50">
                          <th className="px-3 py-1.5 text-left text-gray-500">类型</th>
                          <th className="px-3 py-1.5 text-left text-gray-500">Key</th>
                          <th className="px-3 py-1.5 text-left text-gray-500">变更</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-100">
                        {(diff as DiffEntry[]).map((d, i) => (
                          <tr key={i} className="hover:bg-gray-50">
                            <td className="px-3 py-1.5">
                              <span
                                className={`rounded px-1.5 py-0.5 text-xs font-medium ${
                                  d.type === "added"
                                    ? "bg-green-100 text-green-700"
                                    : d.type === "deleted"
                                      ? "bg-red-100 text-red-700"
                                      : "bg-yellow-100 text-yellow-700"
                                }`}
                              >
                                {d.type === "added" ? "新增" : d.type === "deleted" ? "删除" : "修改"}
                              </span>
                            </td>
                            <td className="px-3 py-1.5 font-mono text-gray-900">{d.key}</td>
                            <td className="max-w-[200px] truncate px-3 py-1.5 text-gray-600">
                              {d.type === "modified" ? (
                                <span>
                                  <s className="text-red-400">{d.oldValue}</s> → {d.newValue}
                                </span>
                              ) : d.type === "added" ? (
                                <span className="text-green-600">{d.newValue}</span>
                              ) : (
                                <s className="text-red-400">{d.oldValue}</s>
                              )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            ) : (
              <div>
                <h4 className="mb-2 text-sm font-medium text-gray-700">待发布内容预览</h4>
                <pre className="max-h-60 overflow-auto whitespace-pre-wrap rounded bg-gray-900 p-3 font-mono text-xs text-gray-300">
                  {draft || "(空)"}
                </pre>
              </div>
            ))}

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">发布备注</label>
            <textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              rows={2}
              placeholder="描述本次发布的变更内容..."
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
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
            onClick={handlePublish}
            disabled={publish.isPending}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {publish.isPending ? "发布中..." : "确认发布"}
          </button>
        </div>
      </div>
    </div>
  );
}

function computeKVDiff(currentItems: ConfigItem[], snapshot: unknown): DiffEntry[] {
  const diffs: DiffEntry[] = [];
  const oldMap = new Map<string, string>();
  if (snapshot != null) {
    try {
      let parsed = snapshot;
      if (typeof parsed === "string") parsed = JSON.parse(parsed);
      if (Array.isArray(parsed)) {
        for (const item of parsed) {
          if (item && typeof item.key === "string") {
            oldMap.set(item.key, String(item.value ?? ""));
          }
        }
      }
    } catch {
      /* ignore */
    }
  }

  const activeItems = currentItems.filter((i) => !i.is_deleted);
  const currentMap = new Map<string, string>();
  for (const item of activeItems) currentMap.set(item.key, item.value ?? "");

  Array.from(currentMap.entries()).forEach(([key, val]) => {
    if (!oldMap.has(key)) diffs.push({ key, type: "added", newValue: val });
    else if (oldMap.get(key) !== val)
      diffs.push({ key, type: "modified", oldValue: oldMap.get(key), newValue: val });
  });
  Array.from(oldMap.entries()).forEach(([key, val]) => {
    if (!currentMap.has(key)) diffs.push({ key, type: "deleted", oldValue: val });
  });
  return diffs;
}

function computeTextDiff(currentDraft: string, snapshot: unknown): DiffEntry[] {
  let oldContent = "";
  if (snapshot) {
    try {
      const parsed = typeof snapshot === "string" ? JSON.parse(snapshot) : snapshot;
      oldContent = typeof parsed === "string" ? parsed : JSON.stringify(parsed);
    } catch {
      /* ignore */
    }
  }
  if (currentDraft.trim() !== oldContent.trim()) {
    return [{ key: "content", type: "modified", oldValue: "(上次发布)", newValue: "(当前草稿)" }];
  }
  return [];
}
