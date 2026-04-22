"use client";

interface ApisixUpstreamFormProps {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  value: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onChange: (val: any) => void;
}

export function ApisixUpstreamForm({ value, onChange }: ApisixUpstreamFormProps) {
  const scheme = value?.scheme || "grpc";
  const lbType = value?.loadbalancer?.type || "roundrobin";
  const retries = value?.retries || 0;
  const connectTimeout = value?.timeout?.connect || "";
  const readTimeout = value?.timeout?.read || "";
  const sendTimeout = value?.timeout?.send || "";

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const update = (patch: any) => onChange({ ...value, ...patch });

  return (
    <div className="space-y-4">
      <h4 className="text-sm font-semibold text-gray-800">ApisixUpstream 配置</h4>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="mb-1 block text-xs font-medium text-gray-600">Scheme</label>
          <select
            value={scheme}
            onChange={(e) => update({ scheme: e.target.value })}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            <option value="grpc">gRPC</option>
            <option value="grpcs">gRPCs (TLS)</option>
            <option value="http">HTTP</option>
            <option value="https">HTTPS</option>
          </select>
          <p className="mt-1 text-xs text-gray-400">gRPC 代理请选择 grpc 或 grpcs</p>
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-gray-600">负载均衡</label>
          <select
            value={lbType}
            onChange={(e) => update({ loadbalancer: { ...value?.loadbalancer, type: e.target.value } })}
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            <option value="roundrobin">Round Robin</option>
            <option value="chash">Consistent Hash</option>
            <option value="ewma">EWMA</option>
            <option value="least_conn">Least Connection</option>
          </select>
        </div>

        <div>
          <label className="mb-1 block text-xs font-medium text-gray-600">重试次数</label>
          <input
            type="number"
            value={retries || ""}
            onChange={(e) => update({ retries: parseInt(e.target.value) || 0 })}
            placeholder="0"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
        </div>
      </div>

      <div>
        <h5 className="mb-2 text-xs font-semibold text-gray-700">超时配置（可选）</h5>
        <div className="grid grid-cols-3 gap-3">
          <div>
            <label className="mb-1 block text-xs text-gray-500">Connect</label>
            <input
              type="text"
              value={connectTimeout}
              onChange={(e) => update({ timeout: { ...value?.timeout, connect: e.target.value } })}
              placeholder="60s"
              className="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm font-mono focus:border-blue-500 focus:outline-none"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs text-gray-500">Read</label>
            <input
              type="text"
              value={readTimeout}
              onChange={(e) => update({ timeout: { ...value?.timeout, read: e.target.value } })}
              placeholder="60s"
              className="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm font-mono focus:border-blue-500 focus:outline-none"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs text-gray-500">Send</label>
            <input
              type="text"
              value={sendTimeout}
              onChange={(e) => update({ timeout: { ...value?.timeout, send: e.target.value } })}
              placeholder="60s"
              className="w-full rounded-md border border-gray-300 px-3 py-1.5 text-sm font-mono focus:border-blue-500 focus:outline-none"
            />
          </div>
        </div>
      </div>

      <p className="text-xs text-gray-400">
        ApisixUpstream 名称需与对应的 K8s Service 同名，APISIX 通过名称自动关联
      </p>
    </div>
  );
}
