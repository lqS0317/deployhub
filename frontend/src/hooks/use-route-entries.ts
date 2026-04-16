"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { RouteEntry, RouteDeployment } from "@/types";

const ROUTE_KEYS = {
  all: ["route-entries"] as const,
  list: (resourceType?: string) =>
    ["route-entries", "list", resourceType] as const,
  deployments: (entryId: number) =>
    ["route-entries", "deployments", entryId] as const,
  preview: (entryId: number, namespace: string) =>
    ["route-entries", "preview", entryId, namespace] as const,
};

export function useRouteEntries(resourceType?: string) {
  return useQuery({
    queryKey: ROUTE_KEYS.list(resourceType),
    queryFn: async () => {
      const res = await apiClient.get("/route-entries", {
        params: resourceType ? { resource_type: resourceType } : undefined,
      });
      const data = res.data;
      const items = Array.isArray(data) ? data : (data as { items?: RouteEntry[] })?.items ?? [];
      return items as RouteEntry[];
    },
  });
}

export function useCreateRouteEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      name: string;
      resource_type: string;
      config: unknown;
    }) => {
      const res = await apiClient.post<RouteEntry>("/route-entries", data);
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ROUTE_KEYS.all });
    },
  });
}

export function useUpdateRouteEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      id: number;
      name: string;
      resource_type: string;
      config: unknown;
    }) => {
      const { id, ...body } = data;
      const res = await apiClient.put<RouteEntry>(
        `/route-entries/${id}`,
        body
      );
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ROUTE_KEYS.all });
    },
  });
}

export function useDeleteRouteEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/route-entries/${id}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ROUTE_KEYS.all });
    },
  });
}

export function useDeployRouteEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      id: number;
      cluster_id: number;
      namespace: string;
    }) => {
      const { id, ...body } = data;
      const res = await apiClient.post<RouteDeployment>(
        `/route-entries/${id}/deploy`,
        body
      );
      return res.data;
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ROUTE_KEYS.all });
      qc.invalidateQueries({
        queryKey: ROUTE_KEYS.deployments(vars.id),
      });
    },
  });
}

export function usePreviewRouteEntry(entryId: number, namespace: string) {
  return useQuery({
    queryKey: ROUTE_KEYS.preview(entryId, namespace),
    queryFn: async () => {
      const res = await apiClient.get<{ yaml: string }>(
        `/route-entries/${entryId}/preview`,
        { params: { namespace } }
      );
      return res.data;
    },
    enabled: entryId > 0 && namespace.length > 0,
  });
}

export function useEntryDeployments(entryId: number) {
  return useQuery({
    queryKey: ROUTE_KEYS.deployments(entryId),
    queryFn: async () => {
      const res = await apiClient.get(
        `/route-entries/${entryId}/deployments`
      );
      const data = res.data;
      const items = Array.isArray(data) ? data : (data as { items?: RouteDeployment[] })?.items ?? [];
      return items as RouteDeployment[];
    },
    enabled: entryId > 0,
  });
}
