"use client";

import { useState } from "react";
import { useGroups, useDeleteGroup } from "@/hooks/use-groups";
import { GroupFormDialog } from "@/components/group/group-form-dialog";
import { GroupDetailPanel } from "@/components/group/group-detail-panel";
import type { Group } from "@/types";

export default function GroupsPage() {
  const { data, isLoading } = useGroups();
  const deleteGroup = useDeleteGroup();
  const [showCreate, setShowCreate] = useState(false);
  const [editingGroup, setEditingGroup] = useState<Group | null>(null);
  const [expandedId, setExpandedId] = useState<number | null>(null);

  const groups: Group[] = data?.items ?? [];

  const handleDelete = (e: React.MouseEvent, id: number) => {
    e.stopPropagation();
    if (window.confirm("删除组将同时移除所有成员和权限配置，确定删除？")) {
      deleteGroup.mutate(id);
      if (expandedId === id) setExpandedId(null);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900">组管理</h2>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          创建组
        </button>
      </div>

      <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
        <table className="w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">名称</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">描述</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">成员数</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">权限数</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">创建时间</th>
              <th className="px-4 py-3 text-right text-xs font-medium text-gray-500">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {isLoading ? (
              <tr><td colSpan={6} className="px-4 py-12 text-center text-sm text-gray-500">加载中...</td></tr>
            ) : groups.length === 0 ? (
              <tr><td colSpan={6} className="px-4 py-12 text-center text-sm text-gray-500">暂无组</td></tr>
            ) : (
              groups.map((g) => (
                <>
                  <tr
                    key={g.id}
                    onClick={() => setExpandedId(expandedId === g.id ? null : g.id)}
                    className="cursor-pointer transition-colors hover:bg-gray-50"
                  >
                    <td className="px-4 py-3 text-sm font-medium text-gray-900">{g.name}</td>
                    <td className="px-4 py-3 text-sm text-gray-600 max-w-[200px] truncate">{g.description || "-"}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{g.member_count ?? 0}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">{g.permission_count ?? 0}</td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      {g.created_at ? new Date(g.created_at).toLocaleDateString("zh-CN") : "-"}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button
                          onClick={(e) => { e.stopPropagation(); setEditingGroup(g); }}
                          className="text-sm text-blue-600 hover:text-blue-800"
                        >
                          编辑
                        </button>
                        <button
                          onClick={(e) => handleDelete(e, g.id)}
                          className="text-sm text-red-600 hover:text-red-800"
                        >
                          删除
                        </button>
                      </div>
                    </td>
                  </tr>
                  {expandedId === g.id && (
                    <tr key={`detail-${g.id}`}>
                      <td colSpan={6} className="bg-gray-50 p-0">
                        <GroupDetailPanel groupId={g.id} />
                      </td>
                    </tr>
                  )}
                </>
              ))
            )}
          </tbody>
        </table>
      </div>

      {showCreate && <GroupFormDialog onClose={() => setShowCreate(false)} />}
      {editingGroup && <GroupFormDialog group={editingGroup} onClose={() => setEditingGroup(null)} />}
    </div>
  );
}
