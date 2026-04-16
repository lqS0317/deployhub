"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { HelmValues } from "@/types";

export function useHelmValues(serviceId: number) {
  return useQuery({
    queryKey: ["helm-values", serviceId],
    queryFn: async () => {
      const res = await apiClient.get<{ items: HelmValues[] }>(`/services/${serviceId}/helm-values`);
      return res.data.items ?? [];
    },
    enabled: serviceId > 0,
  });
}

export function useUpdateHelmValues(serviceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ clusterId, content }: { clusterId: number; content: string }) => {
      const res = await apiClient.put(`/services/${serviceId}/helm-values/${clusterId}`, { content });
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["helm-values", serviceId] });
    },
  });
}

export function useExecuteDeploy() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (deploymentId: number) => {
      const res = await apiClient.post(`/deployments/${deploymentId}/execute`);
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["deployments"] });
    },
  });
}
