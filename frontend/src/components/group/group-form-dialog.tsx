"use client";

import { useState } from "react";
import { useCreateGroup, useUpdateGroup } from "@/hooks/use-groups";
import { showToast } from "@/components/ui/toast";
import type { Group } from "@/types";

interface Props {
  group?: Group;
  onClose: () => void;
}

export function GroupFormDialog({ group, onClose }: Props) {
  const isEdit = !!group;
  const [name, setName] = useState(group?.name ?? "");
  const [description, setDescription] = useState(group?.description ?? "");

  const createGroup = useCreateGroup();
  const updateGroup = useUpdateGroup(group?.id ?? 0);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    const mutation = isEdit ? updateGroup : createGroup;
    mutation.mutate(
      { name: name.trim(), description },
      {
        onSuccess: () => {
          showToast(isEdit ? "组已更新" : "组已创建", "success");
          onClose();
        },
      }
    );
  };

  const isPending = isEdit ? updateGroup.isPending : createGroup.isPending;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative z-10 w-full max-w-md rounded-xl bg-white p-6 shadow-2xl">
        <h2 className="text-lg font-semibold mb-4">{isEdit ? "编辑组" : "创建组"}</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">组名称 *</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如: 后端团队"
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">描述</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
              placeholder="组的用途描述..."
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className="rounded-lg border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
              取消
            </button>
            <button type="submit" disabled={isPending || !name.trim()} className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50">
              {isPending ? "保存中..." : isEdit ? "保存" : "创建"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
