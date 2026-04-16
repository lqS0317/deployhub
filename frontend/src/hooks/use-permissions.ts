"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { EffectivePermission } from "@/types";

export function useUserPermissions(userId: number) {
  return useQuery({
    queryKey: ["permissions", "user", userId],
    queryFn: async () => {
      const res = await apiClient.get<{ items: EffectivePermission[] }>(`/users/${userId}/permissions`);
      return res.data.items ?? [];
    },
    enabled: userId > 0,
  });
}

export function useMyPermissions() {
  return useQuery({
    queryKey: ["permissions", "my"],
    queryFn: async () => {
      const res = await apiClient.get<{ items: EffectivePermission[] }>("/auth/my-permissions");
      return res.data.items ?? [];
    },
  });
}
