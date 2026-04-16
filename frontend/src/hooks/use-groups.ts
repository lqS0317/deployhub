"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { Group, GroupMember, GroupServicePermission, PaginatedResponse } from "@/types";

const GROUP_KEYS = {
  all: ["groups"] as const,
  detail: (id: number) => ["groups", id] as const,
  members: (id: number) => ["groups", id, "members"] as const,
  permissions: (id: number) => ["groups", id, "permissions"] as const,
};

export function useGroups() {
  return useQuery({
    queryKey: GROUP_KEYS.all,
    queryFn: async () => {
      const res = await apiClient.get<PaginatedResponse<Group>>("/groups");
      return res.data;
    },
  });
}

export function useGroup(id: number) {
  return useQuery({
    queryKey: GROUP_KEYS.detail(id),
    queryFn: async () => {
      const res = await apiClient.get<Group>(`/groups/${id}`);
      return res.data;
    },
    enabled: id > 0,
  });
}

export function useCreateGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: { name: string; description: string }) => {
      const res = await apiClient.post<Group>("/groups", data);
      return res.data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: GROUP_KEYS.all }),
  });
}

export function useUpdateGroup(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: { name: string; description: string }) => {
      const res = await apiClient.put<Group>(`/groups/${id}`, data);
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: GROUP_KEYS.all });
      qc.invalidateQueries({ queryKey: GROUP_KEYS.detail(id) });
    },
  });
}

export function useDeleteGroup() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/groups/${id}`);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: GROUP_KEYS.all }),
  });
}

export function useGroupMembers(id: number) {
  return useQuery({
    queryKey: GROUP_KEYS.members(id),
    queryFn: async () => {
      const res = await apiClient.get<{ items: GroupMember[] }>(`/groups/${id}/members`);
      return res.data.items ?? [];
    },
    enabled: id > 0,
  });
}

export function useAddGroupMembers(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (userIds: number[]) => {
      const res = await apiClient.post(`/groups/${id}/members`, { user_ids: userIds });
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: GROUP_KEYS.members(id) });
      qc.invalidateQueries({ queryKey: GROUP_KEYS.all });
    },
  });
}

export function useRemoveGroupMember(groupId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (userId: number) => {
      await apiClient.delete(`/groups/${groupId}/members/${userId}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: GROUP_KEYS.members(groupId) });
      qc.invalidateQueries({ queryKey: GROUP_KEYS.all });
    },
  });
}

export function useGroupPermissions(id: number) {
  return useQuery({
    queryKey: GROUP_KEYS.permissions(id),
    queryFn: async () => {
      const res = await apiClient.get<{ items: GroupServicePermission[] }>(`/groups/${id}/permissions`);
      return res.data.items ?? [];
    },
    enabled: id > 0,
  });
}

export function useAddGroupPermission(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: { service_id: number; role: string }) => {
      const res = await apiClient.post(`/groups/${id}/permissions`, data);
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: GROUP_KEYS.permissions(id) });
      qc.invalidateQueries({ queryKey: GROUP_KEYS.all });
    },
  });
}

export function useUpdateGroupPermission(groupId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ permId, role }: { permId: number; role: string }) => {
      await apiClient.put(`/groups/${groupId}/permissions/${permId}`, { role });
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: GROUP_KEYS.permissions(groupId) }),
  });
}

export function useRemoveGroupPermission(groupId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (permId: number) => {
      await apiClient.delete(`/groups/${groupId}/permissions/${permId}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: GROUP_KEYS.permissions(groupId) });
      qc.invalidateQueries({ queryKey: GROUP_KEYS.all });
    },
  });
}
