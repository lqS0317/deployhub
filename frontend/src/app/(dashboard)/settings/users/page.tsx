"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { CreateUserDialog } from "@/components/user/create-user-dialog";
import { PermissionDialog } from "@/components/user/permission-dialog";
import type { User, EffectivePermission } from "@/types";

// 用户权限摘要组件：内联显示权限和所属组
function UserPermBadges({ userId, role }: { userId: number; role: string }) {
  const [open, setOpen] = useState(false);

  const { data: permissions = [] } = useQuery({
    queryKey: ["permissions", "user", userId],
    queryFn: async () => {
      const res = await apiClient.get<{ items: EffectivePermission[] }>(`/users/${userId}/permissions`);
      return res.data.items ?? [];
    },
    enabled: role === "admin" ? false : true,
  });

  if (role === "admin") {
    return <span className="text-xs text-yellow-600 font-medium">全部服务（管理员）</span>;
  }

  if (permissions.length === 0) {
    return <span className="text-xs text-gray-400">无权限</span>;
  }

  const ROLE_COLORS: Record<string, string> = {
    owner: "bg-purple-100 text-purple-700",
    developer: "bg-blue-100 text-blue-700",
    viewer: "bg-gray-100 text-gray-700",
  };
  const ROLE_LABELS: Record<string, string> = { owner: "拥有者", developer: "开发者", viewer: "查看者" };

  const shown = open ? permissions : permissions.slice(0, 2);
  const hasMore = permissions.length > 2;

  return (
    <div className="flex flex-wrap gap-1 items-center">
      {shown.map((p) => (
        <span key={p.service_id} className={`inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-[10px] font-medium ${ROLE_COLORS[p.role] || "bg-gray-100 text-gray-600"}`}
          title={`${p.service_name}: ${ROLE_LABELS[p.role] || p.role} (来源: ${p.sources.map(s => s.name).join(", ")})`}>
          {p.service_name}
          <span className="opacity-60">·{ROLE_LABELS[p.role]?.[0] || p.role[0]}</span>
        </span>
      ))}
      {hasMore && !open && (
        <button onClick={(e) => { e.stopPropagation(); setOpen(true); }} className="text-[10px] text-blue-600 hover:text-blue-800">
          +{permissions.length - 2} 更多
        </button>
      )}
      {hasMore && open && (
        <button onClick={(e) => { e.stopPropagation(); setOpen(false); }} className="text-[10px] text-gray-400 hover:text-gray-600">
          收起
        </button>
      )}
    </div>
  );
}

// 用户所属组组件
function UserGroupBadges({ userId }: { userId: number }) {
  const [open, setOpen] = useState(false);

  const { data: groups = [] } = useQuery({
    queryKey: ["user-groups", userId],
    queryFn: async () => {
      const res = await apiClient.get<{ items: { id: number; name: string }[] }>(`/users/${userId}/groups`);
      return res.data.items ?? [];
    },
  });

  if (groups.length === 0) {
    return <span className="text-xs text-gray-400">无</span>;
  }

  const shown = open ? groups : groups.slice(0, 2);
  const hasMore = groups.length > 2;

  return (
    <div className="flex flex-wrap gap-1 items-center">
      {shown.map((g) => (
        <span key={g.id} className="inline-flex rounded px-1.5 py-0.5 text-[10px] font-medium bg-indigo-100 text-indigo-700">
          {g.name}
        </span>
      ))}
      {hasMore && !open && (
        <button onClick={(e) => { e.stopPropagation(); setOpen(true); }} className="text-[10px] text-blue-600 hover:text-blue-800">
          +{groups.length - 2} 更多
        </button>
      )}
      {hasMore && open && (
        <button onClick={(e) => { e.stopPropagation(); setOpen(false); }} className="text-[10px] text-gray-400 hover:text-gray-600">
          收起
        </button>
      )}
    </div>
  );
}

export default function UsersPage() {
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [viewPermUser, setViewPermUser] = useState<User | null>(null);

  const { data: usersData, isLoading } = useQuery({
    queryKey: ["users"],
    queryFn: async () => {
      const res = await apiClient.get("/users");
      return res.data;
    },
  });
  const users: User[] = usersData?.items ?? [];

  const roleMutation = useMutation({
    mutationFn: async ({ id, role }: { id: number; role: string }) => {
      await apiClient.put(`/users/${id}/role`, { role });
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["users"] }),
  });

  const statusMutation = useMutation({
    mutationFn: async ({ id, status }: { id: number; status: string }) => {
      await apiClient.put(`/users/${id}/status`, { status });
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["users"] }),
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-900">用户管理</h2>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          创建用户
        </button>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : users.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center">
          <p className="text-gray-500">暂无用户</p>
        </div>
      ) : (
        <div className="rounded-lg border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">用户名</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">邮箱</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">角色</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">状态</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">服务权限</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500">所属组</th>
                <th className="px-4 py-3 text-right text-xs font-medium text-gray-500">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {users.map((user) => (
                <tr key={user.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{user.username}</td>
                  <td className="px-4 py-3 text-sm text-gray-600">{user.email}</td>
                  <td className="px-4 py-3">
                    <select
                      value={user.role}
                      onChange={(e) => roleMutation.mutate({ id: user.id, role: e.target.value })}
                      disabled={roleMutation.isPending}
                      className="rounded border border-gray-300 px-2 py-1 text-xs focus:border-blue-500 focus:outline-none"
                    >
                      <option value="admin">管理员</option>
                      <option value="member">成员</option>
                    </select>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
                      user.status === "active" ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"
                    }`}>
                      {user.status === "active" ? "启用" : "禁用"}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <UserPermBadges userId={user.id} role={user.role} />
                  </td>
                  <td className="px-4 py-3">
                    <UserGroupBadges userId={user.id} />
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => setViewPermUser(user)}
                        className="inline-flex items-center gap-1 rounded-md bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700 hover:bg-blue-100 transition-colors"
                      >
                        <svg className="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                        </svg>
                        权限管理
                      </button>
                      <button
                        onClick={() => statusMutation.mutate({
                          id: user.id,
                          status: user.status === "active" ? "disabled" : "active",
                        })}
                        disabled={statusMutation.isPending}
                        className={`rounded px-2.5 py-1 text-xs font-medium text-white transition-colors disabled:opacity-50 ${
                          user.status === "active" ? "bg-red-600 hover:bg-red-700" : "bg-green-600 hover:bg-green-700"
                        }`}
                      >
                        {user.status === "active" ? "禁用" : "启用"}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showCreate && <CreateUserDialog onClose={() => setShowCreate(false)} />}
      {viewPermUser && (
        <PermissionDialog
          userId={viewPermUser.id}
          username={viewPermUser.username}
          onClose={() => setViewPermUser(null)}
        />
      )}
    </div>
  );
}
