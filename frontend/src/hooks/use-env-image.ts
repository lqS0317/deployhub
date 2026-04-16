"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";

interface EnvImageInfo {
  repository: string;
  tag: string;
  image_pull_policy: string;
  full_image: string;
}

export function useEnvImage(serviceId: number, enabled: boolean = false) {
  return useQuery({
    queryKey: ["env-image", serviceId],
    queryFn: async () => {
      const res = await apiClient.get<EnvImageInfo>(`/services/${serviceId}/env-image`);
      return res.data;
    },
    enabled: enabled && serviceId > 0,
  });
}
