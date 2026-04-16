"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { Approval, PaginatedResponse } from "@/types";

// queryKey 常量
const APPROVAL_KEYS = {
  all: ["approvals"] as const,
  list: (params?: Record<string, unknown>) =>
    ["approvals", "list", params] as const,
  detail: (id: number) => ["approvals", "detail", id] as const,
};

interface ListApprovalsParams {
  status?: string;
  page?: number;
  page_size?: number;
}

/** 获取审批列表 */
export function useApprovals(params?: ListApprovalsParams) {
  return useQuery({
    queryKey: APPROVAL_KEYS.list(params as Record<string, unknown>),
    queryFn: async () => {
      const res = await apiClient.get<PaginatedResponse<Approval>>(
        "/approvals",
        { params }
      );
      return res.data;
    },
  });
}

/** 获取单个审批详情 */
export function useApproval(id: number) {
  return useQuery({
    queryKey: APPROVAL_KEYS.detail(id),
    queryFn: async () => {
      const res = await apiClient.get<Approval>(`/approvals/${id}`);
      return res.data;
    },
    enabled: id > 0,
  });
}

/** 通过审批 */
export function useApproveDeployment(id: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data?: { comment?: string }) => {
      const res = await apiClient.post<Approval>(
        `/approvals/${id}/approve`,
        data
      );
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: APPROVAL_KEYS.all });
      queryClient.invalidateQueries({
        queryKey: APPROVAL_KEYS.detail(id),
      });
      // 审批通过后也刷新部署列表
      queryClient.invalidateQueries({ queryKey: ["deployments"] });
    },
  });
}

/** 拒绝审批 */
export function useRejectDeployment(id: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { comment: string }) => {
      const res = await apiClient.post<Approval>(
        `/approvals/${id}/reject`,
        data
      );
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: APPROVAL_KEYS.all });
      queryClient.invalidateQueries({
        queryKey: APPROVAL_KEYS.detail(id),
      });
      queryClient.invalidateQueries({ queryKey: ["deployments"] });
    },
  });
}
