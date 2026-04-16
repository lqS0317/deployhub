"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { RoutePlugin, PluginDeployment } from "@/types";

const PLUGIN_KEYS = {
  all: ["route-plugins"] as const,
  list: () => ["route-plugins", "list"] as const,
  deployments: (pluginId: number) =>
    ["route-plugins", "deployments", pluginId] as const,
};

export function usePlugins() {
  return useQuery({
    queryKey: PLUGIN_KEYS.list(),
    queryFn: async () => {
      const res = await apiClient.get("/route-plugins");
      const data = res.data;
      const items = Array.isArray(data) ? data : (data as { items?: RoutePlugin[] })?.items ?? [];
      return items as RoutePlugin[];
    },
  });
}

export function useCreatePlugin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      name: string;
      description?: string;
      yaml_content: string;
    }) => {
      const res = await apiClient.post<RoutePlugin>("/route-plugins", data);
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: PLUGIN_KEYS.all });
    },
  });
}

export function useUpdatePlugin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      id: number;
      name: string;
      description?: string;
      yaml_content: string;
    }) => {
      const { id, ...body } = data;
      const res = await apiClient.put<RoutePlugin>(
        `/route-plugins/${id}`,
        body
      );
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: PLUGIN_KEYS.all });
    },
  });
}

export function useDeletePlugin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      await apiClient.delete(`/route-plugins/${id}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: PLUGIN_KEYS.all });
    },
  });
}

export function useDeployPlugin() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      id: number;
      cluster_id: number;
      namespace: string;
    }) => {
      const { id, ...body } = data;
      const res = await apiClient.post<PluginDeployment>(
        `/route-plugins/${id}/deploy`,
        body
      );
      return res.data;
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: PLUGIN_KEYS.all });
      qc.invalidateQueries({
        queryKey: PLUGIN_KEYS.deployments(vars.id),
      });
    },
  });
}

export function usePluginDeployments(pluginId: number) {
  return useQuery({
    queryKey: PLUGIN_KEYS.deployments(pluginId),
    queryFn: async () => {
      const res = await apiClient.get(
        `/route-plugins/${pluginId}/deployments`
      );
      const data = res.data;
      const items = Array.isArray(data) ? data : (data as { items?: PluginDeployment[] })?.items ?? [];
      return items as PluginDeployment[];
    },
    enabled: pluginId > 0,
  });
}
