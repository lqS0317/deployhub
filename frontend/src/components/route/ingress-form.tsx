"use client";

import { useState, useEffect } from "react";

interface AnnotationRow {
  key: string;
  value: string;
}

interface TlsRow {
  hosts: string;
  secretName: string;
}

interface IngressPath {
  path: string;
  pathType: string;
  backendService: string;
  backendPort: string;
}

interface IngressRule {
  host: string;
  paths: IngressPath[];
}

interface IngressConfig {
  ingressClassName: string;
  annotations: Record<string, string>;
  tls: Array<{ hosts: string[]; secretName: string }>;
  rules: Array<{
    host: string;
    paths: Array<{
      path: string;
      pathType: string;
      backendService: string;
      backendPort: number;
    }>;
  }>;
}

interface IngressFormProps {
  value: IngressConfig;
  onChange: (config: IngressConfig) => void;
}

export function IngressForm({ value, onChange }: IngressFormProps) {
  const [ingressClassName, setIngressClassName] = useState(
    value.ingressClassName || ""
  );
  const [annotations, setAnnotations] = useState<AnnotationRow[]>(() => {
    const a = value.annotations || {};
    const rows = Object.entries(a).map(([k, v]) => ({ key: k, value: v }));
    return rows.length > 0 ? rows : [{ key: "", value: "" }];
  });
  const [tls, setTls] = useState<TlsRow[]>(() => {
    const t = value.tls || [];
    return t.length > 0
      ? t.map((r) => ({ hosts: r.hosts?.join(", ") || "", secretName: r.secretName }))
      : [];
  });
  const [rules, setRules] = useState<IngressRule[]>(() => {
    const r = value.rules || [];
    if (r.length > 0) {
      return r.map((rule) => ({
        host: rule.host,
        paths: rule.paths.map((p) => ({
          path: p.path,
          pathType: p.pathType || "Prefix",
          backendService: p.backendService,
          backendPort: String(p.backendPort || ""),
        })),
      }));
    }
    return [
      {
        host: "",
        paths: [
          { path: "/", pathType: "Prefix", backendService: "", backendPort: "" },
        ],
      },
    ];
  });

  useEffect(() => {
    const annot: Record<string, string> = {};
    annotations.forEach((a) => {
      if (a.key.trim()) annot[a.key.trim()] = a.value.trim();
    });
    onChange({
      ingressClassName,
      annotations: annot,
      tls: tls
        .filter((t) => t.hosts.trim())
        .map((t) => ({
          hosts: t.hosts.split(",").map((h) => h.trim()),
          secretName: t.secretName,
        })),
      rules: rules.map((r) => ({
        host: r.host,
        paths: r.paths.map((p) => ({
          path: p.path,
          pathType: p.pathType,
          backendService: p.backendService,
          backendPort: Number(p.backendPort) || 80,
        })),
      })),
    });
  }, [ingressClassName, annotations, tls, rules]);

  const inputCls =
    "w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500";

  return (
    <div className="space-y-4">
      <div>
        <label className="mb-1 block text-sm font-medium text-gray-700">
          IngressClassName
        </label>
        <input
          type="text"
          value={ingressClassName}
          onChange={(e) => setIngressClassName(e.target.value)}
          placeholder="nginx"
          className={inputCls}
        />
      </div>

      {/* Annotations */}
      <div>
        <div className="mb-1 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-700">
            Annotations
          </label>
          <button
            type="button"
            onClick={() =>
              setAnnotations([...annotations, { key: "", value: "" }])
            }
            className="text-xs text-blue-600 hover:text-blue-800"
          >
            + 添加
          </button>
        </div>
        <div className="space-y-2">
          {annotations.map((a, i) => (
            <div key={i} className="flex items-center gap-2">
              <input
                type="text"
                value={a.key}
                onChange={(e) =>
                  setAnnotations(
                    annotations.map((r, idx) =>
                      idx === i ? { ...r, key: e.target.value } : r
                    )
                  )
                }
                placeholder="key"
                className={inputCls}
              />
              <span className="text-gray-400">=</span>
              <input
                type="text"
                value={a.value}
                onChange={(e) =>
                  setAnnotations(
                    annotations.map((r, idx) =>
                      idx === i ? { ...r, value: e.target.value } : r
                    )
                  )
                }
                placeholder="value"
                className={inputCls}
              />
              {annotations.length > 1 && (
                <button
                  type="button"
                  onClick={() =>
                    setAnnotations(annotations.filter((_, idx) => idx !== i))
                  }
                  className="text-red-500 hover:text-red-700 text-lg leading-none"
                >
                  ×
                </button>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* TLS */}
      <div>
        <div className="mb-1 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-700">TLS</label>
          <button
            type="button"
            onClick={() => setTls([...tls, { hosts: "", secretName: "" }])}
            className="text-xs text-blue-600 hover:text-blue-800"
          >
            + 添加
          </button>
        </div>
        <div className="space-y-2">
          {tls.map((t, i) => (
            <div key={i} className="flex items-center gap-2">
              <input
                type="text"
                value={t.hosts}
                onChange={(e) =>
                  setTls(
                    tls.map((r, idx) =>
                      idx === i ? { ...r, hosts: e.target.value } : r
                    )
                  )
                }
                placeholder="hosts（逗号分隔）"
                className={inputCls}
              />
              <input
                type="text"
                value={t.secretName}
                onChange={(e) =>
                  setTls(
                    tls.map((r, idx) =>
                      idx === i ? { ...r, secretName: e.target.value } : r
                    )
                  )
                }
                placeholder="secretName"
                className={inputCls}
              />
              <button
                type="button"
                onClick={() => setTls(tls.filter((_, idx) => idx !== i))}
                className="text-red-500 hover:text-red-700 text-lg leading-none"
              >
                ×
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Rules */}
      <div>
        <div className="mb-1 flex items-center justify-between">
          <label className="text-sm font-medium text-gray-700">Rules</label>
          <button
            type="button"
            onClick={() =>
              setRules([
                ...rules,
                {
                  host: "",
                  paths: [
                    {
                      path: "/",
                      pathType: "Prefix",
                      backendService: "",
                      backendPort: "",
                    },
                  ],
                },
              ])
            }
            className="text-xs text-blue-600 hover:text-blue-800"
          >
            + 添加 Rule
          </button>
        </div>
        <div className="space-y-3">
          {rules.map((rule, ri) => (
            <div
              key={ri}
              className="rounded-md border border-gray-200 bg-gray-50 p-3"
            >
              <div className="mb-2 flex items-center justify-between">
                <input
                  type="text"
                  value={rule.host}
                  onChange={(e) =>
                    setRules(
                      rules.map((r, idx) =>
                        idx === ri ? { ...r, host: e.target.value } : r
                      )
                    )
                  }
                  placeholder="Host"
                  className={inputCls}
                />
                {rules.length > 1 && (
                  <button
                    type="button"
                    onClick={() =>
                      setRules(rules.filter((_, idx) => idx !== ri))
                    }
                    className="ml-2 text-red-500 hover:text-red-700 text-lg leading-none"
                  >
                    ×
                  </button>
                )}
              </div>
              <div className="space-y-2 pl-3">
                {rule.paths.map((p, pi) => (
                  <div key={pi} className="flex items-center gap-2">
                    <input
                      type="text"
                      value={p.path}
                      onChange={(e) => {
                        const newPaths = [...rule.paths];
                        newPaths[pi] = { ...newPaths[pi], path: e.target.value };
                        setRules(
                          rules.map((r, idx) =>
                            idx === ri ? { ...r, paths: newPaths } : r
                          )
                        );
                      }}
                      placeholder="path"
                      className={inputCls}
                    />
                    <select
                      value={p.pathType}
                      onChange={(e) => {
                        const newPaths = [...rule.paths];
                        newPaths[pi] = { ...newPaths[pi], pathType: e.target.value };
                        setRules(
                          rules.map((r, idx) =>
                            idx === ri ? { ...r, paths: newPaths } : r
                          )
                        );
                      }}
                      className={inputCls}
                    >
                      <option value="Prefix">Prefix</option>
                      <option value="Exact">Exact</option>
                      <option value="ImplementationSpecific">
                        ImplementationSpecific
                      </option>
                    </select>
                    <input
                      type="text"
                      value={p.backendService}
                      onChange={(e) => {
                        const newPaths = [...rule.paths];
                        newPaths[pi] = {
                          ...newPaths[pi],
                          backendService: e.target.value,
                        };
                        setRules(
                          rules.map((r, idx) =>
                            idx === ri ? { ...r, paths: newPaths } : r
                          )
                        );
                      }}
                      placeholder="backend service"
                      className={inputCls}
                    />
                    <input
                      type="number"
                      value={p.backendPort}
                      onChange={(e) => {
                        const newPaths = [...rule.paths];
                        newPaths[pi] = {
                          ...newPaths[pi],
                          backendPort: e.target.value,
                        };
                        setRules(
                          rules.map((r, idx) =>
                            idx === ri ? { ...r, paths: newPaths } : r
                          )
                        );
                      }}
                      placeholder="port"
                      className={inputCls}
                    />
                    {rule.paths.length > 1 && (
                      <button
                        type="button"
                        onClick={() => {
                          const newPaths = rule.paths.filter(
                            (_, idx) => idx !== pi
                          );
                          setRules(
                            rules.map((r, idx) =>
                              idx === ri ? { ...r, paths: newPaths } : r
                            )
                          );
                        }}
                        className="text-red-500 hover:text-red-700 text-lg leading-none"
                      >
                        ×
                      </button>
                    )}
                  </div>
                ))}
                <button
                  type="button"
                  onClick={() => {
                    const newPaths = [
                      ...rule.paths,
                      {
                        path: "/",
                        pathType: "Prefix",
                        backendService: "",
                        backendPort: "",
                      },
                    ];
                    setRules(
                      rules.map((r, idx) =>
                        idx === ri ? { ...r, paths: newPaths } : r
                      )
                    );
                  }}
                  className="text-xs text-blue-600 hover:text-blue-800"
                >
                  + 添加 Path
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
