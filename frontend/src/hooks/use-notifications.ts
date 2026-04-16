"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { Notification, PaginatedResponse } from "@/types";

// queryKey 常量
const NOTIF_KEYS = {
  all: ["notifications"] as const,
  list: (params?: Record<string, unknown>) =>
    ["notifications", "list", params] as const,
  unreadCount: ["notifications", "unreadCount"] as const,
};

interface ListNotificationsParams {
  page?: number;
  page_size?: number;
}

/** 获取通知列表 */
export function useNotifications(params?: ListNotificationsParams) {
  return useQuery({
    queryKey: NOTIF_KEYS.list(params as Record<string, unknown>),
    queryFn: async () => {
      const res = await apiClient.get<PaginatedResponse<Notification>>(
        "/notifications",
        { params }
      );
      return res.data;
    },
  });
}

/** 获取未读通知数量 */
export function useUnreadNotificationCount() {
  return useQuery({
    queryKey: NOTIF_KEYS.unreadCount,
    queryFn: async () => {
      const res = await apiClient.get<{ count: number }>(
        "/notifications/unread-count"
      );
      return res.data;
    },
    refetchInterval: 30000,
  });
}

/** 标记单个通知为已读 */
export function useMarkNotificationRead(id: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      await apiClient.put(`/notifications/${id}/read`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: NOTIF_KEYS.all });
      queryClient.invalidateQueries({
        queryKey: NOTIF_KEYS.unreadCount,
      });
    },
  });
}

/** 标记所有通知为已读 */
export function useMarkAllNotificationsRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      await apiClient.put("/notifications/read-all");
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: NOTIF_KEYS.all });
      queryClient.invalidateQueries({
        queryKey: NOTIF_KEYS.unreadCount,
      });
    },
  });
}
