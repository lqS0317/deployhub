"use client";

import { useState } from "react";

interface ApisixUpstreamFormProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  value: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onChange: (val: any) => void;
}

export function ApisixUpstreamForm({ value, onChange }: ApisixUpstreamFormProps) {
  const [showHealthCheck, setShowHealthCheck] = useState(!!value?.healthCheck);
  const [showPassive, setShowPassive] = useState(!!value?.healthCheck?.passive);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const update = (patch: any) => onChange({ ...value, ...patch });

  const scheme = value?.scheme || "";
  const lbType = value?.loadbalancer?.type || "roundrobin";
  const retries = value?.retries || 0;
  const connectTimeout = value?.timeout?.connect || "";
  const readTimeout = value?.timeout?.read || "";
  const sendTimeout = value?.timeout?.send || "";

  const hc = value?.healthCheck || {};
  const active = hc?.active || {};
  const passive = hc?.passive || {};

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const updateHC = (path: string, val: any) => {
    const newHC = { ...hc };
    const parts = path.split(".");
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let obj: any = newHC;
    for (let i = 0; i < parts.length - 1; i++) {
      if (!obj[parts[i]]) obj[parts[i]] = {};
      obj = obj[parts[i]];
    }
    obj[parts[parts.length - 1]] = val;
    update({ healthCheck: newHC });
  };

  const inputCls = "w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500";
  const selectCls = inputCls;
  const labelCls = "mb-1 block text-xs font-medium text-gray-600";
  const hintCls = "mt-1 text-xs text-gray-400";

  return (
    <div className="space-y-5">
      <h4 className="text-sm font-semibold text-gray-800">ApisixUpstream 配置</h4>

      {/* 基础配置 */}
      <div className="grid grid-cols-3 gap-4">
        <div>
          <label className={labelCls}>Scheme</label>
          <select value={scheme} onChange={(e) => update({ scheme: e.target.value })} className={selectCls}>
            <option value="">默认 (HTTP)</option>
            <option value="http">HTTP</option>
            <option value="https">HTTPS</option>
            <option value="grpc">gRPC</option>
            <option value="grpcs">gRPCs (TLS)</option>
          </select>
          <p className={hintCls}>上游服务协议</p>
        </div>
        <div>
          <label className={labelCls}>负载均衡</label>
          <select value={lbType} onChange={(e) => update({ loadbalancer: { ...value?.loadbalancer, type: e.target.value } })} className={selectCls}>
            <option value="roundrobin">Round Robin</option>
            <option value="chash">Consistent Hash</option>
            <option value="ewma">EWMA</option>
            <option value="least_conn">Least Connection</option>
          </select>
        </div>
        <div>
          <label className={labelCls}>重试次数</label>
          <input type="number" value={retries || ""} onChange={(e) => update({ retries: parseInt(e.target.value) || 0 })} placeholder="0" className={inputCls} />
        </div>
      </div>

      {/* 超时 */}
      <div>
        <h5 className="mb-2 text-xs font-semibold text-gray-700">超时配置</h5>
        <div className="grid grid-cols-3 gap-3">
          <div>
            <label className="mb-1 block text-xs text-gray-500">Connect</label>
            <input type="text" value={connectTimeout} onChange={(e) => update({ timeout: { ...value?.timeout, connect: e.target.value } })} placeholder="60s" className={inputCls} />
          </div>
          <div>
            <label className="mb-1 block text-xs text-gray-500">Read</label>
            <input type="text" value={readTimeout} onChange={(e) => update({ timeout: { ...value?.timeout, read: e.target.value } })} placeholder="60s" className={inputCls} />
          </div>
          <div>
            <label className="mb-1 block text-xs text-gray-500">Send</label>
            <input type="text" value={sendTimeout} onChange={(e) => update({ timeout: { ...value?.timeout, send: e.target.value } })} placeholder="60s" className={inputCls} />
          </div>
        </div>
      </div>

      {/* 健康检查 */}
      <div className="border-t border-gray-200 pt-4">
        <label className="flex items-center gap-2 text-sm font-medium text-gray-700">
          <input
            type="checkbox"
            checked={showHealthCheck}
            onChange={(e) => {
              setShowHealthCheck(e.target.checked);
              if (!e.target.checked) update({ healthCheck: undefined });
            }}
            className="rounded border-gray-300"
          />
          启用健康检查
        </label>
      </div>

      {showHealthCheck && (
        <div className="space-y-4 rounded-md border border-gray-200 bg-gray-50 p-4">
          {/* Active 健康检查 */}
          <h5 className="text-xs font-semibold text-gray-700">主动健康检查 (Active)</h5>
          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className={labelCls}>类型</label>
              <select value={active.type || "http"} onChange={(e) => updateHC("active.type", e.target.value)} className={selectCls}>
                <option value="http">HTTP</option>
                <option value="https">HTTPS</option>
                <option value="tcp">TCP</option>
              </select>
            </div>
            <div>
              <label className={labelCls}>检查路径</label>
              <input type="text" value={active.httpPath || ""} onChange={(e) => updateHC("active.httpPath", e.target.value)} placeholder="/health" className={inputCls} />
            </div>
            <div>
              <label className={labelCls}>检查端口</label>
              <input type="number" value={active.port || ""} onChange={(e) => updateHC("active.port", parseInt(e.target.value) || undefined)} placeholder="与服务端口一致" className={inputCls} />
            </div>
            <div>
              <label className={labelCls}>Host</label>
              <input type="text" value={active.host || ""} onChange={(e) => updateHC("active.host", e.target.value)} placeholder="留空用上游 Host" className={inputCls} />
            </div>
            <div>
              <label className={labelCls}>超时 (秒)</label>
              <input type="number" value={active.timeout || ""} onChange={(e) => updateHC("active.timeout", parseInt(e.target.value) || undefined)} placeholder="1" className={inputCls} />
            </div>
            <div>
              <label className={labelCls}>并发数</label>
              <input type="number" value={active.concurrency || ""} onChange={(e) => updateHC("active.concurrency", parseInt(e.target.value) || undefined)} placeholder="10" className={inputCls} />
            </div>
          </div>

          {/* Healthy 条件 */}
          <div className="grid grid-cols-2 gap-4">
            <div className="rounded border border-green-200 bg-green-50 p-3">
              <h6 className="mb-2 text-xs font-semibold text-green-700">Healthy 判定</h6>
              <div className="space-y-2">
                <div>
                  <label className="block text-xs text-gray-500">检查间隔</label>
                  <input type="text" value={active.healthy?.interval || ""} onChange={(e) => updateHC("active.healthy.interval", e.target.value)} placeholder="1s" className={inputCls} />
                </div>
                <div>
                  <label className="block text-xs text-gray-500">连续成功次数</label>
                  <input type="number" value={active.healthy?.successes || ""} onChange={(e) => updateHC("active.healthy.successes", parseInt(e.target.value) || undefined)} placeholder="2" className={inputCls} />
                </div>
                <div>
                  <label className="block text-xs text-gray-500">HTTP 状态码 (逗号分隔)</label>
                  <input
                    type="text"
                    value={(active.healthy?.httpCodes || []).join(",")}
                    onChange={(e) => updateHC("active.healthy.httpCodes", e.target.value ? e.target.value.split(",").map(Number).filter(Boolean) : undefined)}
                    placeholder="200,302"
                    className={inputCls}
                  />
                </div>
              </div>
            </div>
            <div className="rounded border border-red-200 bg-red-50 p-3">
              <h6 className="mb-2 text-xs font-semibold text-red-700">Unhealthy 判定</h6>
              <div className="space-y-2">
                <div>
                  <label className="block text-xs text-gray-500">检查间隔</label>
                  <input type="text" value={active.unhealthy?.interval || ""} onChange={(e) => updateHC("active.unhealthy.interval", e.target.value)} placeholder="1s" className={inputCls} />
                </div>
                <div>
                  <label className="block text-xs text-gray-500">HTTP 失败次数</label>
                  <input type="number" value={active.unhealthy?.httpFailures || ""} onChange={(e) => updateHC("active.unhealthy.httpFailures", parseInt(e.target.value) || undefined)} placeholder="5" className={inputCls} />
                </div>
                <div>
                  <label className="block text-xs text-gray-500">TCP 失败次数</label>
                  <input type="number" value={active.unhealthy?.tcpFailures || ""} onChange={(e) => updateHC("active.unhealthy.tcpFailures", parseInt(e.target.value) || undefined)} placeholder="2" className={inputCls} />
                </div>
                <div>
                  <label className="block text-xs text-gray-500">超时次数</label>
                  <input type="number" value={active.unhealthy?.timeout || ""} onChange={(e) => updateHC("active.unhealthy.timeout", parseInt(e.target.value) || undefined)} placeholder="3" className={inputCls} />
                </div>
                <div>
                  <label className="block text-xs text-gray-500">HTTP 状态码 (逗号分隔)</label>
                  <input
                    type="text"
                    value={(active.unhealthy?.httpCodes || []).join(",")}
                    onChange={(e) => updateHC("active.unhealthy.httpCodes", e.target.value ? e.target.value.split(",").map(Number).filter(Boolean) : undefined)}
                    placeholder="429,500,503"
                    className={inputCls}
                  />
                </div>
              </div>
            </div>
          </div>

          {/* Passive 健康检查 */}
          <div className="border-t border-gray-200 pt-3">
            <label className="flex items-center gap-2 text-xs font-medium text-gray-700">
              <input
                type="checkbox"
                checked={showPassive}
                onChange={(e) => {
                  setShowPassive(e.target.checked);
                  if (!e.target.checked) {
                    const newHC = { ...hc };
                    delete newHC.passive;
                    update({ healthCheck: newHC });
                  }
                }}
                className="rounded border-gray-300"
              />
              启用被动健康检查 (Passive)
            </label>
          </div>

          {showPassive && (
            <div className="space-y-3">
              <div>
                <label className={labelCls}>类型</label>
                <select value={passive.type || "http"} onChange={(e) => updateHC("passive.type", e.target.value)} className={selectCls + " w-40"}>
                  <option value="http">HTTP</option>
                  <option value="https">HTTPS</option>
                  <option value="tcp">TCP</option>
                </select>
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="rounded border border-green-200 bg-green-50 p-3">
                  <h6 className="mb-2 text-xs font-semibold text-green-700">Healthy 判定</h6>
                  <div className="space-y-2">
                    <div>
                      <label className="block text-xs text-gray-500">连续成功次数</label>
                      <input type="number" value={passive.healthy?.successes || ""} onChange={(e) => updateHC("passive.healthy.successes", parseInt(e.target.value) || undefined)} placeholder="5" className={inputCls} />
                    </div>
                    <div>
                      <label className="block text-xs text-gray-500">HTTP 状态码</label>
                      <input
                        type="text"
                        value={(passive.healthy?.httpCodes || []).join(",")}
                        onChange={(e) => updateHC("passive.healthy.httpCodes", e.target.value ? e.target.value.split(",").map(Number).filter(Boolean) : undefined)}
                        placeholder="200,201,301"
                        className={inputCls}
                      />
                    </div>
                  </div>
                </div>
                <div className="rounded border border-red-200 bg-red-50 p-3">
                  <h6 className="mb-2 text-xs font-semibold text-red-700">Unhealthy 判定</h6>
                  <div className="space-y-2">
                    <div>
                      <label className="block text-xs text-gray-500">HTTP 失败次数</label>
                      <input type="number" value={passive.unhealthy?.httpFailures || ""} onChange={(e) => updateHC("passive.unhealthy.httpFailures", parseInt(e.target.value) || undefined)} placeholder="5" className={inputCls} />
                    </div>
                    <div>
                      <label className="block text-xs text-gray-500">TCP 失败次数</label>
                      <input type="number" value={passive.unhealthy?.tcpFailures || ""} onChange={(e) => updateHC("passive.unhealthy.tcpFailures", parseInt(e.target.value) || undefined)} placeholder="2" className={inputCls} />
                    </div>
                    <div>
                      <label className="block text-xs text-gray-500">超时次数</label>
                      <input type="number" value={passive.unhealthy?.timeout || ""} onChange={(e) => updateHC("passive.unhealthy.timeout", parseInt(e.target.value) || undefined)} placeholder="7" className={inputCls} />
                    </div>
                    <div>
                      <label className="block text-xs text-gray-500">HTTP 状态码</label>
                      <input
                        type="text"
                        value={(passive.unhealthy?.httpCodes || []).join(",")}
                        onChange={(e) => updateHC("passive.unhealthy.httpCodes", e.target.value ? e.target.value.split(",").map(Number).filter(Boolean) : undefined)}
                        placeholder="429,500,503"
                        className={inputCls}
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      <p className="text-xs text-gray-400">
        ApisixUpstream 名称需与对应的 K8s Service 同名，APISIX 通过名称自动关联
      </p>
    </div>
  );
}
