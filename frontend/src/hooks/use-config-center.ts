"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type {
  ConfigEntry,
  ConfigItem,
  ConfigRelease,
} from "@/types";

const CONFIG_KEYS = {
  entries: (serviceId: number, clusterId: number) =>
    ["config-entries", serviceId, clusterId] as const,
  items: (entryId: number) => ["config-items", entryId] as const,
  draft: (entryId: number) => ["config-draft", entryId] as const,
  releases: (entryId: number) => ["config-releases", entryId] as const,
};

// ---- ConfigEntry (多条目) ----

export function useConfigEntries(serviceId: number, clusterId: number) {
  return useQuery({
    queryKey: CONFIG_KEYS.entries(serviceId, clusterId),
    queryFn: async () => {
      const res = await apiClient.get<ConfigEntry[]>(
        `/services/${serviceId}/config-entries`,
        { params: { cluster_id: clusterId } }
      );
      return Array.isArray(res.data) ? res.data : [];
    },
    enabled: serviceId > 0 && clusterId > 0,
  });
}

export function useCreateEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      serviceId: number;
      cluster_id: number;
      name: string;
      config_type: string;
      format: string;
      mount_path?: string;
    }) => {
      const { serviceId, ...body } = data;
      const res = await apiClient.post<ConfigEntry>(
        `/services/${serviceId}/config-entries`,
        body
      );
      return res.data;
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({
        queryKey: CONFIG_KEYS.entries(vars.serviceId, vars.cluster_id),
      });
    },
  });
}

export function useUpdateEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      entryId: number;
      name?: string;
      config_type?: string;
      format?: string;
    }) => {
      const { entryId, ...body } = data;
      const res = await apiClient.put<ConfigEntry>(
        `/config-entries/${entryId}`,
        body
      );
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["config-entries"] });
    },
  });
}

export function useDeleteEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (entryId: number) => {
      await apiClient.delete(`/config-entries/${entryId}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["config-entries"] });
    },
  });
}

// ---- Items ----

export function useConfigItems(entryId: number) {
  return useQuery({
    queryKey: CONFIG_KEYS.items(entryId),
    queryFn: async () => {
      const res = await apiClient.get<ConfigItem[]>(
        `/config-entries/${entryId}/items`
      );
      return Array.isArray(res.data) ? res.data : [];
    },
    enabled: entryId > 0,
  });
}

export function useCreateItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      entryId: number;
      key: string;
      value: string;
      comment?: string;
    }) => {
      const { entryId, ...body } = data;
      const res = await apiClient.post<ConfigItem>(
        `/config-entries/${entryId}/items`,
        body
      );
      return res.data;
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: CONFIG_KEYS.items(vars.entryId) });
    },
  });
}

export function useUpdateItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      entryId: number;
      itemId: number;
      value: string;
      comment?: string;
    }) => {
      const { entryId, itemId, ...body } = data;
      const res = await apiClient.put<ConfigItem>(
        `/config-entries/${entryId}/items/${itemId}`,
        body
      );
      return res.data;
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: CONFIG_KEYS.items(vars.entryId) });
    },
  });
}

export function useDeleteItem() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: { entryId: number; itemId: number }) => {
      await apiClient.delete(
        `/config-entries/${data.entryId}/items/${data.itemId}`
      );
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: CONFIG_KEYS.items(vars.entryId) });
    },
  });
}

// ---- Draft ----

export function useConfigDraft(entryId: number) {
  return useQuery({
    queryKey: CONFIG_KEYS.draft(entryId),
    queryFn: async () => {
      const res = await apiClient.get<{ content: string }>(
        `/config-entries/${entryId}/draft`
      );
      return res.data?.content ?? "";
    },
    enabled: entryId > 0,
  });
}

export function useSaveDraft() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: { entryId: number; content: string }) => {
      await apiClient.put(`/config-entries/${data.entryId}/draft`, {
        content: data.content,
      });
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: CONFIG_KEYS.draft(vars.entryId) });
    },
  });
}

// ---- Publish / Rollback ----

export function usePublish() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: { entryId: number; comment?: string }) => {
      const res = await apiClient.post<ConfigRelease>(
        `/config-entries/${data.entryId}/release`,
        { comment: data.comment ?? "" }
      );
      return res.data;
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({
        queryKey: CONFIG_KEYS.releases(vars.entryId),
      });
      qc.invalidateQueries({
        queryKey: CONFIG_KEYS.items(vars.entryId),
      });
      qc.invalidateQueries({
        queryKey: CONFIG_KEYS.draft(vars.entryId),
      });
    },
  });
}

export function useRollbackConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (data: {
      entryId: number;
      target_version: number;
      comment?: string;
    }) => {
      const res = await apiClient.post<ConfigRelease>(
        `/config-entries/${data.entryId}/rollback`,
        { target_version: data.target_version, comment: data.comment ?? "" }
      );
      return res.data;
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({
        queryKey: CONFIG_KEYS.releases(vars.entryId),
      });
      qc.invalidateQueries({
        queryKey: CONFIG_KEYS.items(vars.entryId),
      });
      qc.invalidateQueries({
        queryKey: CONFIG_KEYS.draft(vars.entryId),
      });
    },
  });
}

// ---- Releases ----

export function useConfigReleases(entryId: number) {
  return useQuery({
    queryKey: CONFIG_KEYS.releases(entryId),
    queryFn: async () => {
      const res = await apiClient.get<ConfigRelease[]>(
        `/config-entries/${entryId}/releases`
      );
      return Array.isArray(res.data) ? res.data : [];
    },
    enabled: entryId > 0,
  });
}

