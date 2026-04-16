"use client";

import { useState, useMemo } from "react";
import {
  useConfigItems,
  useConfigDraft,
  useConfigReleases,
  useSaveDraft,
  useConfigEntries,
} from "@/hooks/use-config-center";
import { showToast } from "@/components/ui/toast";
import { KvTableEditor } from "./kv-table-editor";
import { CodeEditor } from "./code-editor";
import { ReleaseHistory } from "./release-history";
import { PublishDialog } from "./publish-dialog";
import { RollbackDialog } from "./rollback-dialog";
import type { ConfigItem, ConfigEntry, ConfigRelease } from "@/types";

interface EntryDetailProps {
  entryId: number;
  serviceId: number;
  clusterId: number;
  onBack: () => void;
}

const typeBadge = (t: string) => {
  const map: Record<string, { bg: string; label: string }> = {
    env: { bg: "bg-blue-100 text-blue-800", label: "Env" },
    configmap: { bg: "bg-green-100 text-green-800", label: "ConfigMap" },
    secret: { bg: "bg-orange-100 text-orange-800", label: "Secret" },
  };
  const info = map[t] || { bg: "bg-gray-100 text-gray-700", label: t };
  return (
    <span className={`inline-flex rounded px-1.5 py-0.5 text-[10px] font-medium ${info.bg}`}>
      {info.label}
    </span>
  );
};

type Tab = "items" | "changes" | "releases";

export function EntryDetail({ entryId, serviceId, clusterId, onBack }: EntryDetailProps) {
  const [activeTab, setActiveTab] = useState<Tab>("items");
  const [showPublish, setShowPublish] = useState(false);
  const [showRollback, setShowRollback] = useState(false);

  const { data: entries = [] } = useConfigEntries(serviceId, clusterId);
  const entry: ConfigEntry | undefined = entries.find((e) => e.id === entryId);

  const { data: items = [] } = useConfigItems(entryId);
  const { data: draftContent = "" } = useConfigDraft(entryId);
  const { data: releases = [] } = useConfigReleases(entryId);
  const saveDraft = useSaveDraft();
  const [localDraft, setLocalDraft] = useState("");

  const lastPublished = releases.find((r: ConfigRelease) => r.status === "published");

  const pendingChanges = useMemo(() => {
    if (!entry) return [];
    if (entry.format === "properties") {
      return computeKVChanges(items, lastPublished?.snapshot);
    }
    const currentDraftVal = localDraft || draftContent;
    const hasTextChange = computeTextChanged(currentDraftVal, lastPublished?.snapshot);
    return hasTextChange
      ? [{ key: "content", type: "modified" as const, newValue: "(草稿已修改)" }]
      : [];
  }, [items, lastPublished, entry, localDraft, draftContent]);

  const draftInitialized = localDraft || draftContent;
  const currentDraft = localDraft || draftContent;

  const handleSaveDraft = () => {
    saveDraft.mutate(
      { entryId, content: currentDraft },
      {
        onSuccess: () => showToast("草稿已保存", "success"),
        onError: () => showToast("保存失败", "error"),
      }
    );
  };

  if (!entry) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <div className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  const isProperties = entry.format === "properties";
  const isSecret = entry.config_type === "secret";

  const tabs: { key: Tab; label: string }[] = [
    { key: "items", label: "配置项" },
    { key: "changes", label: "更改历史" },
    { key: "releases", label: "发布历史" },
  ];

  const handleViewSnapshot = (release: ConfigRelease) => {
    const snapshot =
      typeof release.snapshot === "string"
        ? release.snapshot
        : JSON.stringify(release.snapshot, null, 2);
    window.alert(`v${release.version} 快照:\n\n${snapshot}`);
  };

  return (
    <div className="flex-1 overflow-y-auto">
      {/* 顶部操作栏 */}
      <div className="flex items-center justify-between border-b border-gray-200 px-6 py-3">
        <div className="flex items-center gap-3">
          <button
            onClick={onBack}
            className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
          >
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h2 className="flex items-center gap-2 text-lg font-semibold text-gray-900">
              {entry.name}
              {typeBadge(entry.config_type)}
              <span className="inline-flex rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-600">
                {entry.format}
              </span>
            </h2>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowPublish(true)}
            className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700"
          >
            发布
            {pendingChanges.length > 0 && (
              <span className="ml-1 rounded-full bg-white/20 px-1.5 text-xs">
                {pendingChanges.length}
              </span>
            )}
          </button>
          <button
            onClick={() => setShowRollback(true)}
            className="rounded-md border border-orange-300 bg-white px-3 py-1.5 text-sm font-medium text-orange-600 hover:bg-orange-50"
          >
            回滚
          </button>
        </div>
      </div>

      {/* Tab 导航 */}
      <div className="flex border-b border-gray-200 px-6">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`border-b-2 px-4 py-2.5 text-sm font-medium transition-colors ${
              activeTab === tab.key
                ? "border-blue-600 text-blue-600"
                : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab 内容 */}
      <div className="p-6">
        {activeTab === "items" && (
          <>
            {isProperties ? (
              <KvTableEditor
                items={items}
                entryId={entryId}
                isSecret={isSecret}
                canEdit
              />
            ) : (
              <CodeEditor
                content={draftInitialized ? currentDraft : ""}
                onChange={setLocalDraft}
                format={entry.format as "yaml" | "json"}
                canEdit
                saving={saveDraft.isPending}
                onSave={handleSaveDraft}
              />
            )}
          </>
        )}

        {activeTab === "changes" && (
          <div className="space-y-3">
            {pendingChanges.length === 0 ? (
              <div className="rounded-lg border border-gray-200 bg-gray-50 p-6 text-center text-sm text-gray-500">
                当前无待发布变更
              </div>
            ) : (
              <>
                <p className="text-sm text-gray-700">
                  与上次发布版本对比，共{" "}
                  <strong className="text-blue-600">{pendingChanges.length}</strong> 项变更：
                </p>
                <div className="overflow-hidden rounded-lg border border-gray-200">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-gray-200 bg-gray-50">
                        <th className="px-4 py-2 text-left text-xs text-gray-500">类型</th>
                        <th className="px-4 py-2 text-left text-xs text-gray-500">Key</th>
                        <th className="px-4 py-2 text-left text-xs text-gray-500">详情</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                      {pendingChanges.map((d, i) => (
                        <tr key={i} className="hover:bg-gray-50">
                          <td className="px-4 py-2">
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
                          <td className="px-4 py-2 font-mono text-gray-900">{d.key}</td>
                          <td className="max-w-[250px] truncate px-4 py-2 text-xs text-gray-600">
                            {d.type === "modified" && d.oldValue ? (
                              <span>
                                <s className="text-red-400">{d.oldValue}</s> → {d.newValue}
                              </span>
                            ) : (
                              d.newValue || d.oldValue || ""
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </>
            )}
          </div>
        )}

        {activeTab === "releases" && (
          <ReleaseHistory
            entryId={entryId}
            onRollback={() => setShowRollback(true)}
            onViewSnapshot={handleViewSnapshot}
          />
        )}
      </div>

      {/* 弹窗 */}
      <PublishDialog
        open={showPublish}
        onClose={() => setShowPublish(false)}
        entryId={entryId}
        format={entry.format}
      />
      <RollbackDialog
        open={showRollback}
        onClose={() => setShowRollback(false)}
        entryId={entryId}
      />
    </div>
  );
}

interface ChangeEntry {
  key: string;
  type: "added" | "modified" | "deleted";
  oldValue?: string;
  newValue?: string;
}

function computeKVChanges(currentItems: ConfigItem[], snapshot: unknown): ChangeEntry[] {
  const changes: ChangeEntry[] = [];
  const oldMap = new Map<string, string>();

  if (snapshot != null) {
    try {
      const parsed = typeof snapshot === "string" ? JSON.parse(snapshot) : snapshot;
      if (Array.isArray(parsed)) {
        for (const item of parsed) {
          if (item?.key) oldMap.set(item.key, String(item.value ?? ""));
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
    if (!oldMap.has(key)) changes.push({ key, type: "added", newValue: val });
    else if (oldMap.get(key) !== val)
      changes.push({ key, type: "modified", oldValue: oldMap.get(key), newValue: val });
  });
  Array.from(oldMap.entries()).forEach(([key, val]) => {
    if (!currentMap.has(key)) changes.push({ key, type: "deleted", oldValue: val });
  });
  return changes;
}

function computeTextChanged(currentDraft: string, snapshot: unknown): boolean {
  let old = "";
  if (snapshot != null) {
    try {
      const p = typeof snapshot === "string" ? JSON.parse(snapshot) : snapshot;
      old = typeof p === "string" ? p : JSON.stringify(p);
    } catch {
      /* ignore */
    }
  }
  return currentDraft.trim() !== old.trim();
}
