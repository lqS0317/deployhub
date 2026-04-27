"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { ClusterNamespace } from "@/types";

interface ClusterNamespacesResponse {
  items: ClusterNamespace[];
}

export function useClusterNamespaces(clusterId?: number) {
  return useQuery({
    queryKey: ["cluster-namespaces", clusterId],
    queryFn: async () => {
      const res = await apiClient.get<ClusterNamespacesResponse>(`/clusters/${clusterId}/namespaces`);
      return res.data;
    },
    enabled: typeof clusterId === "number" && clusterId > 0,
    staleTime: 30_000,
  });
}
