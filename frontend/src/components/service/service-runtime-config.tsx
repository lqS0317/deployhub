"use client";

import { DirectConfigForm, getEmptyConfig, type DirectConfigData } from "@/components/deploy/direct-config-form";

export interface ServiceRuntimeConfig {
  workloadType: string;
  port: number;
  replicas: number;
  cpuRequest: string;
  memRequest: string;
  cpuLimit: string;
  memLimit: string;
  config: DirectConfigData;
}

export function getEmptyRuntimeConfig(): ServiceRuntimeConfig {
  return {
    workloadType: "deployment",
    port: 0,
    replicas: 1,
    cpuRequest: "",
    memRequest: "",
    cpuLimit: "",
    memLimit: "",
    config: getEmptyConfig(),
  };
}

interface Props {
  value: ServiceRuntimeConfig;
  onChange: (val: ServiceRuntimeConfig) => void;
}

export function ServiceRuntimeConfigForm({ value, onChange }: Props) {
  const update = <K extends keyof ServiceRuntimeConfig>(key: K, val: ServiceRuntimeConfig[K]) => {
    onChange({ ...value, [key]: val });
  };

  return (
    <div className="space-y-3">
      {/* 基础配置 */}
      <div className="grid grid-cols-3 gap-2">
        <div>
          <label className="mb-1 block text-xs text-gray-600">工作负载类型</label>
          <select value={value.workloadType} onChange={(e) => update("workloadType", e.target.value)}
            className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm">
            <option value="deployment">Deployment</option>
            <option value="statefulset">StatefulSet</option>
          </select>
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-600">Pod 端口</label>
          <input type="number" value={value.port || ""} placeholder="8080"
            onChange={(e) => update("port", Number(e.target.value) || 0)}
            className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm" />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-600">默认副本数</label>
          <input type="number" min={1} value={value.replicas}
            onChange={(e) => update("replicas", Number(e.target.value) || 1)}
            className="w-full rounded border border-gray-300 px-2 py-1.5 text-sm" />
        </div>
      </div>

      {/* 资源限制 */}
      <div className="grid grid-cols-4 gap-2">
        <div>
          <label className="mb-1 block text-xs text-gray-600">CPU Request</label>
          <input type="text" value={value.cpuRequest} placeholder="100m"
            onChange={(e) => update("cpuRequest", e.target.value)}
            className="w-full rounded border border-gray-300 px-2 py-1.5 text-xs font-mono" />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-600">Mem Request</label>
          <input type="text" value={value.memRequest} placeholder="128Mi"
            onChange={(e) => update("memRequest", e.target.value)}
            className="w-full rounded border border-gray-300 px-2 py-1.5 text-xs font-mono" />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-600">CPU Limit</label>
          <input type="text" value={value.cpuLimit} placeholder="500m"
            onChange={(e) => update("cpuLimit", e.target.value)}
            className="w-full rounded border border-gray-300 px-2 py-1.5 text-xs font-mono" />
        </div>
        <div>
          <label className="mb-1 block text-xs text-gray-600">Mem Limit</label>
          <input type="text" value={value.memLimit} placeholder="512Mi"
            onChange={(e) => update("memLimit", e.target.value)}
            className="w-full rounded border border-gray-300 px-2 py-1.5 text-xs font-mono" />
        </div>
      </div>

      {/* 高级配置（启动命令 + 健康检查） */}
      <DirectConfigForm value={value.config} onChange={(c) => update("config", c)} workloadType={value.workloadType} />
    </div>
  );
}

// 从 Service API 响应还原 RuntimeConfig
export function fromServiceDefaults(svc: Record<string, unknown>): ServiceRuntimeConfig {
  const cfg = getEmptyConfig();
  if (svc.default_liveness_probe && typeof svc.default_liveness_probe === "object") cfg.livenessProbe = svc.default_liveness_probe as typeof cfg.livenessProbe;
  if (svc.default_readiness_probe && typeof svc.default_readiness_probe === "object") cfg.readinessProbe = svc.default_readiness_probe as typeof cfg.readinessProbe;
  if (svc.default_command && Array.isArray(svc.default_command)) cfg.command = svc.default_command as string[];
  if (svc.default_args && Array.isArray(svc.default_args)) cfg.args = svc.default_args as string[];

  return {
    workloadType: (svc.default_workload_type as string) || "deployment",
    port: (svc.default_port as number) || 0,
    replicas: (svc.default_replicas as number) || 1,
    cpuRequest: (svc.default_cpu_request as string) || "",
    memRequest: (svc.default_mem_request as string) || "",
    cpuLimit: (svc.default_cpu_limit as string) || "",
    memLimit: (svc.default_mem_limit as string) || "",
    config: cfg,
  };
}

// 将 RuntimeConfig 序列化为 API 请求字段
export function toServiceDefaultFields(rc: ServiceRuntimeConfig): Record<string, unknown> {
  return {
    default_workload_type: rc.workloadType,
    default_port: rc.port,
    default_replicas: rc.replicas,
    default_cpu_request: rc.cpuRequest,
    default_mem_request: rc.memRequest,
    default_cpu_limit: rc.cpuLimit,
    default_mem_limit: rc.memLimit,
    default_liveness_probe: rc.config.livenessProbe.type ? rc.config.livenessProbe : undefined,
    default_readiness_probe: rc.config.readinessProbe.type ? rc.config.readinessProbe : undefined,
    default_command: rc.config.command.length > 0 ? rc.config.command : undefined,
    default_args: rc.config.args.length > 0 ? rc.config.args : undefined,
  };
}
