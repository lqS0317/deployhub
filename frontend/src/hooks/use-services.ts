"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { Service, ServiceMember, PaginatedResponse } from "@/types";

// queryKey 常量
const SERVICE_KEYS = {
  all: ["services"] as const,
  list: (params?: Record<string, unknown>) =>
    ["services", "list", params] as const,
  detail: (id: number) => ["services", "detail", id] as const,
  members: (serviceId: number) =>
    ["services", "members", serviceId] as const,
};

interface ListServicesParams {
  search?: string;
  page?: number;
  page_size?: number;
}

/** 获取服务列表 */
export function useServices(params?: ListServicesParams) {
  return useQuery({
    queryKey: SERVICE_KEYS.list(params as Record<string, unknown>),
    queryFn: async () => {
      const res = await apiClient.get<PaginatedResponse<Service>>(
        "/services",
        { params }
      );
      return res.data;
    },
  });
}

/** 获取单个服务详情 */
export function useService(id: number | string) {
  const numId = Number(id);
  return useQuery({
    queryKey: SERVICE_KEYS.detail(numId),
    queryFn: async () => {
      const res = await apiClient.get<Service>(`/services/${numId}`);
      return res.data;
    },
    enabled: numId > 0,
  });
}

/** 创建服务 */
export function useCreateService() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Partial<Service>) => {
      const res = await apiClient.post<Service>("/services", data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_KEYS.all });
    },
  });
}

/** 更新服务 */
export function useUpdateService(id: number | string) {
  const queryClient = useQueryClient();
  const numId = Number(id);

  return useMutation({
    mutationFn: async (data: Partial<Service>) => {
      const res = await apiClient.put<Service>(`/services/${numId}`, data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_KEYS.all });
      queryClient.invalidateQueries({ queryKey: SERVICE_KEYS.detail(numId) });
    },
  });
}

/** 删除服务 */
export function useDeleteService() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: number | string) => {
      await apiClient.delete(`/services/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_KEYS.all });
    },
  });
}

/** 获取服务成员列表 */
export function useServiceMembers(serviceId: number | string) {
  const sid = Number(serviceId);
  return useQuery({
    queryKey: SERVICE_KEYS.members(sid),
    queryFn: async () => {
      const res = await apiClient.get<ServiceMember[]>(
        `/services/${sid}/members`
      );
      return res.data;
    },
    enabled: sid > 0,
  });
}

/** 添加服务成员 */
export function useAddServiceMember(serviceId: number | string) {
  const queryClient = useQueryClient();
  const sid = Number(serviceId);

  return useMutation({
    mutationFn: async (data: { user_id: number; role: string }) => {
      const res = await apiClient.post<ServiceMember>(
        `/services/${sid}/members`,
        data
      );
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: SERVICE_KEYS.members(sid),
      });
    },
  });
}

/** useAddMember 别名 */
export const useAddMember = useAddServiceMember;

/** 移除服务成员 */
export function useRemoveMember(serviceId: number | string) {
  const queryClient = useQueryClient();
  const sid = Number(serviceId);

  return useMutation({
    mutationFn: async (memberId: number | string) => {
      await apiClient.delete(`/services/${sid}/members/${memberId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: SERVICE_KEYS.members(sid),
      });
    },
  });
}
