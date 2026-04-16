"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { Build, PaginatedResponse } from "@/types";

// queryKey 常量
const BUILD_KEYS = {
  all: ["builds"] as const,
  list: (params?: Record<string, unknown>) =>
    ["builds", "list", params] as const,
  detail: (id: number) => ["builds", "detail", id] as const,
  log: (id: number) => ["builds", "log", id] as const,
};

interface ListBuildsParams {
  service_id?: number | string;
  page?: number;
  page_size?: number;
}

/** 获取构建列表 */
export function useBuilds(params?: ListBuildsParams) {
  return useQuery({
    queryKey: BUILD_KEYS.list(params as Record<string, unknown>),
    queryFn: async () => {
      const res = await apiClient.get<PaginatedResponse<Build>>("/builds", {
        params,
      });
      return res.data;
    },
  });
}

/** 获取单个构建详情 */
export function useBuild(id: number) {
  return useQuery({
    queryKey: BUILD_KEYS.detail(id),
    queryFn: async () => {
      const res = await apiClient.get<Build>(`/builds/${id}`);
      return res.data;
    },
    enabled: id > 0,
  });
}

/** 触发构建 */
export function useTriggerBuild() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      service_id: number;
      git_branch: string;
      git_commit?: string;
      image_tag?: string;
      name?: string;
      dockerfile_path?: string;
      registry_id?: number;
      image_repo?: string;
      build_context?: string;
      build_cluster_id?: number;
    }) => {
      const res = await apiClient.post<Build>("/builds", data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: BUILD_KEYS.all });
    },
  });
}

/** 取消构建 */
export function useCancelBuild() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: number | string) => {
      await apiClient.post(`/builds/${id}/cancel`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: BUILD_KEYS.all });
    },
  });
}

/** 删除构建 */
export function useDeleteBuild() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/builds/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: BUILD_KEYS.all });
    },
  });
}

/** 重新构建 */
export function useRetryBuild() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: number) => {
      const res = await apiClient.post<Build>(`/builds/${id}/retry`);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: BUILD_KEYS.all });
    },
  });
}

/** 获取构建日志 */
export function useBuildLog(id: number) {
  return useQuery({
    queryKey: BUILD_KEYS.log(id),
    queryFn: async () => {
      const res = await apiClient.get<{ log: string }>(`/builds/${id}/log`);
      return res.data;
    },
    enabled: id > 0,
  });
}
