"use client";

import { useState, useMemo } from "react";
import {
  useCreateItem,
  useUpdateItem,
  useDeleteItem,
} from "@/hooks/use-config-center";
import { showToast } from "@/components/ui/toast";
import type { ConfigItem } from "@/types";

interface KvTableEditorProps {
  items: ConfigItem[];
  entryId: number;
  isSecret: boolean;
  canEdit: boolean;
}

interface EditingCell {
  itemId: number;
  field: "value" | "comment";
  text: string;
}

export function KvTableEditor({
  items,
  entryId,
  isSecret,
  canEdit,
}: KvTableEditorProps) {
  const [search, setSearch] = useState("");
  const [editingCell, setEditingCell] = useState<EditingCell | null>(null);
  const [revealedIds, setRevealedIds] = useState<Set<number>>(new Set());
  const [newKey, setNewKey] = useState("");
  const [newValue, setNewValue] = useState("");
  const [newComment, setNewComment] = useState("");

  const createItem = useCreateItem();
  const updateItem = useUpdateItem();
  const deleteItem = useDeleteItem();

  const filtered = useMemo(() => {
    const active = items.filter((i) => !i.is_deleted);
    if (!search) return active;
    const q = search.toLowerCase();
    return active.filter(
      (i) =>
        i.key.toLowerCase().includes(q) ||
        i.value.toLowerCase().includes(q) ||
        i.comment?.toLowerCase().includes(q)
    );
  }, [items, search]);

  const startEdit = (item: ConfigItem, field: "value" | "comment") => {
    if (!canEdit) return;
    setEditingCell({
      itemId: item.id,
      field,
      text: field === "value" ? item.value : item.comment ?? "",
    });
  };

  const commitEdit = () => {
    if (!editingCell) return;
    const item = items.find((i) => i.id === editingCell.itemId);
    if (!item) return;

    updateItem.mutate(
      {
        entryId,
        itemId: editingCell.itemId,
        value: editingCell.field === "value" ? editingCell.text : item.value,
        comment: editingCell.field === "comment" ? editingCell.text : item.comment ?? "",
      },
      {
        onSuccess: () => showToast("已保存", "success"),
        onError: () => showToast("保存失败", "error"),
      }
    );
    setEditingCell(null);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") commitEdit();
    if (e.key === "Escape") setEditingCell(null);
  };

  const handleAddRow = () => {
    if (!newKey.trim()) return;
    createItem.mutate(
      { entryId, key: newKey.trim(), value: newValue, comment: newComment },
      {
        onSuccess: () => {
          showToast("配置项已添加", "success");
          setNewKey("");
          setNewValue("");
          setNewComment("");
        },
        onError: () => showToast("添加失败", "error"),
      }
    );
  };

  const handleDelete = (item: ConfigItem) => {
    if (!window.confirm(`确定删除配置项「${item.key}」？`)) return;
    deleteItem.mutate(
      { entryId, itemId: item.id },
      {
        onSuccess: () => showToast("已删除", "success"),
        onError: () => showToast("删除失败", "error"),
      }
    );
  };

  const toggleReveal = (id: number) => {
    setRevealedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const displayValue = (item: ConfigItem) => {
    if (isSecret && !revealedIds.has(item.id)) return "••••••";
    return item.value;
  };

  return (
    <div className="space-y-3">
      {/* 搜索 */}
      <div className="relative max-w-xs">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="搜索 Key / Value..."
          className="w-full rounded-md border border-gray-300 bg-white px-3 py-1.5 pl-8 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
        <svg
          className="absolute left-2.5 top-2 h-3.5 w-3.5 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      </div>

      {/* 表格 */}
      <div className="overflow-hidden rounded-lg border border-gray-200">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                Key
              </th>
              <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                Value
              </th>
              <th className="px-4 py-2.5 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                备注
              </th>
              {canEdit && (
                <th className="px-4 py-2.5 text-right text-xs font-medium uppercase tracking-wider text-gray-500">
                  操作
                </th>
              )}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 bg-white">
            {filtered.length === 0 && (
              <tr>
                <td
                  colSpan={canEdit ? 4 : 3}
                  className="px-4 py-8 text-center text-sm text-gray-400"
                >
                  {search ? "未找到匹配项" : "暂无配置项"}
                </td>
              </tr>
            )}
            {filtered.map((item) => (
              <tr key={item.id} className="hover:bg-gray-50">
                <td className="whitespace-nowrap px-4 py-2 text-sm font-mono text-gray-900">
                  {item.key}
                </td>
                <td
                  className="max-w-xs cursor-pointer px-4 py-2 text-sm font-mono text-gray-700"
                  onClick={() => startEdit(item, "value")}
                >
                  {editingCell?.itemId === item.id && editingCell.field === "value" ? (
                    <input
                      autoFocus
                      value={editingCell.text}
                      onChange={(e) => setEditingCell({ ...editingCell, text: e.target.value })}
                      onBlur={commitEdit}
                      onKeyDown={handleKeyDown}
                      className="w-full rounded border border-blue-400 px-1.5 py-0.5 text-sm font-mono focus:outline-none"
                    />
                  ) : (
                    <div className="flex items-center gap-1.5">
                      <span className="truncate">{displayValue(item)}</span>
                      {isSecret && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            toggleReveal(item.id);
                          }}
                          className="shrink-0 text-xs text-gray-400 hover:text-gray-600"
                        >
                          {revealedIds.has(item.id) ? "隐藏" : "显示"}
                        </button>
                      )}
                    </div>
                  )}
                </td>
                <td
                  className="max-w-xs cursor-pointer px-4 py-2 text-sm text-gray-500"
                  onClick={() => startEdit(item, "comment")}
                >
                  {editingCell?.itemId === item.id && editingCell.field === "comment" ? (
                    <input
                      autoFocus
                      value={editingCell.text}
                      onChange={(e) => setEditingCell({ ...editingCell, text: e.target.value })}
                      onBlur={commitEdit}
                      onKeyDown={handleKeyDown}
                      className="w-full rounded border border-blue-400 px-1.5 py-0.5 text-sm focus:outline-none"
                    />
                  ) : (
                    <span className="truncate">
                      {item.comment || (canEdit ? "点击添加" : "-")}
                    </span>
                  )}
                </td>
                {canEdit && (
                  <td className="whitespace-nowrap px-4 py-2 text-right">
                    <button
                      onClick={() => handleDelete(item)}
                      className="text-xs text-red-600 hover:text-red-800"
                    >
                      删除
                    </button>
                  </td>
                )}
              </tr>
            ))}

            {/* 新增行 */}
            {canEdit && (
              <tr className="bg-gray-50/50">
                <td className="px-4 py-2">
                  <input
                    type="text"
                    value={newKey}
                    onChange={(e) => setNewKey(e.target.value)}
                    placeholder="新 Key"
                    className="w-full rounded border border-gray-300 px-2 py-1 text-sm font-mono focus:border-blue-400 focus:outline-none"
                  />
                </td>
                <td className="px-4 py-2">
                  <input
                    type="text"
                    value={newValue}
                    onChange={(e) => setNewValue(e.target.value)}
                    placeholder="Value"
                    className="w-full rounded border border-gray-300 px-2 py-1 text-sm font-mono focus:border-blue-400 focus:outline-none"
                  />
                </td>
                <td className="px-4 py-2">
                  <input
                    type="text"
                    value={newComment}
                    onChange={(e) => setNewComment(e.target.value)}
                    placeholder="备注"
                    className="w-full rounded border border-gray-300 px-2 py-1 text-sm focus:border-blue-400 focus:outline-none"
                  />
                </td>
                <td className="px-4 py-2 text-right">
                  <button
                    onClick={handleAddRow}
                    disabled={!newKey.trim() || createItem.isPending}
                    className="rounded bg-blue-600 px-3 py-1 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                  >
                    添加
                  </button>
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
