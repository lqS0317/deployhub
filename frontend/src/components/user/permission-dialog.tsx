"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useUserPermissions } from "@/hooks/use-permissions";
import apiClient from "@/lib/api-client";
import { showToast } from "@/components/ui/toast";

const ROLE_LABELS: Record<string, string> = {
  owner: "拥有者",
  developer: "开发者",
  viewer: "查看者",
};

const ROLE_COLORS: Record<string, string> = {
  owner: "bg-purple-100 text-purple-700",
  developer: "bg-blue-100 text-blue-700",
  viewer: "bg-gray-100 text-gray-700",
};

const ROLES = [
  { value: "viewer", label: "查看者" },
  { value: "developer", label: "开发者" },
  { value: "owner", label: "拥有者" },
];

interface Props {
  userId: number;
  username: string;
  onClose: () => void;
}

export function PermissionDialog({ userId, username, onClose }: Props) {
  const qc = useQueryClient();
  const { data: permissions = [], isLoading } = useUserPermissions(userId);
  const [addServiceId, setAddServiceId] = useState("");
  const [addRole, setAddRole] = useState("viewer");

  const { data: servicesData } = useQuery({
    queryKey: ["services-all-for-perm"],
    queryFn: () => apiClient.get("/services").then((r) => r.data),
  });
  const allServices = servicesData?.items ?? [];

  const addMutation = useMutation({
    mutationFn: async () => {
      await apiClient.post(`/services/${addServiceId}/members`, { user_id: userId, role: addRole });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["permissions", "user", userId] });
      setAddServiceId("");
      showToast("个人权限已分配", "success");
    },
  });

  const removeMutation = useMutation({
    mutationFn: async (serviceId: number) => {
      const membersRes = await apiClient.get(`/services/${serviceId}/members`);
      const members = membersRes.data?.items ?? membersRes.data ?? [];
      const member = members.find((m: { user_id: number }) => m.user_id === userId);
      if (member) {
        await apiClient.delete(`/services/${serviceId}/members/${member.id}`);
      }
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["permissions", "user", userId] });
      showToast("个人权限已移除", "success");
    },
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative z-10 w-full max-w-2xl max-h-[80vh] overflow-y-auto rounded-xl bg-white p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold">{username} 的权限管理</h2>
          <button onClick={onClose} className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600">
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* 添加个人权限 */}
        <div className="mb-4 rounded-lg border border-gray-200 p-3">
          <h3 className="text-sm font-medium text-gray-700 mb-2">分配个人 Service 权限</h3>
          <div className="flex gap-2">
            <select
              value={addServiceId}
              onChange={(e) => setAddServiceId(e.target.value)}
              className="flex-1 rounded border border-gray-300 px-2 py-1.5 text-sm"
            >
              <option value="">选择 Service...</option>
              {allServices.map((s: { id: number; name: string }) => (
                <option key={s.id} value={s.id}>{s.name}</option>
              ))}
            </select>
            <select value={addRole} onChange={(e) => setAddRole(e.target.value)} className="rounded border border-gray-300 px-2 py-1.5 text-sm">
              {ROLES.map((r) => <option key={r.value} value={r.value}>{r.label}</option>)}
            </select>
            <button
              onClick={() => addMutation.mutate()}
              disabled={!addServiceId || addMutation.isPending}
              className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
            >
              分配
            </button>
          </div>
        </div>

        {/* 有效权限列表 */}
        {isLoading ? (
          <p className="text-center text-sm text-gray-500 py-8">加载中...</p>
        ) : permissions.length === 0 ? (
          <p className="text-center text-sm text-gray-500 py-8">暂无权限</p>
        ) : (
          <table className="w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">Service</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">有效角色</th>
                <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">来源</th>
                <th className="px-4 py-2 text-right text-xs font-medium text-gray-500">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {permissions.map((p) => {
                const hasPersonal = p.sources.some((s) => s.type === "personal");
                return (
                  <tr key={p.service_id}>
                    <td className="px-4 py-2 text-sm font-medium text-gray-900">{p.service_name}</td>
                    <td className="px-4 py-2">
                      <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${ROLE_COLORS[p.role] || "bg-gray-100 text-gray-600"}`}>
                        {ROLE_LABELS[p.role] || p.role}
                      </span>
                    </td>
                    <td className="px-4 py-2">
                      <div className="flex flex-wrap gap-1">
                        {p.sources.map((s, i) => (
                          <span
                            key={i}
                            className={`inline-flex rounded px-1.5 py-0.5 text-xs ${
                              s.type === "admin" ? "bg-yellow-100 text-yellow-700" :
                              s.type === "personal" ? "bg-green-100 text-green-700" :
                              "bg-blue-100 text-blue-700"
                            }`}
                          >
                            {s.name}
                          </span>
                        ))}
                      </div>
                    </td>
                    <td className="px-4 py-2 text-right">
                      {hasPersonal && (
                        <button
                          onClick={() => removeMutation.mutate(p.service_id)}
                          disabled={removeMutation.isPending}
                          className="text-xs text-red-500 hover:text-red-700 disabled:opacity-50"
                        >
                          移除个人权限
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
