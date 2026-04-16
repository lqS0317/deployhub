"use client";

import { useState, useEffect } from "react";
import { usePlugins } from "@/hooks/use-route-plugins";

interface ARBackend {
  serviceName: string;
  servicePort: string;
  weight: string;
  resolveGranularity: string;
}

interface ARPlugin {
  name: string;
  enable: boolean;
}

interface ARTimeout {
  connect: string;
  read: string;
  send: string;
}

interface ARRule {
  name: string;
  priority: string;
  hosts: string;
  paths: string;
  methods: string;
  backends: ARBackend[];
  plugins: ARPlugin[];
  timeout: ARTimeout;
  pluginConfigName: string;
}

interface ApisixRouteFormProps {
  value: Record<string, unknown>;
  onChange: (config: Record<string, unknown>) => void;
}

const inputCls = "w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500";
const smInputCls = "w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-blue-500 focus:outline-none";

const emptyBackend = (): ARBackend => ({ serviceName: "", servicePort: "80", weight: "", resolveGranularity: "" });
const emptyRule = (): ARRule => ({
  name: "rule-1", priority: "", hosts: "", paths: "/*", methods: "",
  backends: [emptyBackend()], plugins: [], timeout: { connect: "", read: "", send: "" }, pluginConfigName: "",
});

export function ApisixRouteForm({ value, onChange }: ApisixRouteFormProps) {
  const { data: availablePlugins = [] } = usePlugins();
  const [annotations, setAnnotations] = useState<{ key: string; value: string }[]>(
    Object.entries((value.annotations as Record<string, string>) || {}).map(([k, v]) => ({ key: k, value: v as string }))
  );
  const [rules, setRules] = useState<ARRule[]>(() => {
    const http = (value.http as any[]) || [];
    if (http.length > 0) {
      return http.map((r: any) => ({
        name: r.name || "",
        priority: r.priority ? String(r.priority) : "",
        hosts: (r.match?.hosts || []).join(", "),
        paths: (r.match?.paths || []).join(", "),
        methods: (r.match?.methods || []).join(", "),
        backends: (r.backends || []).map((b: any) => ({
          serviceName: b.serviceName || "", servicePort: String(b.servicePort || "80"),
          weight: b.weight ? String(b.weight) : "", resolveGranularity: b.resolveGranularity || "",
        })),
        plugins: (r.plugins || []).map((p: any) => ({ name: p.name || "", enable: p.enable !== false })),
        timeout: { connect: r.timeout?.connect || "", read: r.timeout?.read || "", send: r.timeout?.send || "" },
        pluginConfigName: r.plugin_config_name || "",
      }));
    }
    return [emptyRule()];
  });

  useEffect(() => {
    const annots: Record<string, string> = {};
    annotations.forEach(a => { if (a.key) annots[a.key] = a.value; });

    onChange({
      annotations: Object.keys(annots).length > 0 ? annots : undefined,
      http: rules.map(r => {
        const match: Record<string, unknown> = {
          paths: r.paths.split(",").map(s => s.trim()).filter(Boolean),
        };
        const hosts = r.hosts.split(",").map(s => s.trim()).filter(Boolean);
        if (hosts.length > 0) match.hosts = hosts;
        const methods = r.methods.split(",").map(s => s.trim()).filter(Boolean);
        if (methods.length > 0) match.methods = methods;

        const rule: Record<string, unknown> = {
          name: r.name,
          match,
          backends: r.backends.filter(b => b.serviceName).map(b => {
            const bk: Record<string, unknown> = { serviceName: b.serviceName, servicePort: Number(b.servicePort) || 80 };
            if (b.weight) bk.weight = Number(b.weight);
            if (b.resolveGranularity) bk.resolveGranularity = b.resolveGranularity;
            return bk;
          }),
        };
        if (r.priority) rule.priority = Number(r.priority);
        if (r.plugins.length > 0) {
          rule.plugins = r.plugins.filter(p => p.name).map(p => ({ name: p.name, enable: p.enable }));
        }
        const tm: Record<string, string> = {};
        if (r.timeout.connect) tm.connect = r.timeout.connect;
        if (r.timeout.read) tm.read = r.timeout.read;
        if (r.timeout.send) tm.send = r.timeout.send;
        if (Object.keys(tm).length > 0) rule.timeout = tm;
        if (r.pluginConfigName) rule.plugin_config_name = r.pluginConfigName;
        return rule;
      }),
    });
  }, [rules, annotations]);

  return (
    <div className="space-y-4">
      {/* Annotations */}
      <div>
        <div className="mb-1 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-700">Annotations</label>
          <button type="button" onClick={() => setAnnotations([...annotations, { key: "", value: "" }])}
            className="text-xs text-blue-600 hover:text-blue-800">+ 添加</button>
        </div>
        {annotations.map((a, i) => (
          <div key={i} className="mb-1 flex items-center gap-1.5">
            <input type="text" value={a.key} placeholder="key" onChange={e => { const arr = [...annotations]; arr[i] = { ...arr[i], key: e.target.value }; setAnnotations(arr); }} className={smInputCls} />
            <input type="text" value={a.value} placeholder="value" onChange={e => { const arr = [...annotations]; arr[i] = { ...arr[i], value: e.target.value }; setAnnotations(arr); }} className={smInputCls} />
            <button type="button" onClick={() => setAnnotations(annotations.filter((_, idx) => idx !== i))} className="text-red-400 hover:text-red-600 text-xs">✕</button>
          </div>
        ))}
      </div>

      {/* HTTP Rules */}
      <div>
        <div className="mb-1 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-700">HTTP Rules</label>
          <button type="button" onClick={() => setRules([...rules, { ...emptyRule(), name: `rule-${rules.length + 1}` }])}
            className="text-xs text-blue-600 hover:text-blue-800">+ 添加 Rule</button>
        </div>

        {rules.map((rule, ri) => (
          <div key={ri} className="mb-3 rounded-md border border-gray-200 bg-gray-50 p-3 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-xs font-medium text-gray-600">Rule #{ri + 1}</span>
              {rules.length > 1 && (
                <button type="button" onClick={() => setRules(rules.filter((_, idx) => idx !== ri))}
                  className="text-red-500 hover:text-red-700 text-lg leading-none">×</button>
              )}
            </div>

            {/* Name + Priority */}
            <div className="grid grid-cols-3 gap-2">
              <div>
                <label className="block text-xs text-gray-500">Name (必填)</label>
                <input type="text" value={rule.name} onChange={e => setRules(rules.map((r, i) => i === ri ? { ...r, name: e.target.value } : r))}
                  placeholder="rule-1" className={smInputCls} />
              </div>
              <div>
                <label className="block text-xs text-gray-500">Priority</label>
                <input type="number" value={rule.priority} onChange={e => setRules(rules.map((r, i) => i === ri ? { ...r, priority: e.target.value } : r))}
                  placeholder="自动" className={smInputCls} />
              </div>
              <div>
                <label className="block text-xs text-gray-500">PluginConfig</label>
                <input type="text" value={rule.pluginConfigName} onChange={e => setRules(rules.map((r, i) => i === ri ? { ...r, pluginConfigName: e.target.value } : r))}
                  placeholder="插件配置名称" className={smInputCls} />
              </div>
            </div>

            {/* Match */}
            <div className="grid grid-cols-3 gap-2">
              <div>
                <label className="block text-xs text-gray-500">Hosts（逗号分隔）</label>
                <input type="text" value={rule.hosts} onChange={e => setRules(rules.map((r, i) => i === ri ? { ...r, hosts: e.target.value } : r))}
                  placeholder="example.com, *.example.com" className={smInputCls} />
              </div>
              <div>
                <label className="block text-xs text-gray-500">Paths（逗号分隔）</label>
                <input type="text" value={rule.paths} onChange={e => setRules(rules.map((r, i) => i === ri ? { ...r, paths: e.target.value } : r))}
                  placeholder="/api/*, /health" className={smInputCls} />
              </div>
              <div>
                <label className="block text-xs text-gray-500">Methods（逗号分隔）</label>
                <input type="text" value={rule.methods} onChange={e => setRules(rules.map((r, i) => i === ri ? { ...r, methods: e.target.value } : r))}
                  placeholder="GET, POST" className={smInputCls} />
              </div>
            </div>

            {/* Backends */}
            <div>
              <div className="mb-1 flex items-center justify-between">
                <label className="text-xs font-medium text-gray-500">Backends</label>
                <button type="button" onClick={() => {
                  const nr = [...rules]; nr[ri] = { ...nr[ri], backends: [...nr[ri].backends, emptyBackend()] }; setRules(nr);
                }} className="text-xs text-blue-600">+ 添加</button>
              </div>
              {rule.backends.map((b, bi) => (
                <div key={bi} className="mb-1 flex items-center gap-1.5">
                  <div className="flex-1">
                    <input type="text" value={b.serviceName} placeholder="serviceName"
                      onChange={e => { const bs = [...rule.backends]; bs[bi] = { ...bs[bi], serviceName: e.target.value }; setRules(rules.map((r, i) => i === ri ? { ...r, backends: bs } : r)); }}
                      className={smInputCls} />
                  </div>
                  <div className="w-16">
                    <input type="number" value={b.servicePort} placeholder="80"
                      onChange={e => { const bs = [...rule.backends]; bs[bi] = { ...bs[bi], servicePort: e.target.value }; setRules(rules.map((r, i) => i === ri ? { ...r, backends: bs } : r)); }}
                      className={smInputCls} />
                  </div>
                  <div className="w-14">
                    <input type="number" value={b.weight} placeholder="wt"
                      onChange={e => { const bs = [...rule.backends]; bs[bi] = { ...bs[bi], weight: e.target.value }; setRules(rules.map((r, i) => i === ri ? { ...r, backends: bs } : r)); }}
                      className={smInputCls} />
                      </div>
                  <div className="w-24">
                    <select value={b.resolveGranularity}
                      onChange={e => { const bs = [...rule.backends]; bs[bi] = { ...bs[bi], resolveGranularity: e.target.value }; setRules(rules.map((r, i) => i === ri ? { ...r, backends: bs } : r)); }}
                      className={smInputCls}>
                      <option value="">默认</option>
                      <option value="endpoints">endpoints</option>
                      <option value="service">service</option>
                    </select>
                    </div>
                  {rule.backends.length > 1 && (
                    <button type="button" onClick={() => {
                      const bs = rule.backends.filter((_, idx) => idx !== bi);
                      setRules(rules.map((r, i) => i === ri ? { ...r, backends: bs } : r));
                    }} className="text-red-400 hover:text-red-600 text-xs">✕</button>
                  )}
                </div>
              ))}
            </div>

            {/* Timeout */}
            <div>
              <label className="mb-1 block text-xs font-medium text-gray-500">Timeout</label>
              <div className="flex gap-2">
                <div className="flex-1">
                  <label className="block text-xs text-gray-400">connect</label>
                  <input type="text" value={rule.timeout.connect} placeholder="60s"
                    onChange={e => setRules(rules.map((r, i) => i === ri ? { ...r, timeout: { ...r.timeout, connect: e.target.value } } : r))}
                    className={smInputCls} />
                </div>
                <div className="flex-1">
                  <label className="block text-xs text-gray-400">read</label>
                  <input type="text" value={rule.timeout.read} placeholder="60s"
                    onChange={e => setRules(rules.map((r, i) => i === ri ? { ...r, timeout: { ...r.timeout, read: e.target.value } } : r))}
                    className={smInputCls} />
                </div>
                <div className="flex-1">
                  <label className="block text-xs text-gray-400">send</label>
                  <input type="text" value={rule.timeout.send} placeholder="60s"
                    onChange={e => setRules(rules.map((r, i) => i === ri ? { ...r, timeout: { ...r.timeout, send: e.target.value } } : r))}
                    className={smInputCls} />
                </div>
              </div>
            </div>

            {/* Plugins */}
            <div>
              <div className="mb-1 flex items-center justify-between">
                <label className="text-xs font-medium text-gray-500">Plugins</label>
                <button type="button" onClick={() => {
                  const ps = [...rule.plugins, { name: "", enable: true }];
                  setRules(rules.map((r, i) => i === ri ? { ...r, plugins: ps } : r));
                }} className="text-xs text-blue-600 hover:text-blue-800">+ 手动添加</button>
              </div>
              {rule.plugins.map((p, pi) => (
                <div key={pi} className="mb-1 flex items-center gap-2">
                  <input type="text" value={p.name} placeholder="plugin 名称"
                    onChange={e => { const ps = [...rule.plugins]; ps[pi] = { ...ps[pi], name: e.target.value }; setRules(rules.map((r, i) => i === ri ? { ...r, plugins: ps } : r)); }}
                    className={smInputCls} />
                  <label className="flex items-center gap-1 text-xs whitespace-nowrap">
                    <input type="checkbox" checked={p.enable}
                      onChange={e => { const ps = [...rule.plugins]; ps[pi] = { ...ps[pi], enable: e.target.checked }; setRules(rules.map((r, i) => i === ri ? { ...r, plugins: ps } : r)); }} />
                    启用
                  </label>
                  <button type="button" onClick={() => {
                    const ps = rule.plugins.filter((_, idx) => idx !== pi);
                    setRules(rules.map((r, i) => i === ri ? { ...r, plugins: ps } : r));
                  }} className="text-red-400 hover:text-red-600 text-xs">✕</button>
                </div>
              ))}
              {rule.plugins.length === 0 && availablePlugins.length === 0 && (
                <p className="text-xs text-gray-400">暂无插件，可手动添加或在插件中心创建</p>
              )}
              {availablePlugins.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-1">
                  <span className="text-xs text-gray-400">从插件中心选择：</span>
                  {availablePlugins.filter(ap => !rule.plugins.some(p => p.name === ap.name)).map(ap => (
                    <button key={ap.id} type="button" onClick={() => {
                      const ps = [...rule.plugins, { name: ap.name, enable: true }];
                      setRules(rules.map((r, i) => i === ri ? { ...r, plugins: ps } : r));
                    }} className="rounded bg-blue-50 px-1.5 py-0.5 text-xs text-blue-600 hover:bg-blue-100">
                      + {ap.name}
              </button>
                  ))}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
