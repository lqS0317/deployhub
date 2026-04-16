"use client";

import { useState, useEffect } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { showToast } from "@/components/ui/toast";

interface SystemSetting {
  key: string;
  value: string;
  description: string;
  updated_at: string;
}

const SETTING_LABELS: Record<string, { label: string; placeholder: string; help: string }> = {
  helm_job_namespace: {
    label: "Helm Job 命名空间",
    placeholder: "deployhub-jobs",
    help: "Helm Runner Job 运行的命名空间，与服务部署的命名空间分离。修改后下次部署立即生效，无需重启服务。",
  },
  env_values_map: {
    label: "环境 Values 映射",
    placeholder: "qanet:qa,testnet:testnet,mainnet:mainnet",
    help: "集群环境到 Helm values 文件后缀的映射。格式: env1:suffix1,env2:suffix2。例如 qanet:qa 表示 qanet 环境使用 app-qa.yaml。",
  },
};

export default function SystemSettingsPage() {
  const queryClient = useQueryClient();
  const [editValues, setEditValues] = useState<Record<string, string>>({});
  const [dirty, setDirty] = useState<Record<string, boolean>>({});

  const { data, isLoading } = useQuery({
    queryKey: ["system-settings"],
    queryFn: async () => {
      const res = await apiClient.get("/system-settings");
      return res.data;
    },
  });

  const settings: SystemSetting[] = data?.items ?? [];

  useEffect(() => {
    if (settings.length > 0) {
      const values: Record<string, string> = {};
      settings.forEach((s) => { values[s.key] = s.value; });
      setEditValues(values);
      setDirty({});
    }
  }, [data]);

  const updateMutation = useMutation({
    mutationFn: async ({ key, value, description }: { key: string; value: string; description: string }) => {
      await apiClient.put(`/system-settings/${key}`, { value, description });
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["system-settings"] });
      setDirty((prev) => ({ ...prev, [variables.key]: false }));
      showToast("配置已保存", "success");
    },
  });

  const handleChange = (key: string, value: string) => {
    setEditValues((prev) => ({ ...prev, [key]: value }));
    const original = settings.find((s) => s.key === key);
    setDirty((prev) => ({ ...prev, [key]: original?.value !== value }));
  };

  const handleSave = (setting: SystemSetting) => {
    updateMutation.mutate({
      key: setting.key,
      value: editValues[setting.key] ?? setting.value,
      description: setting.description,
    });
  };

  if (isLoading) {
    return <div className="py-12 text-center text-gray-500">加载中...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-gray-200 bg-white">
        <div className="border-b border-gray-200 px-6 py-4">
          <h3 className="text-base font-semibold text-gray-900">运行时配置</h3>
          <p className="mt-1 text-sm text-gray-500">这些配置存储在数据库中，修改后立即生效，无需重启服务</p>
        </div>
        <div className="divide-y divide-gray-100">
          {settings.map((setting) => {
            const meta = SETTING_LABELS[setting.key];
            return (
              <div key={setting.key} className="px-6 py-5">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-2">
                      <label className="text-sm font-medium text-gray-900">
                        {meta?.label || setting.key}
                      </label>
                      <code className="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-500">
                        {setting.key}
                      </code>
                    </div>
                    {(meta?.help || setting.description) && (
                      <p className="text-xs text-gray-500">{meta?.help || setting.description}</p>
                    )}
                    <input
                      type="text"
                      value={editValues[setting.key] ?? setting.value}
                      onChange={(e) => handleChange(setting.key, e.target.value)}
                      placeholder={meta?.placeholder || ""}
                      className="w-full rounded-lg border border-gray-300 px-3 py-2 font-mono text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                    {setting.updated_at && (
                      <p className="text-xs text-gray-400">
                        最后更新: {new Date(setting.updated_at).toLocaleString("zh-CN")}
                      </p>
                    )}
                  </div>
                  <button
                    onClick={() => handleSave(setting)}
                    disabled={!dirty[setting.key] || updateMutation.isPending}
                    className="mt-6 shrink-0 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-40"
                  >
                    {updateMutation.isPending ? "保存中..." : "保存"}
                  </button>
                </div>
              </div>
            );
          })}
          {settings.length === 0 && (
            <div className="px-6 py-12 text-center text-sm text-gray-500">
              暂无配置项。运行数据库迁移后会自动创建默认配置。
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
