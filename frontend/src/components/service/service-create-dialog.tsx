"use client";

import { useState, useEffect } from "react";
import { useCreateService, useUpdateService } from "@/hooks/use-services";
import apiClient from "@/lib/api-client";
import { useQuery } from "@tanstack/react-query";
import type { Service } from "@/types";
import { ServiceRuntimeConfigForm, getEmptyRuntimeConfig, fromServiceDefaults, toServiceDefaultFields, type ServiceRuntimeConfig } from "@/components/service/service-runtime-config";

interface ServiceFormDialogProps {
  service?: Service;
  onClose: () => void;
}

interface FormData {
  name: string;
  display_name: string;
  description: string;
  deploy_type: string;
  service_type: string;
  language: string;
  language_version: string;
  git_repo_id: string;
  git_branch: string;
}

const INITIAL_FORM: FormData = {
  name: "",
  display_name: "",
  description: "",
  deploy_type: "direct",
  service_type: "",
  language: "",
  language_version: "",
  git_repo_id: "",
  git_branch: "main",
};

const DEPLOY_TYPES = [
  { value: "direct", label: "直接部署" },
  { value: "helm", label: "Helm 部署" },
  { value: "other", label: "其他（预留）" },
];

const SERVICE_TYPES = [
  { value: "backend", label: "后端服务" },
  { value: "frontend", label: "前端应用" },
  { value: "worker", label: "Worker" },
  { value: "cronjob", label: "定时任务" },
  { value: "middleware", label: "中间件" },
  { value: "other", label: "其他" },
];

const LANGUAGES = [
  { value: "go", label: "Go" },
  { value: "node", label: "Node.js" },
  { value: "java", label: "Java" },
  { value: "python", label: "Python" },
  { value: "rust", label: "Rust" },
  { value: "elixir", label: "Elixir" },
  { value: "other", label: "其他" },
];

function serviceToForm(svc: Service): FormData {
  return {
    name: svc.name,
    display_name: svc.display_name || "",
    description: svc.description || "",
    deploy_type: svc.deploy_type || "direct",
    service_type: svc.service_type || "",
    language: svc.language || "",
    language_version: svc.language_version || "",
    git_repo_id: String(svc.git_repo_id),
    git_branch: svc.git_branch || "main",
  };
}

export function ServiceCreateDialog({ onClose }: { onClose: () => void }) {
  return <ServiceFormDialog onClose={onClose} />;
}

export function ServiceFormDialog({ service, onClose }: ServiceFormDialogProps) {
  const isEdit = !!service;
  const [form, setForm] = useState<FormData>(service ? serviceToForm(service) : INITIAL_FORM);
  const [runtimeConfig, setRuntimeConfig] = useState<ServiceRuntimeConfig>(
    service ? fromServiceDefaults(service as unknown as Record<string, unknown>) : getEmptyRuntimeConfig()
  );
  const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});
  const createService = useCreateService();
  const updateService = useUpdateService(service?.id ?? 0);

  useEffect(() => {
    if (service) setForm(serviceToForm(service));
  }, [service]);

  const { data: gitRepos } = useQuery({
    queryKey: ["git-repos"],
    queryFn: () => apiClient.get("/git-repos").then((r) => r.data),
  });

  const updateField = <K extends keyof FormData>(key: K, value: FormData[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
    if (errors[key]) setErrors((prev) => ({ ...prev, [key]: undefined }));
  };

  const validate = (): boolean => {
    const newErrors: Partial<Record<keyof FormData, string>> = {};
    if (!form.name.trim()) newErrors.name = "服务名称不能为空";
    if (!form.git_repo_id) newErrors.git_repo_id = "请选择 Git 仓库";
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    const runtimeFields = form.deploy_type === "direct" ? toServiceDefaultFields(runtimeConfig) : {};
    const payload = {
      ...form,
      git_repo_id: Number(form.git_repo_id),
      ...runtimeFields,
    };

    const mutation = isEdit ? updateService : createService;
    mutation.mutate(payload as Partial<Service>, {
      onSuccess: () => onClose(),
    });
  };

  const isPending = isEdit ? updateService.isPending : createService.isPending;

  const inputClass = (field: keyof FormData) =>
    `w-full rounded-lg border px-3 py-2 text-sm focus:outline-none focus:ring-1 ${
      errors[field] ? "border-red-300 focus:border-red-500 focus:ring-red-500"
        : "border-gray-300 focus:border-blue-500 focus:ring-blue-500"
    }`;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />

      <div className="relative z-10 w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-xl bg-white p-6 shadow-2xl">
        <div className="flex items-center justify-between border-b border-gray-200 pb-4">
          <h2 className="text-lg font-semibold text-gray-900">
            {isEdit ? "编辑服务" : "创建服务"}
          </h2>
          <button onClick={onClose} className="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600">
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          {/* 服务名称 + 显示名称 */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                服务名称 <span className="text-red-500">*</span>
              </label>
              <input
                type="text" value={form.name}
                onChange={(e) => updateField("name", e.target.value)}
                placeholder="例如: payment-api"
                disabled={isEdit}
                className={`${inputClass("name")} disabled:bg-gray-50 disabled:text-gray-500`}
              />
              {errors.name && <p className="mt-1 text-xs text-red-500">{errors.name}</p>}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">显示名称</label>
              <input
                type="text" value={form.display_name}
                onChange={(e) => updateField("display_name", e.target.value)}
                placeholder="例如: 支付服务"
                className={inputClass("display_name")}
              />
            </div>
          </div>

          {/* 描述 */}
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">描述</label>
            <textarea
              value={form.description}
              onChange={(e) => updateField("description", e.target.value)}
              placeholder="服务描述信息..."
              rows={2}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          {/* 业务类型 + 开发语言 */}
          <div className="grid grid-cols-3 gap-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">部署类型</label>
              <select value={form.deploy_type} onChange={(e) => updateField("deploy_type", e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500">
                {DEPLOY_TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">业务类型</label>
              <select value={form.service_type} onChange={(e) => updateField("service_type", e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500">
                <option value="">请选择</option>
                {SERVICE_TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">开发语言</label>
              <select value={form.language} onChange={(e) => updateField("language", e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500">
                <option value="">请选择</option>
                {LANGUAGES.map((l) => <option key={l.value} value={l.value}>{l.label}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">运行时版本</label>
              <input
                type="text" value={form.language_version}
                onChange={(e) => updateField("language_version", e.target.value)}
                placeholder="如 go1.22 / node20"
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          </div>

          {/* Git 仓库 + 默认分支 */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">
                Git 仓库 <span className="text-red-500">*</span>
              </label>
              <select value={form.git_repo_id} onChange={(e) => updateField("git_repo_id", e.target.value)}
                className={inputClass("git_repo_id")}>
                <option value="">请选择</option>
                {(gitRepos?.items ?? []).map((r: { id: string; name: string }) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
              {errors.git_repo_id && <p className="mt-1 text-xs text-red-500">{errors.git_repo_id}</p>}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">默认分支</label>
              <input type="text" value={form.git_branch}
                onChange={(e) => updateField("git_branch", e.target.value)}
                placeholder="main"
                className={inputClass("git_branch")}
              />
            </div>
          </div>

          {/* 服务配置分隔线 + 运行时配置 */}
          {form.deploy_type === "direct" && (
            <>
              <div className="flex items-center gap-3 pt-2">
                <div className="h-px flex-1 bg-gray-200" />
                <span className="text-xs font-medium text-gray-400 uppercase tracking-wider">服务配置</span>
                <div className="h-px flex-1 bg-gray-200" />
              </div>
              <ServiceRuntimeConfigForm value={runtimeConfig} onChange={setRuntimeConfig} />
            </>
          )}

          <div className="flex justify-end gap-3 border-t border-gray-200 pt-4">
            <button type="button" onClick={onClose}
              className="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">
              取消
            </button>
            <button type="submit" disabled={isPending}
              className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50">
              {isPending ? "保存中..." : isEdit ? "保存" : "创建服务"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
