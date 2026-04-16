"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { GitRepository } from "@/types";

interface GitRepoForm {
  name: string;
  url: string;
  provider: string;
  auth_type: string;
  credential: string;
}

const emptyForm: GitRepoForm = {
  name: "",
  url: "",
  provider: "github",
  auth_type: "token",
  credential: "",
};

// Git 仓库管理页面：CRUD Git 仓库配置、测试连接
export default function GitReposPage() {
  const queryClient = useQueryClient();
  const [showDialog, setShowDialog] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<GitRepoForm>(emptyForm);
  const [testingId, setTestingId] = useState<number | null>(null);
  const [testResult, setTestResult] = useState<Record<number, { ok: boolean; msg: string }>>({});

  const { data: reposData, isLoading } = useQuery({
    queryKey: ["git-repos"],
    queryFn: async () => {
      const res = await apiClient.get("/git-repos");
      return res.data;
    },
  });
  const repos: GitRepository[] = reposData?.items ?? [];

  const createMutation = useMutation({
    mutationFn: async (data: GitRepoForm) => {
      const res = await apiClient.post("/git-repos", data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["git-repos"] });
      closeDialog();
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: number; data: GitRepoForm }) => {
      const res = await apiClient.put(`/git-repos/${id}`, data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["git-repos"] });
      closeDialog();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/git-repos/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["git-repos"] });
    },
  });

  const testConnection = async (id: number) => {
    setTestingId(id);
    try {
      const res = await apiClient.post(`/git-repos/${id}/test-connection`);
      setTestResult((prev) => ({
        ...prev,
        [id]: { ok: true, msg: res.data?.message || "连接成功" },
      }));
    } catch {
      setTestResult((prev) => ({
        ...prev,
        [id]: { ok: false, msg: "连接失败" },
      }));
    } finally {
      setTestingId(null);
    }
  };

  const openCreate = () => {
    setEditingId(null);
    setForm(emptyForm);
    setShowDialog(true);
  };

  const openEdit = (repo: GitRepository) => {
    setEditingId(repo.id);
    setForm({
      name: repo.name,
      url: repo.url,
      provider: repo.provider,
      auth_type: repo.auth_type || "token",
      credential: "",
    });
    setShowDialog(true);
  };

  const closeDialog = () => {
    setShowDialog(false);
    setEditingId(null);
    setForm(emptyForm);
  };

  const handleSubmit = () => {
    if (editingId) {
      updateMutation.mutate({ id: editingId, data: form });
    } else {
      createMutation.mutate(form);
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <button
          onClick={openCreate}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 transition-colors"
        >
          添加 Git 仓库
        </button>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : repos.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center">
          <p className="text-gray-500">暂无 Git 仓库，点击「添加 Git 仓库」开始</p>
        </div>
      ) : (
        <div className="rounded-lg border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">名称</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">URL</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Provider</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">认证方式</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {repos.map((repo) => (
                <tr key={repo.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{repo.name}</td>
                  <td className="px-4 py-3 text-sm text-gray-600 font-mono truncate max-w-[250px]">
                    {repo.url}
                  </td>
                  <td className="px-4 py-3">
                    <span className="inline-flex rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-700">
                      {repo.provider}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">{repo.auth_type || "-"}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => testConnection(repo.id)}
                        disabled={testingId === repo.id}
                        className="text-xs text-blue-600 hover:text-blue-800 disabled:opacity-50"
                      >
                        {testingId === repo.id ? "测试中..." : "测试连接"}
                      </button>
                      {testResult[repo.id] && (
                        <span
                          className={`text-xs ${
                            testResult[repo.id].ok ? "text-green-600" : "text-red-600"
                          }`}
                        >
                          {testResult[repo.id].msg}
                        </span>
                      )}
                      <button
                        onClick={() => openEdit(repo)}
                        className="text-xs text-gray-600 hover:text-gray-800"
                      >
                        编辑
                      </button>
                      <button
                        onClick={() => {
                          if (confirm("确定删除该仓库？")) deleteMutation.mutate(repo.id);
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
                {editingId ? "编辑 Git 仓库" : "添加 Git 仓库"}
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
                  placeholder="例如: my-app-repo"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">仓库 URL</label>
                <input
                  type="text"
                  value={form.url}
                  onChange={(e) => setForm({ ...form, url: e.target.value })}
                  placeholder="https://github.com/org/repo.git"
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
                  <option value="github">GitHub</option>
                  <option value="gitlab">GitLab</option>
                  <option value="gitee">Gitee</option>
                  <option value="bitbucket">Bitbucket</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">认证方式</label>
                <select
                  value={form.auth_type}
                  onChange={(e) => setForm({ ...form, auth_type: e.target.value })}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                >
                  <option value="token">Token</option>
                  <option value="ssh">SSH Key</option>
                  <option value="basic">Basic Auth</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {form.auth_type === "ssh" ? "SSH 私钥" : "认证令牌"}
                </label>
                <textarea
                  value={form.credential}
                  onChange={(e) => setForm({ ...form, credential: e.target.value })}
                  rows={3}
                  placeholder={form.auth_type === "ssh" ? "粘贴 SSH 私钥..." : "粘贴 Token..."}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
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
