"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type {
  ConfigTemplate,
  ConfigEnvValue,
  ConfigVersion,
  ConfigDeployment,
} from "@/types";

// queryKey 常量
const CONFIG_KEYS = {
  all: ["configs"] as const,
  templates: (serviceId: number) =>
    ["configs", "templates", serviceId] as const,
  template: (id: number) => ["configs", "template", id] as const,
  envValues: (templateId: number) =>
    ["configs", "envValues", templateId] as const,
  versions: (templateId: number) =>
    ["configs", "versions", templateId] as const,
};

/** 获取服务的配置模板列表 */
export function useConfigTemplates(serviceId: number) {
  return useQuery({
    queryKey: CONFIG_KEYS.templates(serviceId),
    queryFn: async () => {
      const res = await apiClient.get<ConfigTemplate[]>(
        `/services/${serviceId}/configs`
      );
      return res.data;
    },
    enabled: serviceId > 0,
  });
}

/** 创建配置模板 */
export function useCreateConfigTemplate(serviceId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      name: string;
      config_type: string;
      template_content: string;
    }) => {
      const res = await apiClient.post<ConfigTemplate>(
        `/services/${serviceId}/configs`,
        data
      );
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: CONFIG_KEYS.templates(serviceId),
      });
    },
  });
}

/** 获取配置模板详情 */
export function useConfigTemplate(id: number) {
  return useQuery({
    queryKey: CONFIG_KEYS.template(id),
    queryFn: async () => {
      const res = await apiClient.get<ConfigTemplate>(`/configs/${id}`);
      return res.data;
    },
    enabled: id > 0,
  });
}

/** 更新配置模板 */
export function useUpdateConfigTemplate(id: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: Partial<ConfigTemplate>) => {
      const res = await apiClient.put<ConfigTemplate>(
        `/configs/${id}`,
        data
      );
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CONFIG_KEYS.all });
      queryClient.invalidateQueries({ queryKey: CONFIG_KEYS.template(id) });
    },
  });
}

/** 删除配置模板 */
export function useDeleteConfigTemplate(id: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async () => {
      await apiClient.delete(`/configs/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CONFIG_KEYS.all });
    },
  });
}

/** 获取配置模板的环境变量值列表 */
export function useConfigEnvValues(templateId: number) {
  return useQuery({
    queryKey: CONFIG_KEYS.envValues(templateId),
    queryFn: async () => {
      const res = await apiClient.get<ConfigEnvValue[]>(
        `/configs/${templateId}/env-values`
      );
      return res.data;
    },
    enabled: templateId > 0,
  });
}

/** 设置某环境的变量值 */
export function useSetConfigEnvValues(templateId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      cluster_id: number;
      vars: Record<string, string>;
    }) => {
      const res = await apiClient.put<ConfigEnvValue>(
        `/configs/${templateId}/env-values/${data.cluster_id}`,
        { vars: data.vars }
      );
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: CONFIG_KEYS.envValues(templateId),
      });
    },
  });
}

/** 预览渲染结果 */
export function useRenderConfigPreview(templateId: number) {
  return useMutation({
    mutationFn: async (data: { cluster_id: number }) => {
      const res = await apiClient.post<{ rendered_content: string }>(
        `/configs/${templateId}/render`,
        data
      );
      return res.data;
    },
  });
}

/** 获取配置版本列表 */
export function useConfigVersions(templateId: number) {
  return useQuery({
    queryKey: CONFIG_KEYS.versions(templateId),
    queryFn: async () => {
      const res = await apiClient.get<ConfigVersion[]>(
        `/configs/${templateId}/versions`
      );
      return res.data;
    },
    enabled: templateId > 0,
  });
}

/** 版本 Diff 对比 */
export function useConfigVersionDiff(
  templateId: number,
  versionId: number
) {
  return useQuery({
    queryKey: ["configs", "diff", templateId, versionId],
    queryFn: async () => {
      const res = await apiClient.get<{
        old_content: string;
        new_content: string;
      }>(`/configs/${templateId}/versions/${versionId}/diff`);
      return res.data;
    },
    enabled: templateId > 0 && versionId > 0,
  });
}

/** 下发配置到集群 */
export function useDeployConfig(templateId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: {
      cluster_id: number;
      namespace: string;
      resource_name: string;
    }) => {
      const res = await apiClient.post<ConfigDeployment>(
        `/configs/${templateId}/deploy`,
        data
      );
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: CONFIG_KEYS.versions(templateId),
      });
    },
  });
}
