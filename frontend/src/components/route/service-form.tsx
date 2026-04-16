"use client";

import { useState, useEffect } from "react";

interface ServicePort {
  name: string;
  port: string;
  targetPort: string;
  protocol: string;
}

interface SelectorRow {
  key: string;
  value: string;
}

interface ServiceConfig {
  type: string;
  selector: Record<string, string>;
  ports: Array<{
    name: string;
    port: number;
    targetPort: number;
    protocol: string;
  }>;
}

interface ServiceFormProps {
  value: ServiceConfig;
  onChange: (config: ServiceConfig) => void;
}

export function ServiceForm({ value, onChange }: ServiceFormProps) {
  const [type, setType] = useState(value.type || "ClusterIP");
  const [selectors, setSelectors] = useState<SelectorRow[]>(() => {
    const s = value.selector || {};
    const rows = Object.entries(s).map(([k, v]) => ({ key: k, value: v }));
    return rows.length > 0 ? rows : [{ key: "", value: "" }];
  });
  const [ports, setPorts] = useState<ServicePort[]>(() => {
    const p = value.ports || [];
    const rows = p.map((r) => ({
      name: r.name || "",
      port: String(r.port || ""),
      targetPort: String(r.targetPort || ""),
      protocol: r.protocol || "TCP",
    }));
    return rows.length > 0
      ? rows
      : [{ name: "http", port: "80", targetPort: "8080", protocol: "TCP" }];
  });

  useEffect(() => {
    const selector: Record<string, string> = {};
    selectors.forEach((s) => {
      if (s.key.trim()) selector[s.key.trim()] = s.value.trim();
    });
    const parsedPorts = ports
      .filter((p) => p.port)
      .map((p) => ({
        name: p.name,
        port: Number(p.port),
        targetPort: Number(p.targetPort) || Number(p.port),
        protocol: p.protocol,
      }));
    onChange({ type, selector, ports: parsedPorts });
  }, [type, selectors, ports]);

  const addSelector = () =>
    setSelectors([...selectors, { key: "", value: "" }]);
  const removeSelector = (i: number) =>
    setSelectors(selectors.filter((_, idx) => idx !== i));
  const updateSelector = (i: number, field: "key" | "value", val: string) =>
    setSelectors(selectors.map((s, idx) => (idx === i ? { ...s, [field]: val } : s)));

  const addPort = () =>
    setPorts([...ports, { name: "", port: "", targetPort: "", protocol: "TCP" }]);
  const removePort = (i: number) =>
    setPorts(ports.filter((_, idx) => idx !== i));
  const updatePort = (i: number, field: keyof ServicePort, val: string) =>
    setPorts(ports.map((p, idx) => (idx === i ? { ...p, [field]: val } : p)));

  const inputCls =
    "w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500";

  return (
    <div className="space-y-4">
      <div>
        <label className="mb-1 block text-sm font-medium text-gray-700">
          Service 类型
        </label>
        <select
          value={type}
          onChange={(e) => setType(e.target.value)}
          className={inputCls}
        >
          <option value="ClusterIP">ClusterIP</option>
          <option value="NodePort">NodePort</option>
          <option value="LoadBalancer">LoadBalancer</option>
        </select>
      </div>

      <div>
        <div className="mb-1 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-700">Selector</label>
          <button
            type="button"
            onClick={addSelector}
            className="text-xs text-blue-600 hover:text-blue-800"
          >
            + 添加
          </button>
        </div>
        <div className="space-y-2">
          {selectors.map((s, i) => (
            <div key={i} className="flex items-center gap-2">
              <input
                type="text"
                value={s.key}
                onChange={(e) => updateSelector(i, "key", e.target.value)}
                placeholder="key"
                className={inputCls}
              />
              <span className="text-gray-400">=</span>
              <input
                type="text"
                value={s.value}
                onChange={(e) => updateSelector(i, "value", e.target.value)}
                placeholder="value"
                className={inputCls}
              />
              {selectors.length > 1 && (
                <button
                  type="button"
                  onClick={() => removeSelector(i)}
                  className="text-red-500 hover:text-red-700 text-lg leading-none"
                >
                  ×
                </button>
              )}
            </div>
          ))}
        </div>
      </div>

      <div>
        <div className="mb-1 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-700">Ports</label>
          <button
            type="button"
            onClick={addPort}
            className="text-xs text-blue-600 hover:text-blue-800"
          >
            + 添加
          </button>
        </div>
        <div className="space-y-2">
          {ports.map((p, i) => (
            <div key={i} className="flex items-center gap-2">
              <input
                type="text"
                value={p.name}
                onChange={(e) => updatePort(i, "name", e.target.value)}
                placeholder="name"
                className={inputCls}
              />
              <input
                type="number"
                value={p.port}
                onChange={(e) => updatePort(i, "port", e.target.value)}
                placeholder="port"
                className={inputCls}
              />
              <input
                type="number"
                value={p.targetPort}
                onChange={(e) => updatePort(i, "targetPort", e.target.value)}
                placeholder="targetPort"
                className={inputCls}
              />
              <select
                value={p.protocol}
                onChange={(e) => updatePort(i, "protocol", e.target.value)}
                className={inputCls}
              >
                <option value="TCP">TCP</option>
                <option value="UDP">UDP</option>
              </select>
              {ports.length > 1 && (
                <button
                  type="button"
                  onClick={() => removePort(i)}
                  className="text-red-500 hover:text-red-700 text-lg leading-none"
                >
                  ×
                </button>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
