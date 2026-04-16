"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  useGroupMembers, useAddGroupMembers, useRemoveGroupMember,
  useGroupPermissions, useAddGroupPermission, useUpdateGroupPermission, useRemoveGroupPermission,
} from "@/hooks/use-groups";
import apiClient from "@/lib/api-client";
import { showToast } from "@/components/ui/toast";

const ROLES = [
  { value: "viewer", label: "查看者" },
  { value: "developer", label: "开发者" },
  { value: "owner", label: "拥有者" },
];

export function GroupDetailPanel({ groupId }: { groupId: number }) {
  return (
    <div className="grid grid-cols-2 gap-4 p-4">
      <MembersSection groupId={groupId} />
      <PermissionsSection groupId={groupId} />
    </div>
  );
}

function MembersSection({ groupId }: { groupId: number }) {
  const { data: members = [], isLoading } = useGroupMembers(groupId);
  const addMembers = useAddGroupMembers(groupId);
  const removeMember = useRemoveGroupMember(groupId);
  const [selectedUserId, setSelectedUserId] = useState("");

  const { data: usersData } = useQuery({
    queryKey: ["users-for-group"],
    queryFn: () => apiClient.get("/users").then((r) => r.data),
  });
  const allUsers = usersData?.items ?? [];
  const existingIds = new Set(members.map((m) => m.user_id));
  const availableUsers = allUsers.filter((u: { id: number }) => !existingIds.has(u.id));

  return (
    <div>
      <h3 className="text-sm font-semibold text-gray-900 mb-2">成员</h3>
      <div className="flex gap-2 mb-3">
        <select
          value={selectedUserId}
          onChange={(e) => setSelectedUserId(e.target.value)}
          className="flex-1 rounded border border-gray-300 px-2 py-1.5 text-sm"
        >
          <option value="">选择用户...</option>
          {availableUsers.map((u: { id: number; username: string }) => (
            <option key={u.id} value={u.id}>{u.username}</option>
          ))}
        </select>
        <button
          onClick={() => {
            if (!selectedUserId) return;
            addMembers.mutate([Number(selectedUserId)], {
              onSuccess: () => { setSelectedUserId(""); showToast("成员已添加", "success"); },
            });
          }}
          disabled={!selectedUserId || addMembers.isPending}
          className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
        >
          添加
        </button>
      </div>
      {isLoading ? (
        <p className="text-sm text-gray-500">加载中...</p>
      ) : members.length === 0 ? (
        <p className="text-sm text-gray-400">暂无成员</p>
      ) : (
        <ul className="space-y-1">
          {members.map((m) => (
            <li key={m.id} className="flex items-center justify-between rounded bg-white px-3 py-2 text-sm">
              <span>{m.user?.username || `用户 #${m.user_id}`}</span>
              <button
                onClick={() => removeMember.mutate(m.user_id)}
                className="text-xs text-red-500 hover:text-red-700"
              >
                移除
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

function PermissionsSection({ groupId }: { groupId: number }) {
  const { data: perms = [], isLoading } = useGroupPermissions(groupId);
  const addPerm = useAddGroupPermission(groupId);
  const updatePerm = useUpdateGroupPermission(groupId);
  const removePerm = useRemoveGroupPermission(groupId);
  const [serviceId, setServiceId] = useState("");
  const [role, setRole] = useState("viewer");

  const { data: servicesData } = useQuery({
    queryKey: ["services-for-group"],
    queryFn: () => apiClient.get("/services").then((r) => r.data),
  });
  const allServices = servicesData?.items ?? [];
  const existingServiceIds = new Set(perms.map((p) => p.service_id));
  const availableServices = allServices.filter((s: { id: number }) => !existingServiceIds.has(s.id));

  return (
    <div>
      <h3 className="text-sm font-semibold text-gray-900 mb-2">Service 权限</h3>
      <div className="flex gap-2 mb-3">
        <select
          value={serviceId}
          onChange={(e) => setServiceId(e.target.value)}
          className="flex-1 rounded border border-gray-300 px-2 py-1.5 text-sm"
        >
          <option value="">选择 Service...</option>
          {availableServices.map((s: { id: number; name: string }) => (
            <option key={s.id} value={s.id}>{s.name}</option>
          ))}
        </select>
        <select value={role} onChange={(e) => setRole(e.target.value)} className="rounded border border-gray-300 px-2 py-1.5 text-sm">
          {ROLES.map((r) => <option key={r.value} value={r.value}>{r.label}</option>)}
        </select>
        <button
          onClick={() => {
            if (!serviceId) return;
            addPerm.mutate({ service_id: Number(serviceId), role }, {
              onSuccess: () => { setServiceId(""); showToast("权限已添加", "success"); },
            });
          }}
          disabled={!serviceId || addPerm.isPending}
          className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-50"
        >
          添加
        </button>
      </div>
      {isLoading ? (
        <p className="text-sm text-gray-500">加载中...</p>
      ) : perms.length === 0 ? (
        <p className="text-sm text-gray-400">暂无权限配置</p>
      ) : (
        <ul className="space-y-1">
          {perms.map((p) => (
            <li key={p.id} className="flex items-center justify-between rounded bg-white px-3 py-2 text-sm">
              <span>{p.service?.name || `Service #${p.service_id}`}</span>
              <div className="flex items-center gap-2">
                <select
                  value={p.role}
                  onChange={(e) => updatePerm.mutate({ permId: p.id, role: e.target.value })}
                  className="rounded border border-gray-300 px-1.5 py-1 text-xs"
                >
                  {ROLES.map((r) => <option key={r.value} value={r.value}>{r.label}</option>)}
                </select>
                <button
                  onClick={() => removePerm.mutate(p.id)}
                  className="text-xs text-red-500 hover:text-red-700"
                >
                  移除
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
