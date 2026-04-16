"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { ImageRegistry } from "@/types";

interface RegistryForm {
  name: string;
  url: string;
  provider: string;
  is_default: boolean;
  username: string;
  password: string;
}

const emptyForm: RegistryForm = {
  name: "",
  url: "",
  provider: "docker-hub",
  is_default: false,
  username: "",
  password: "",
};

// 镜像仓库管理页面：CRUD 镜像仓库配置
export default function RegistriesPage() {
  const queryClient = useQueryClient();
  const [showDialog, setShowDialog] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<RegistryForm>(emptyForm);

  const { data: registriesData, isLoading } = useQuery({
    queryKey: ["registries"],
    queryFn: async () => {
      const res = await apiClient.get("/registries");
      return res.data;
    },
  });
  const registries: ImageRegistry[] = registriesData?.items ?? [];

  const createMutation = useMutation({
    mutationFn: async (data: Record<string, string>) => {
      const res = await apiClient.post("/registries", data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["registries"] });
      closeDialog();
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: number; data: Record<string, string> }) => {
      const res = await apiClient.put(`/registries/${id}`, data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["registries"] });
      closeDialog();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/registries/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["registries"] });
    },
  });

  const openCreate = () => {
    setEditingId(null);
    setForm(emptyForm);
    setShowDialog(true);
  };

  const openEdit = (registry: ImageRegistry) => {
    setEditingId(registry.id);
    setForm({
      name: registry.name,
      url: registry.url,
      provider: registry.provider,
      is_default: registry.is_default,
      username: "",
      password: "",
    });
    setShowDialog(true);
  };

  const closeDialog = () => {
    setShowDialog(false);
    setEditingId(null);
    setForm(emptyForm);
  };

  const handleSubmit = () => {
    const payload = {
      name: form.name,
      url: form.url,
      provider: form.provider,
      auth_config: JSON.stringify({ username: form.username, password: form.password }),
    };
    if (editingId) {
      updateMutation.mutate({ id: editingId, data: payload });
    } else {
      createMutation.mutate(payload);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <button
          onClick={openCreate}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 transition-colors"
        >
          添加镜像仓库
        </button>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : registries.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center">
          <p className="text-gray-500">暂无镜像仓库，点击「添加镜像仓库」开始</p>
        </div>
      ) : (
        <div className="rounded-lg border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">名称</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">URL</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Provider</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">是否默认</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {registries.map((registry) => (
                <tr key={registry.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{registry.name}</td>
                  <td className="px-4 py-3 text-sm text-gray-600 font-mono truncate max-w-[250px]">
                    {registry.url}
                  </td>
                  <td className="px-4 py-3">
                    <span className="inline-flex rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-700">
                      {registry.provider}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    {registry.is_default ? (
                      <span className="inline-flex rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800">
                        默认
                      </span>
                    ) : (
                      <span className="text-sm text-gray-400">-</span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => openEdit(registry)}
                        className="text-xs text-gray-600 hover:text-gray-800"
                      >
                        编辑
                      </button>
                      <button
                        onClick={() => {
                          if (confirm("确定删除该镜像仓库？")) deleteMutation.mutate(registry.id);
                        }}
                        className="text-xs text-red-600 hover:text-red-800"
                      >
                        删除
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* 创建/编辑对话框 */}
      {showDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/50" onClick={closeDialog} />
          <div className="relative w-full max-w-lg rounded-lg bg-white shadow-xl">
            <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
              <h3 className="text-lg font-semibold text-gray-900">
                {editingId ? "编辑镜像仓库" : "添加镜像仓库"}
              </h3>
              <button onClick={closeDialog} className="text-gray-400 hover:text-gray-600">✕</button>
            </div>
            <div className="px-6 py-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">仓库名称</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="例如: docker-hub-main"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">仓库 URL</label>
                <input
                  type="text"
                  value={form.url}
                  onChange={(e) => setForm({ ...form, url: e.target.value })}
                  placeholder="https://registry.example.com"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Provider</label>
                <select
                  value={form.provider}
                  onChange={(e) => setForm({ ...form, provider: e.target.value })}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                >
                  <option value="docker-hub">Docker Hub</option>
                  <option value="harbor">Harbor</option>
                  <option value="acr">阿里云 ACR</option>
                  <option value="ecr">AWS ECR</option>
                  <option value="gcr">Google GCR</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">用户名</label>
                <input
                  type="text"
                  value={form.username}
                  onChange={(e) => setForm({ ...form, username: e.target.value })}
                  placeholder="registry 用户名"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">密码</label>
                <input
                  type="password"
                  value={form.password}
                  onChange={(e) => setForm({ ...form, password: e.target.value })}
                  placeholder="registry 密码"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_default"
                  checked={form.is_default}
                  onChange={(e) => setForm({ ...form, is_default: e.target.checked })}
                  className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <label htmlFor="is_default" className="text-sm text-gray-700">
                  设为默认仓库
                </label>
              </div>
            </div>
            <div className="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
              <button
                onClick={closeDialog}
                className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                取消
              </button>
              <button
                onClick={handleSubmit}
                disabled={!form.name || !form.url || createMutation.isPending || updateMutation.isPending}
                className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                {createMutation.isPending || updateMutation.isPending
                  ? "保存中..."
                  : editingId
                    ? "保存"
                    : "添加"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
