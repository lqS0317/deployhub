"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { useAddClusterNamespace, useClusterNamespaces, useDeleteClusterNamespace, useSyncClusterNamespaces } from "@/hooks/use-namespaces";
import type { Cluster, ClusterNamespace } from "@/types";

interface ClusterForm {
  name: string;
  display_name: string;
  env: string;
  api_server: string;
  kubeconfig: string;
  helm_service_account: string;
  build_service_account: string;
}

const emptyForm: ClusterForm = {
  name: "",
  display_name: "",
  env: "devnet",
  api_server: "",
  kubeconfig: "",
  helm_service_account: "",
  build_service_account: "",
};

// 集群管理页面：CRUD 集群配置、测试连接
export default function ClustersPage() {
  const queryClient = useQueryClient();
  const [showDialog, setShowDialog] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<ClusterForm>(emptyForm);
  const [testingId, setTestingId] = useState<number | null>(null);
  const [testResult, setTestResult] = useState<Record<number, { ok: boolean; msg: string }>>({});
  const [namespaceCluster, setNamespaceCluster] = useState<Cluster | null>(null);
  const [namespaceInput, setNamespaceInput] = useState("");
  const [namespaceIsDefault, setNamespaceIsDefault] = useState(false);

  const { data: clustersData, isLoading } = useQuery({
    queryKey: ["clusters"],
    queryFn: async () => {
      const res = await apiClient.get("/clusters");
      return res.data;
    },
  });
  const clusters: Cluster[] = clustersData?.items ?? [];
  const namespaceClusterId = namespaceCluster?.id ?? 0;

  const { data: clusterNamespaces, isLoading: namespacesLoading } = useClusterNamespaces(namespaceClusterId);
  const addNamespaceMutation = useAddClusterNamespace(namespaceClusterId);
  const deleteNamespaceMutation = useDeleteClusterNamespace(namespaceClusterId);
  const syncNamespacesMutation = useSyncClusterNamespaces(namespaceClusterId);

  const createMutation = useMutation({
    mutationFn: async (data: ClusterForm) => {
      const res = await apiClient.post("/clusters", data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["clusters"] });
      closeDialog();
    },
  });

  const updateMutation = useMutation({
    mutationFn: async ({ id, data }: { id: number; data: ClusterForm }) => {
      const res = await apiClient.put(`/clusters/${id}`, data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["clusters"] });
      closeDialog();
    },
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/clusters/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["clusters"] });
    },
  });

  const testConnection = async (id: number) => {
    setTestingId(id);
    try {
      const res = await apiClient.post(`/clusters/${id}/test`);
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

  const openEdit = (cluster: Cluster) => {
    setEditingId(cluster.id);
    setForm({
      name: cluster.name,
      display_name: cluster.display_name || "",
      env: cluster.env,
      api_server: cluster.api_server || "",
      kubeconfig: "",
      helm_service_account: cluster.helm_service_account || "",
      build_service_account: cluster.build_service_account || "",
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

  const openNamespaceDialog = (cluster: Cluster) => {
    setNamespaceCluster(cluster);
    setNamespaceInput("");
    setNamespaceIsDefault(false);
  };

  const closeNamespaceDialog = () => {
    setNamespaceCluster(null);
    setNamespaceInput("");
    setNamespaceIsDefault(false);
  };

  const addNamespace = () => {
    if (!namespaceInput.trim() || !namespaceClusterId) return;
    addNamespaceMutation.mutate(
      { namespace: namespaceInput.trim(), is_default: namespaceIsDefault },
      {
        onSuccess: () => {
          setNamespaceInput("");
          setNamespaceIsDefault(false);
        },
      }
    );
  };

  const statusBadge = (status: string) => {
    const styles: Record<string, string> = {
      connected: "bg-green-100 text-green-800",
      disconnected: "bg-red-100 text-red-800",
      unknown: "bg-gray-100 text-gray-800",
    };
    const labels: Record<string, string> = {
      connected: "已连接",
      disconnected: "未连接",
      unknown: "未知",
    };
    return (
      <span
        className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
          styles[status] || styles.unknown
        }`}
      >
        {labels[status] || status}
      </span>
    );
  };

  const envBadge = (env: string) => {
    const styles: Record<string, string> = {
      mainnet: "bg-red-100 text-red-800",
      testnet: "bg-yellow-100 text-yellow-800",
      qanet: "bg-blue-100 text-blue-800",
      devnet: "bg-green-100 text-green-800",
    };
    return (
      <span
        className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
          styles[env] || "bg-gray-100 text-gray-800"
        }`}
      >
        {env}
      </span>
    );
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <button
          onClick={openCreate}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 transition-colors"
        >
          添加集群
        </button>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : clusters.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center">
          <p className="text-gray-500">暂无集群，点击「添加集群」开始</p>
        </div>
      ) : (
        <div className="rounded-lg border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">名称</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">显示名</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">环境</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">API Server</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {clusters.map((cluster) => (
                <tr key={cluster.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{cluster.name}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{cluster.display_name || "-"}</td>
                  <td className="px-4 py-3">{envBadge(cluster.env)}</td>
                  <td className="px-4 py-3 text-sm text-gray-600 font-mono truncate max-w-[200px]">
                    {cluster.api_server}
                  </td>
                  <td className="px-4 py-3">
                    {statusBadge(cluster.status || "unknown")}
                    {testResult[cluster.id] && (
                      <span
                        className={`ml-2 text-xs ${
                          testResult[cluster.id].ok ? "text-green-600" : "text-red-600"
                        }`}
                      >
                        {testResult[cluster.id].msg}
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => testConnection(cluster.id)}
                        disabled={testingId === cluster.id}
                        className="text-xs text-blue-600 hover:text-blue-800 disabled:opacity-50"
                      >
                        {testingId === cluster.id ? "测试中..." : "测试连接"}
                      </button>
                      <button
                        onClick={() => openEdit(cluster)}
                        className="text-xs text-gray-600 hover:text-gray-800"
                      >
                        编辑
                      </button>
                      <button
                        onClick={() => openNamespaceDialog(cluster)}
                        className="text-xs text-purple-600 hover:text-purple-800"
                      >
                        命名空间映射
                      </button>
                      <button
                        onClick={() => {
                          if (confirm("确定删除该集群？")) deleteMutation.mutate(cluster.id);
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
                {editingId ? "编辑集群" : "添加集群"}
              </h3>
              <button onClick={closeDialog} className="text-gray-400 hover:text-gray-600">✕</button>
            </div>
            <div className="px-6 py-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">集群名称</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="例如: prod-cluster-01"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">显示名称</label>
                <input
                  type="text"
                  value={form.display_name}
                  onChange={(e) => setForm({ ...form, display_name: e.target.value })}
                  placeholder="例如: 生产集群 01"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">环境</label>
                <select
                  value={form.env}
                  onChange={(e) => setForm({ ...form, env: e.target.value })}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                >
                  <option value="devnet">Devnet</option>
                  <option value="qanet">QAnet</option>
                  <option value="testnet">Testnet</option>
                  <option value="mainnet">Mainnet</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">API Server</label>
                <input
                  type="text"
                  value={form.api_server}
                  onChange={(e) => setForm({ ...form, api_server: e.target.value })}
                  placeholder="https://k8s-api.example.com:6443"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Kubeconfig</label>
                <textarea
                  value={form.kubeconfig}
                  onChange={(e) => setForm({ ...form, kubeconfig: e.target.value })}
                  rows={4}
                  placeholder="粘贴 kubeconfig 内容..."
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Helm ServiceAccount</label>
                <input
                  type="text"
                  value={form.helm_service_account}
                  onChange={(e) => setForm({ ...form, helm_service_account: e.target.value })}
                  placeholder="留空则使用 default，例如: helm-deployer"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
                <p className="mt-1 text-xs text-gray-500">Helm Runner Job 使用的 ServiceAccount，留空使用 default</p>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Build ServiceAccount</label>
                <input
                  type="text"
                  value={form.build_service_account}
                  onChange={(e) => setForm({ ...form, build_service_account: e.target.value })}
                  placeholder="留空则使用 default，例如: kaniko-builder"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                />
                <p className="mt-1 text-xs text-gray-500">Kaniko 构建 Job 使用的 ServiceAccount（IRSA 推送 ECR 等场景），留空使用 default</p>
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
                disabled={!form.name || !form.api_server || createMutation.isPending || updateMutation.isPending}
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

      {/* 命名空间映射管理 */}
      {namespaceCluster && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/50" onClick={closeNamespaceDialog} />
          <div className="relative w-full max-w-2xl rounded-lg bg-white shadow-xl">
            <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
              <h3 className="text-lg font-semibold text-gray-900">
                命名空间映射 - {namespaceCluster.display_name || namespaceCluster.name}
              </h3>
              <button onClick={closeNamespaceDialog} className="text-gray-400 hover:text-gray-600">✕</button>
            </div>
            <div className="space-y-4 px-6 py-4">
              <div className="rounded-md border border-yellow-200 bg-yellow-50 p-3 text-xs text-yellow-800">
                发布流程只允许选择此处已登记的 namespace。未配置映射将无法发起部署。
              </div>

              <div className="flex flex-wrap items-end gap-2">
                <div className="min-w-[240px] flex-1">
                  <label className="mb-1 block text-sm font-medium text-gray-700">新增 namespace</label>
                  <input
                    type="text"
                    value={namespaceInput}
                    onChange={(e) => setNamespaceInput(e.target.value)}
                    placeholder="例如：ops / production"
                    className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  />
                </div>
                <label className="mb-2 flex items-center gap-2 text-sm text-gray-700">
                  <input
                    type="checkbox"
                    checked={namespaceIsDefault}
                    onChange={(e) => setNamespaceIsDefault(e.target.checked)}
                  />
                  设为默认
                </label>
                <button
                  onClick={addNamespace}
                  disabled={!namespaceInput.trim() || addNamespaceMutation.isPending}
                  className="mb-0.5 rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
                >
                  {addNamespaceMutation.isPending ? "添加中..." : "添加"}
                </button>
                <button
                  onClick={() => syncNamespacesMutation.mutate()}
                  disabled={syncNamespacesMutation.isPending}
                  className="mb-0.5 rounded-md border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:opacity-50"
                >
                  {syncNamespacesMutation.isPending ? "同步中..." : "从集群同步"}
                </button>
              </div>

              <div className="max-h-72 overflow-y-auto rounded-lg border border-gray-200">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-gray-200 bg-gray-50">
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">Namespace</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">默认</th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {namespacesLoading ? (
                      <tr>
                        <td className="px-3 py-4 text-sm text-gray-500" colSpan={3}>加载中...</td>
                      </tr>
                    ) : (clusterNamespaces ?? []).length === 0 ? (
                      <tr>
                        <td className="px-3 py-4 text-sm text-gray-500" colSpan={3}>
                          当前无 namespace 映射，请先添加
                        </td>
                      </tr>
                    ) : (
                      (clusterNamespaces as ClusterNamespace[]).map((ns) => (
                        <tr key={ns.id}>
                          <td className="px-3 py-2 text-sm font-mono text-gray-800">{ns.namespace}</td>
                          <td className="px-3 py-2 text-sm">
                            {ns.is_default ? (
                              <span className="rounded bg-green-100 px-2 py-0.5 text-xs text-green-700">默认</span>
                            ) : (
                              <span className="text-gray-400">-</span>
                            )}
                          </td>
                          <td className="px-3 py-2 text-sm">
                            <button
                              onClick={() => deleteNamespaceMutation.mutate(ns.id)}
                              disabled={deleteNamespaceMutation.isPending}
                              className="text-red-600 hover:text-red-800 disabled:opacity-50"
                            >
                              删除
                            </button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </div>
            <div className="flex justify-end border-t border-gray-200 px-6 py-4">
              <button
                onClick={closeNamespaceDialog}
                className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
              >
                关闭
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
