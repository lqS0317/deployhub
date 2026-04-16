"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { ClusterNamespace } from "@/types";

export function useClusterNamespaces(clusterId: number) {
  return useQuery({
    queryKey: ["cluster-namespaces", clusterId],
    queryFn: async () => {
      const res = await apiClient.get<{ items: ClusterNamespace[] }>(`/clusters/${clusterId}/namespaces`);
      return res.data.items ?? [];
    },
    enabled: clusterId > 0,
  });
}

export function useAddClusterNamespace(clusterId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: { namespace: string; is_default?: boolean }) => {
      const res = await apiClient.post(`/clusters/${clusterId}/namespaces`, data);
      return res.data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["cluster-namespaces", clusterId] }),
  });
}

export function useDeleteClusterNamespace(clusterId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (nsId: number) => {
      await apiClient.delete(`/clusters/${clusterId}/namespaces/${nsId}`);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["cluster-namespaces", clusterId] }),
  });
}

export function useSyncClusterNamespaces(clusterId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const res = await apiClient.post(`/clusters/${clusterId}/namespaces/sync`);
      return res.data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["cluster-namespaces", clusterId] }),
  });
}
