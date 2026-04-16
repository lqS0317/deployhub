"use client";

import { useState, useEffect } from "react";
import { useCreatePlugin, useUpdatePlugin } from "@/hooks/use-route-plugins";
import { showToast } from "@/components/ui/toast";
import type { RoutePlugin } from "@/types";

interface CreatePluginDialogProps {
  open: boolean;
  onClose: () => void;
  editPlugin?: RoutePlugin | null;
}

export function CreatePluginDialog({
  open,
  onClose,
  editPlugin,
}: CreatePluginDialogProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [yamlContent, setYamlContent] = useState("");

  const createPlugin = useCreatePlugin();
  const updatePlugin = useUpdatePlugin();
  const isEdit = !!editPlugin;

  useEffect(() => {
    if (editPlugin) {
      setName(editPlugin.name);
      setDescription(editPlugin.description || "");
      setYamlContent(editPlugin.yaml_content || "");
    } else {
      setName("");
      setDescription("");
      setYamlContent("");
    }
  }, [editPlugin, open]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !yamlContent.trim()) return;

    const payload = {
      name: name.trim(),
      description: description.trim(),
      yaml_content: yamlContent,
    };

    if (isEdit) {
      updatePlugin.mutate(
        { id: editPlugin.id, ...payload },
        {
          onSuccess: () => {
            showToast("插件已更新", "success");
            onClose();
          },
          onError: () => showToast("更新失败", "error"),
        }
      );
    } else {
      createPlugin.mutate(payload, {
        onSuccess: () => {
          showToast("插件已创建", "success");
          onClose();
        },
        onError: () => showToast("创建失败", "error"),
      });
    }
  };

  if (!open) return null;

  const isPending = createPlugin.isPending || updatePlugin.isPending;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative w-full max-w-2xl rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">
            {isEdit ? "编辑插件" : "新建插件"}
          </h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              名称 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如: rate-limit"
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              描述
            </label>
            <input
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="插件用途描述"
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              YAML 内容 <span className="text-red-500">*</span>
            </label>
            <textarea
              value={yamlContent}
              onChange={(e) => setYamlContent(e.target.value)}
              rows={14}
              placeholder="apiVersion: traefik.io/v1alpha1&#10;kind: Middleware&#10;..."
              className="w-full rounded-md border border-gray-300 px-3 py-2 font-mono text-sm leading-relaxed focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div className="flex justify-end gap-3 border-t border-gray-200 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={!name.trim() || !yamlContent.trim() || isPending}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {isPending
                ? isEdit
                  ? "更新中..."
                  : "创建中..."
                : isEdit
                  ? "更新"
                  : "创建"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
