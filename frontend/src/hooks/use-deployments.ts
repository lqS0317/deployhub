"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { Deployment, PaginatedResponse } from "@/types";

// queryKey 常量
const DEPLOY_KEYS = {
  all: ["deployments"] as const,
  list: (params?: Record<string, unknown>) =>
    ["deployments", "list", params] as const,
  detail: (id: number) => ["deployments", "detail", id] as const,
};

interface ListDeploymentsParams {
  service_id?: number | string;
  page?: number;
  page_size?: number;
}

const ACTIVE_STATUSES = new Set([
  "previewing", "deploying", "pod_checking", "pending_approval",
]);

/** 获取部署列表 */
export function useDeployments(params?: ListDeploymentsParams) {
  return useQuery({
    queryKey: DEPLOY_KEYS.list(params as Record<string, unknown>),
    queryFn: async () => {
      const res = await apiClient.get<PaginatedResponse<Deployment>>(
        "/deployments",
        { params }
      );
      return res.data;
    },
    refetchInterval: 5000,
  });
}

/** 获取单个部署详情，进行中状态自动 3 秒轮询 */
export function useDeployment(id: number | string) {
  const numId = Number(id);
  const query = useQuery({
    queryKey: DEPLOY_KEYS.detail(numId),
    queryFn: async () => {
      const res = await apiClient.get<Deployment>(`/deployments/${numId}`);
      return res.data;
    },
    enabled: numId > 0,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      return status && ACTIVE_STATUSES.has(status) ? 3000 : false;
    },
  });
  return query;
}

/** 创建部署（发起发布） */
export function useCreateDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      service_id: number;
      cluster_id: number;
      namespace: string;
      build_id?: number;
      image_tag?: string;
      replicas?: number;
      image_source?: string;
      external_image?: string;
      deploy_type?: string;
      workload_type?: string;
      port?: number;
      helm_repo_id?: number;
      helm_chart_path?: string;
      helm_release_name?: string;
      helm_chart_branch?: string;
      helm_service_account?: string;
      direct_mode?: string;
      raw_yaml?: string;
    }) => {
      const res = await apiClient.post<Deployment>("/deployments", data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DEPLOY_KEYS.all });
    },
  });
}

/** 触发 dry-run 预览 */
export function usePreviewDeploy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      const res = await apiClient.post(`/deployments/${id}/preview`);
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: DEPLOY_KEYS.all }),
  });
}

/** 获取预览结果 */
export function useDeployPreview(id: number) {
  return useQuery({
    queryKey: ["deploy-preview", id],
    queryFn: async () => {
      const res = await apiClient.get(`/deployments/${id}/preview`);
      return res.data;
    },
    enabled: id > 0,
  });
}

/** 取消部署 */
export function useCancelDeploy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      await apiClient.post(`/deployments/${id}/cancel`);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: DEPLOY_KEYS.all }),
  });
}

/** 删除部署 */
export function useDeleteDeploy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/deployments/${id}`);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: DEPLOY_KEYS.all }),
  });
}

/** 执行部署（预览确认后） */
export function useExecuteDeploy() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      const res = await apiClient.post(`/deployments/${id}/execute`);
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: DEPLOY_KEYS.all }),
  });
}

/** 回滚部署 */
export function useRollback(id: number | string) {
  const queryClient = useQueryClient();
  const numId = Number(id);

  return useMutation({
    mutationFn: async () => {
      const res = await apiClient.post<Deployment>(
        `/deployments/${numId}/rollback`
      );
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DEPLOY_KEYS.all });
    },
  });
}

/** 回滚部署（无需预设 id，在 mutate 时传入） */
export function useRollbackDeployment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { deploymentId: number | string }) => {
      const res = await apiClient.post<Deployment>(
        `/deployments/${data.deploymentId}/rollback`
      );
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: DEPLOY_KEYS.all });
    },
  });
}
