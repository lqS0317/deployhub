"use client";

import { useState } from "react";

// --- 类型定义（与后端 ConfigExecutor 的 JSONB 结构一致） ---

export interface ProbeConfig {
  type: string; // http | tcp | exec | ""
  path?: string;
  port?: number;
  command?: string;
  initialDelaySeconds?: number;
  periodSeconds?: number;
}

export interface DirectConfigData {
  command: string[];
  args: string[];
  livenessProbe: ProbeConfig;
  readinessProbe: ProbeConfig;
}

const EMPTY_CONFIG: DirectConfigData = {
  command: [],
  args: [],
  livenessProbe: { type: "" },
  readinessProbe: { type: "" },
};

interface Props {
  value: DirectConfigData;
  onChange: (val: DirectConfigData) => void;
  workloadType: string;
}

export function getEmptyConfig(): DirectConfigData {
  return JSON.parse(JSON.stringify(EMPTY_CONFIG));
}

export function DirectConfigForm({ value, onChange }: Props) {
  const [expandedSection, setExpandedSection] = useState<string | null>("cmd");
  const [cmdText, setCmdText] = useState(value.command.join(" "));
  const [argsText, setArgsText] = useState(value.args.join(" "));

  const update = <K extends keyof DirectConfigData>(key: K, val: DirectConfigData[K]) => {
    onChange({ ...value, [key]: val });
  };

  const Section = ({ id, title, children, badge }: { id: string; title: string; children: React.ReactNode; badge?: string }) => (
    <div className="border border-gray-200 rounded-lg overflow-hidden">
      <button type="button" onClick={() => setExpandedSection(expandedSection === id ? null : id)}
        className="w-full flex items-center justify-between px-3 py-2 bg-gray-50 text-sm font-medium text-gray-700 hover:bg-gray-100">
        <span>{title} {badge && <span className="ml-1 text-xs text-blue-600">({badge})</span>}</span>
        <span className="text-gray-400">{expandedSection === id ? "▲" : "▼"}</span>
      </button>
      {expandedSection === id && <div className="p-3 space-y-2">{children}</div>}
    </div>
  );

  return (
    <div className="space-y-2">
      {/* 启动命令 */}
      <Section id="cmd" title="启动命令" badge={value.command.length > 0 || value.args.length > 0 ? "已配置" : undefined}>
        <div className="space-y-2">
          <div>
            <label className="block text-xs text-gray-600 font-medium mb-1">Command（覆盖镜像 ENTRYPOINT）</label>
            <input type="text" value={cmdText} placeholder="留空使用镜像默认，例如: /bin/sh -c"
              onChange={(e) => setCmdText(e.target.value)}
              onBlur={() => update("command", cmdText.trim() ? cmdText.trim().split(/\s+/) : [])}
              className="w-full rounded border border-gray-300 px-2 py-1 text-xs font-mono" />
            <p className="mt-0.5 text-xs text-gray-400">按空格分隔，如 /bin/sh -c</p>
          </div>
          <div>
            <label className="block text-xs text-gray-600 font-medium mb-1">Args（覆盖镜像 CMD）</label>
            <input type="text" value={argsText} placeholder="留空使用镜像默认，例如: --port=8080 --config=/etc/app.yaml"
              onChange={(e) => setArgsText(e.target.value)}
              onBlur={() => update("args", argsText.trim() ? argsText.trim().split(/\s+/) : [])}
              className="w-full rounded border border-gray-300 px-2 py-1 text-xs font-mono" />
            <p className="mt-0.5 text-xs text-gray-400">按空格分隔</p>
          </div>
        </div>
      </Section>

      {/* 健康检查 */}
      <Section id="probes" title="健康检查"
        badge={value.livenessProbe.type || value.readinessProbe.type ? "已配置" : undefined}>
        <ProbeEditor label="存活探针 (Liveness)" value={value.livenessProbe}
          onChange={(p) => update("livenessProbe", p)} />
        <div className="border-t border-gray-100 my-2" />
        <ProbeEditor label="就绪探针 (Readiness)" value={value.readinessProbe}
          onChange={(p) => update("readinessProbe", p)} />
      </Section>
    </div>
  );
}

// --- Probe 编辑子组件 ---
function ProbeEditor({ label, value, onChange }: { label: string; value: ProbeConfig; onChange: (v: ProbeConfig) => void }) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs text-gray-600 font-medium">{label}</label>
      <div className="flex items-center gap-2">
        <select value={value.type} onChange={(e) => onChange({ ...value, type: e.target.value })}
          className="rounded border border-gray-300 px-2 py-1 text-xs">
          <option value="">不配置</option>
          <option value="http">HTTP</option>
          <option value="tcp">TCP</option>
          <option value="exec">Exec</option>
        </select>
        {value.type === "http" && (
          <>
            <input type="text" value={value.path || ""} placeholder="/healthz"
              onChange={(e) => onChange({ ...value, path: e.target.value })}
              className="flex-1 rounded border border-gray-300 px-2 py-1 text-xs" />
            <input type="number" value={value.port || ""} placeholder="端口"
              onChange={(e) => onChange({ ...value, port: Number(e.target.value) || undefined })}
              className="w-16 rounded border border-gray-300 px-2 py-1 text-xs" />
          </>
        )}
        {value.type === "tcp" && (
          <input type="number" value={value.port || ""} placeholder="端口"
            onChange={(e) => onChange({ ...value, port: Number(e.target.value) || undefined })}
            className="w-20 rounded border border-gray-300 px-2 py-1 text-xs" />
        )}
        {value.type === "exec" && (
          <input type="text" value={value.command || ""} placeholder="cat /tmp/healthy"
            onChange={(e) => onChange({ ...value, command: e.target.value })}
            className="flex-1 rounded border border-gray-300 px-2 py-1 text-xs font-mono" />
        )}
      </div>
      {value.type && (
        <div className="flex items-center gap-2">
          <label className="text-xs text-gray-500">延迟(s):</label>
          <input type="number" value={value.initialDelaySeconds || ""} placeholder="0"
            onChange={(e) => onChange({ ...value, initialDelaySeconds: Number(e.target.value) || undefined })}
            className="w-14 rounded border border-gray-300 px-1.5 py-0.5 text-xs" />
          <label className="text-xs text-gray-500">间隔(s):</label>
          <input type="number" value={value.periodSeconds || ""} placeholder="10"
            onChange={(e) => onChange({ ...value, periodSeconds: Number(e.target.value) || undefined })}
            className="w-14 rounded border border-gray-300 px-1.5 py-0.5 text-xs" />
        </div>
      )}
    </div>
  );
}
