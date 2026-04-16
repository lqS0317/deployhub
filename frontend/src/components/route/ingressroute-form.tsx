"use client";

import { useState, useEffect } from "react";
import { usePlugins } from "@/hooks/use-route-plugins";

interface IRService {
  name: string;
  port: string;
  namespace: string;
  kind: string;
  passHostHeader: boolean;
  scheme: string;
  nativeLB: boolean;
  weight: string;
}

interface IRMiddleware {
  name: string;
  namespace: string;
}

interface IRRoute {
  match: string;
  kind: string;
  priority: string;
  services: IRService[];
  middlewares: IRMiddleware[];
}

interface IRTLSDomain {
  main: string;
  sans: string;
}

interface IngressRouteFormProps {
  value: Record<string, unknown>;
  onChange: (config: Record<string, unknown>) => void;
}

const inputCls = "w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500";
const smInputCls = "w-full rounded border border-gray-300 px-2 py-1 text-xs focus:border-blue-500 focus:outline-none";

const emptyService = (): IRService => ({ name: "", port: "", namespace: "", kind: "Service", passHostHeader: true, scheme: "", nativeLB: false, weight: "" });
const emptyRoute = (): IRRoute => ({ match: "Host(`example.com`)", kind: "Rule", priority: "", services: [emptyService()], middlewares: [] });

export function IngressRouteForm({ value, onChange }: IngressRouteFormProps) {
  const { data: plugins = [] } = usePlugins();

  const [entryPoints, setEntryPoints] = useState((value.entryPoints as string[])?.join(", ") || "websecure");
  const [certResolver, setCertResolver] = useState((value.tls as any)?.certResolver || "");
  const [secretName, setSecretName] = useState((value.tls as any)?.secretName || "");
  const [tlsDomains, setTlsDomains] = useState<IRTLSDomain[]>((value.tls as any)?.domains?.map((d: any) => ({ main: d.main || "", sans: d.sans?.join(", ") || "" })) || []);
  const [tlsOptionsName, setTlsOptionsName] = useState((value.tls as any)?.options?.name || "");
  const [annotations, setAnnotations] = useState<{ key: string; value: string }[]>(
    Object.entries((value.annotations as Record<string, string>) || {}).map(([k, v]) => ({ key: k, value: v as string }))
  );
  const [routes, setRoutes] = useState<IRRoute[]>(() => {
    const r = (value.routes as any[]) || [];
    if (r.length > 0) {
      return r.map((rt: any) => ({
        match: rt.match || "",
        kind: rt.kind || "Rule",
        priority: rt.priority ? String(rt.priority) : "",
        services: (rt.services || []).map((s: any) => ({
          name: s.name || "", port: String(s.port || ""), namespace: s.namespace || "",
          kind: s.kind || "Service", passHostHeader: s.passHostHeader !== false,
          scheme: s.scheme || "", nativeLB: !!s.nativeLB, weight: s.weight ? String(s.weight) : "",
        })),
        middlewares: (rt.middlewares || []).map((m: any) => typeof m === "string" ? { name: m, namespace: "" } : { name: m.name || "", namespace: m.namespace || "" }),
      }));
    }
    return [emptyRoute()];
  });

  // 同步到父组件
  useEffect(() => {
    const eps = entryPoints.split(",").map(e => e.trim()).filter(Boolean);
    const annots: Record<string, string> = {};
    annotations.forEach(a => { if (a.key) annots[a.key] = a.value; });

    const tls: Record<string, unknown> = {};
    if (certResolver) tls.certResolver = certResolver;
    if (secretName) tls.secretName = secretName;
    if (tlsDomains.length > 0) {
      tls.domains = tlsDomains.filter(d => d.main).map(d => ({
        main: d.main, sans: d.sans ? d.sans.split(",").map(s => s.trim()).filter(Boolean) : undefined,
      }));
    }
    if (tlsOptionsName) tls.options = { name: tlsOptionsName };

    onChange({
      entryPoints: eps,
      annotations: Object.keys(annots).length > 0 ? annots : undefined,
      routes: routes.map(r => ({
        kind: r.kind || "Rule",
        match: r.match,
        priority: r.priority ? Number(r.priority) : undefined,
        services: r.services.filter(s => s.name).map(s => {
          const svc: Record<string, unknown> = { name: s.name, port: Number(s.port) || 80 };
          if (s.namespace) svc.namespace = s.namespace;
          if (s.kind && s.kind !== "Service") svc.kind = s.kind;
          if (!s.passHostHeader) svc.passHostHeader = false;
          if (s.scheme) svc.scheme = s.scheme;
          if (s.nativeLB) svc.nativeLB = true;
          if (s.weight) svc.weight = Number(s.weight);
          return svc;
        }),
        middlewares: r.middlewares.filter(m => m.name).map(m => {
          const mw: Record<string, string> = { name: m.name };
          if (m.namespace) mw.namespace = m.namespace;
          return mw;
        }),
      })),
      tls: Object.keys(tls).length > 0 ? tls : undefined,
    });
  }, [entryPoints, routes, certResolver, secretName, tlsDomains, tlsOptionsName, annotations]);

  return (
    <div className="space-y-4">
      {/* EntryPoints */}
      <div>
        <label className="mb-1 block text-sm font-medium text-gray-700">EntryPoints</label>
        <input type="text" value={entryPoints} onChange={e => setEntryPoints(e.target.value)}
          placeholder="web, websecure" className={inputCls} />
        <p className="mt-1 text-xs text-gray-400">逗号分隔，如 web, websecure</p>
      </div>

      {/* Annotations */}
      <div>
        <div className="mb-1 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-700">Annotations</label>
          <button type="button" onClick={() => setAnnotations([...annotations, { key: "", value: "" }])}
            className="text-xs text-blue-600 hover:text-blue-800">+ 添加</button>
        </div>
        {annotations.map((a, i) => (
          <div key={i} className="mb-1 flex items-center gap-1.5">
            <input type="text" value={a.key} placeholder="key"
              onChange={e => { const arr = [...annotations]; arr[i] = { ...arr[i], key: e.target.value }; setAnnotations(arr); }}
              className={smInputCls} />
            <input type="text" value={a.value} placeholder="value"
              onChange={e => { const arr = [...annotations]; arr[i] = { ...arr[i], value: e.target.value }; setAnnotations(arr); }}
              className={smInputCls} />
            <button type="button" onClick={() => setAnnotations(annotations.filter((_, idx) => idx !== i))}
              className="text-red-400 hover:text-red-600 text-xs">✕</button>
          </div>
        ))}
      </div>

      {/* TLS */}
      <div className="rounded-md border border-gray-200 p-3 space-y-2">
        <label className="text-sm font-medium text-gray-700">TLS</label>
        <div className="grid grid-cols-3 gap-2">
          <div>
            <label className="block text-xs text-gray-500">certResolver</label>
            <input type="text" value={certResolver} onChange={e => setCertResolver(e.target.value)}
              placeholder="letsencrypt" className={smInputCls} />
          </div>
          <div>
            <label className="block text-xs text-gray-500">secretName</label>
            <input type="text" value={secretName} onChange={e => setSecretName(e.target.value)}
              placeholder="tls-secret" className={smInputCls} />
          </div>
          <div>
            <label className="block text-xs text-gray-500">TLS Options</label>
            <input type="text" value={tlsOptionsName} onChange={e => setTlsOptionsName(e.target.value)}
              placeholder="TLSOption 名称" className={smInputCls} />
          </div>
        </div>
        {/* TLS Domains */}
        <div>
          <div className="flex items-center justify-between">
            <label className="text-xs text-gray-500">Domains</label>
            <button type="button" onClick={() => setTlsDomains([...tlsDomains, { main: "", sans: "" }])}
              className="text-xs text-blue-600">+ 添加域名</button>
          </div>
          {tlsDomains.map((d, i) => (
            <div key={i} className="mt-1 flex items-center gap-1.5">
              <input type="text" value={d.main} placeholder="main: example.com"
                onChange={e => { const arr = [...tlsDomains]; arr[i] = { ...arr[i], main: e.target.value }; setTlsDomains(arr); }}
                className={smInputCls} />
              <input type="text" value={d.sans} placeholder="SANs: *.example.com (逗号分隔)"
                onChange={e => { const arr = [...tlsDomains]; arr[i] = { ...arr[i], sans: e.target.value }; setTlsDomains(arr); }}
                className={smInputCls} />
              <button type="button" onClick={() => setTlsDomains(tlsDomains.filter((_, idx) => idx !== i))}
                className="text-red-400 hover:text-red-600 text-xs">✕</button>
            </div>
          ))}
        </div>
      </div>

      {/* Routes */}
      <div>
        <div className="mb-1 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-700">Routes</label>
          <button type="button" onClick={() => setRoutes([...routes, emptyRoute()])}
            className="text-xs text-blue-600 hover:text-blue-800">+ 添加 Route</button>
        </div>
        <div className="space-y-3">
          {routes.map((route, ri) => (
            <div key={ri} className="rounded-md border border-gray-200 bg-gray-50 p-3 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-gray-600">Route #{ri + 1}</span>
                {routes.length > 1 && (
                  <button type="button" onClick={() => setRoutes(routes.filter((_, idx) => idx !== ri))}
                    className="text-red-500 hover:text-red-700 text-lg leading-none">×</button>
                )}
              </div>

              {/* Match + Priority */}
              <div className="grid grid-cols-4 gap-2">
                <div className="col-span-3">
                  <label className="block text-xs text-gray-500">Match 表达式</label>
                  <input type="text" value={route.match}
                    onChange={e => setRoutes(routes.map((r, idx) => idx === ri ? { ...r, match: e.target.value } : r))}
                    placeholder="Host(`example.com`) && PathPrefix(`/api`)" className={smInputCls} />
                </div>
                <div>
                  <label className="block text-xs text-gray-500">Priority</label>
                  <input type="number" value={route.priority}
                    onChange={e => setRoutes(routes.map((r, idx) => idx === ri ? { ...r, priority: e.target.value } : r))}
                    placeholder="自动" className={smInputCls} />
                </div>
              </div>

              {/* Services */}
              <div>
                <div className="mb-1 flex items-center justify-between">
                  <label className="text-xs font-medium text-gray-500">Services</label>
                  <button type="button" onClick={() => {
                    const newRoutes = [...routes];
                    newRoutes[ri] = { ...newRoutes[ri], services: [...newRoutes[ri].services, emptyService()] };
                    setRoutes(newRoutes);
                  }} className="text-xs text-blue-600">+ 添加</button>
                </div>
                {route.services.map((svc, si) => (
                  <div key={si} className="mb-1.5 rounded border border-gray-100 bg-white p-2 space-y-1">
                    <div className="flex items-center gap-1.5">
                      <div className="flex-1">
                        <label className="block text-xs text-gray-400">名称</label>
                        <input type="text" value={svc.name} placeholder="service-name"
                          onChange={e => { const s = [...route.services]; s[si] = { ...s[si], name: e.target.value }; setRoutes(routes.map((r, idx) => idx === ri ? { ...r, services: s } : r)); }}
                          className={smInputCls} />
                      </div>
                      <div className="w-16">
                        <label className="block text-xs text-gray-400">端口</label>
                        <input type="number" value={svc.port} placeholder="80"
                          onChange={e => { const s = [...route.services]; s[si] = { ...s[si], port: e.target.value }; setRoutes(routes.map((r, idx) => idx === ri ? { ...r, services: s } : r)); }}
                          className={smInputCls} />
                      </div>
                      <div className="w-24">
                        <label className="block text-xs text-gray-400">命名空间</label>
                        <input type="text" value={svc.namespace} placeholder="同NS"
                          onChange={e => { const s = [...route.services]; s[si] = { ...s[si], namespace: e.target.value }; setRoutes(routes.map((r, idx) => idx === ri ? { ...r, services: s } : r)); }}
                          className={smInputCls} />
                      </div>
                      <div className="w-28">
                        <label className="block text-xs text-gray-400">Kind</label>
                        <select value={svc.kind}
                          onChange={e => { const s = [...route.services]; s[si] = { ...s[si], kind: e.target.value }; setRoutes(routes.map((r, idx) => idx === ri ? { ...r, services: s } : r)); }}
                          className={smInputCls}>
                          <option value="Service">Service</option>
                          <option value="TraefikService">TraefikService</option>
                        </select>
                      </div>
                      {route.services.length > 1 && (
                        <button type="button" onClick={() => {
                          const s = route.services.filter((_, idx) => idx !== si);
                          setRoutes(routes.map((r, idx) => idx === ri ? { ...r, services: s } : r));
                        }} className="text-red-400 hover:text-red-600 text-xs mt-4">✕</button>
                      )}
                    </div>
                    <div className="flex items-center gap-3 text-xs text-gray-500">
                      <label className="flex items-center gap-1">
                        <input type="checkbox" checked={svc.passHostHeader}
                          onChange={e => { const s = [...route.services]; s[si] = { ...s[si], passHostHeader: e.target.checked }; setRoutes(routes.map((r, idx) => idx === ri ? { ...r, services: s } : r)); }} />
                        passHostHeader
                      </label>
                      <label className="flex items-center gap-1">
                        <input type="checkbox" checked={svc.nativeLB}
                          onChange={e => { const s = [...route.services]; s[si] = { ...s[si], nativeLB: e.target.checked }; setRoutes(routes.map((r, idx) => idx === ri ? { ...r, services: s } : r)); }} />
                        nativeLB
                      </label>
                      <div className="flex items-center gap-1">
                        <span>scheme:</span>
                        <input type="text" value={svc.scheme} placeholder="http"
                          onChange={e => { const s = [...route.services]; s[si] = { ...s[si], scheme: e.target.value }; setRoutes(routes.map((r, idx) => idx === ri ? { ...r, services: s } : r)); }}
                          className="w-16 rounded border border-gray-300 px-1 py-0.5 text-xs" />
                      </div>
                      <div className="flex items-center gap-1">
                        <span>weight:</span>
                        <input type="number" value={svc.weight} placeholder="0"
                          onChange={e => { const s = [...route.services]; s[si] = { ...s[si], weight: e.target.value }; setRoutes(routes.map((r, idx) => idx === ri ? { ...r, services: s } : r)); }}
                          className="w-14 rounded border border-gray-300 px-1 py-0.5 text-xs" />
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {/* Middlewares */}
              <div>
                <label className="mb-1 block text-xs font-medium text-gray-500">Middlewares</label>
                {route.middlewares.map((mw, mi) => (
                  <div key={mi} className="mb-1 flex items-center gap-1.5">
                    <input type="text" value={mw.name} placeholder="middleware 名称"
                      onChange={e => {
                        const mws = [...route.middlewares]; mws[mi] = { ...mws[mi], name: e.target.value };
                        setRoutes(routes.map((r, idx) => idx === ri ? { ...r, middlewares: mws } : r));
                      }} className={smInputCls} />
                    <input type="text" value={mw.namespace} placeholder="命名空间(可选)"
                      onChange={e => {
                        const mws = [...route.middlewares]; mws[mi] = { ...mws[mi], namespace: e.target.value };
                        setRoutes(routes.map((r, idx) => idx === ri ? { ...r, middlewares: mws } : r));
                      }} className="w-28 rounded border border-gray-300 px-2 py-1 text-xs" />
                    <button type="button" onClick={() => {
                      const mws = route.middlewares.filter((_, idx) => idx !== mi);
                      setRoutes(routes.map((r, idx) => idx === ri ? { ...r, middlewares: mws } : r));
                    }} className="text-red-400 hover:text-red-600 text-xs">✕</button>
                  </div>
                ))}
                <div className="flex items-center gap-2">
                  <button type="button" onClick={() => {
                    const mws = [...route.middlewares, { name: "", namespace: "" }];
                    setRoutes(routes.map((r, idx) => idx === ri ? { ...r, middlewares: mws } : r));
                  }} className="text-xs text-blue-600">+ 手动添加</button>
                  {plugins.length > 0 && (
                    <span className="text-xs text-gray-400">或从插件中心选择：
                      {plugins.filter(p => !route.middlewares.some(m => m.name === p.name)).slice(0, 5).map(p => (
                        <button key={p.id} type="button" onClick={() => {
                          const mws = [...route.middlewares, { name: p.name, namespace: "" }];
                          setRoutes(routes.map((r, idx) => idx === ri ? { ...r, middlewares: mws } : r));
                        }} className="ml-1 rounded bg-blue-50 px-1.5 py-0.5 text-xs text-blue-600 hover:bg-blue-100">{p.name}</button>
                      ))}
                    </span>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
