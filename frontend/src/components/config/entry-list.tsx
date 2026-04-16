"use client";

import { useConfigEntries, useDeleteEntry } from "@/hooks/use-config-center";
import { showToast } from "@/components/ui/toast";
import type { ConfigEntry } from "@/types";

interface EntryListProps {
  serviceId: number;
  clusterId: number;
  selectedEntryId: number | null;
  onSelect: (entryId: number) => void;
  onCreateClick: () => void;
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

const formatBadge = (f: string) => (
  <span className="inline-flex rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-600">
    {f}
  </span>
);

export function EntryList({
  serviceId,
  clusterId,
  selectedEntryId,
  onSelect,
  onCreateClick,
}: EntryListProps) {
  const { data: entries = [], isLoading } = useConfigEntries(serviceId, clusterId);
  const deleteEntry = useDeleteEntry();

  const handleDelete = (e: React.MouseEvent, entry: ConfigEntry) => {
    e.stopPropagation();
    if (!window.confirm(`确定删除配置条目「${entry.name}」？此操作不可恢复。`)) return;
    deleteEntry.mutate(entry.id, {
      onSuccess: () => showToast("已删除", "success"),
      onError: () => showToast("删除失败", "error"),
    });
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="inline-block h-5 w-5 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  return (
    <div>
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-sm font-medium text-gray-700">
          配置条目 <span className="text-gray-400">({entries.length})</span>
        </h3>
        <button
          onClick={onCreateClick}
          className="rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
        >
          新建配置条目
        </button>
      </div>

      {entries.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 py-12 text-center">
          <p className="text-sm text-gray-400">暂无配置条目</p>
          <button
            onClick={onCreateClick}
            className="mt-3 rounded-md bg-blue-600 px-4 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
          >
            新建配置条目
          </button>
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border border-gray-200">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  名称
                </th>
                <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  类型
                </th>
                <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                  格式
                </th>
                <th className="px-4 py-2.5 text-right text-xs font-medium uppercase tracking-wider text-gray-500">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 bg-white">
              {entries.map((entry) => (
                <tr
                  key={entry.id}
                  onClick={() => onSelect(entry.id)}
                  className={`cursor-pointer transition-colors ${
                    selectedEntryId === entry.id
                      ? "bg-blue-50"
                      : "hover:bg-gray-50"
                  }`}
                >
                  <td className="whitespace-nowrap px-4 py-2.5 text-sm font-medium text-gray-900">
                    {entry.name}
                  </td>
                  <td className="whitespace-nowrap px-4 py-2.5">
                    {typeBadge(entry.config_type)}
                  </td>
                  <td className="whitespace-nowrap px-4 py-2.5">
                    {formatBadge(entry.format)}
                  </td>
                  <td className="whitespace-nowrap px-4 py-2.5 text-right">
                    <button
                      onClick={(e) => handleDelete(e, entry)}
                      disabled={deleteEntry.isPending}
                      className="text-xs text-red-600 hover:text-red-800 disabled:opacity-50"
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
    </div>
  );
}
