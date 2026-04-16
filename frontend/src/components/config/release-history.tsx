"use client";

import { useConfigReleases } from "@/hooks/use-config-center";
import type { ConfigRelease } from "@/types";

interface ReleaseHistoryProps {
  entryId: number;
  onRollback?: (release: ConfigRelease) => void;
  onViewSnapshot?: (release: ConfigRelease) => void;
}

const statusBadge = (status: string) => {
  const colors: Record<string, string> = {
    published: "bg-green-100 text-green-800",
    rolled_back: "bg-orange-100 text-orange-800",
  };
  const labels: Record<string, string> = {
    published: "已发布",
    rolled_back: "已回滚",
  };
  return (
    <span
      className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
        colors[status] || "bg-gray-100 text-gray-700"
      }`}
    >
      {labels[status] || status}
    </span>
  );
};

export function ReleaseHistory({ entryId, onRollback, onViewSnapshot }: ReleaseHistoryProps) {
  const { data: releases = [], isLoading } = useConfigReleases(entryId);

  if (isLoading) {
    return <div className="py-8 text-center text-sm text-gray-500">加载中...</div>;
  }

  if (releases.length === 0) {
    return <div className="py-8 text-center text-sm text-gray-400">暂无发布记录</div>;
  }

  return (
    <div className="overflow-hidden rounded-lg border border-gray-200">
      <table className="min-w-full divide-y divide-gray-200">
        <thead className="bg-gray-50">
          <tr>
            <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
              版本
            </th>
            <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
              状态
            </th>
            <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
              备注
            </th>
            <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
              发布者
            </th>
            <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
              时间
            </th>
            <th className="px-4 py-2.5 text-right text-xs font-medium uppercase tracking-wider text-gray-500">
              操作
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-100 bg-white">
          {releases.map((r) => (
            <tr key={r.id} className="hover:bg-gray-50">
              <td className="whitespace-nowrap px-4 py-2.5 text-sm font-medium text-gray-900">
                v{r.version}
              </td>
              <td className="whitespace-nowrap px-4 py-2.5">{statusBadge(r.status)}</td>
              <td className="max-w-xs truncate px-4 py-2.5 text-sm text-gray-600">
                {r.comment || "-"}
              </td>
              <td className="whitespace-nowrap px-4 py-2.5 text-sm text-gray-600">
                {r.created_by?.username || "-"}
              </td>
              <td className="whitespace-nowrap px-4 py-2.5 text-sm text-gray-500">
                {new Date(r.created_at).toLocaleString("zh-CN")}
              </td>
              <td className="whitespace-nowrap px-4 py-2.5 text-right">
                <div className="flex items-center justify-end gap-3">
                  {onViewSnapshot && (
                    <button
                      onClick={() => onViewSnapshot(r)}
                      className="text-xs text-blue-600 hover:text-blue-800"
                    >
                      查看快照
                    </button>
                  )}
                  {onRollback && r.status === "published" && (
                    <button
                      onClick={() => onRollback(r)}
                      className="text-xs text-orange-600 hover:text-orange-800"
                    >
                      回滚到此版本
                    </button>
                  )}
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
