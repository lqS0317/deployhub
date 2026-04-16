"use client";

import { useState } from "react";
import { useCreateEntry } from "@/hooks/use-config-center";
import { showToast } from "@/components/ui/toast";

interface CreateEntryDialogProps {
  open: boolean;
  onClose: () => void;
  serviceId: number;
  clusterId: number;
}

export function CreateEntryDialog({
  open,
  onClose,
  serviceId,
  clusterId,
}: CreateEntryDialogProps) {
  const [name, setName] = useState("");
  const [configType, setConfigType] = useState("env");
  const [format, setFormat] = useState("properties");
  const [mountPath, setMountPath] = useState("");

  const createEntry = useCreateEntry();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    createEntry.mutate(
      {
        serviceId,
        cluster_id: clusterId,
        name: name.trim(),
        config_type: configType,
        format: configType === "env" ? "properties" : format,
        mount_path: configType !== "env" ? (mountPath || `/etc/config/${name.trim()}`) : "",
      },
      {
        onSuccess: () => {
          showToast("配置条目已创建", "success");
          setName("");
          setConfigType("env");
          setFormat("properties");
          setMountPath("");
          onClose();
        },
        onError: () => showToast("创建失败", "error"),
      }
    );
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative w-full max-w-md rounded-lg bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">新建配置条目</h3>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              名称 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="例如: app-config"
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              类型
            </label>
            <select
              value={configType}
              onChange={(e) => {
                setConfigType(e.target.value);
                if (e.target.value === "env" || e.target.value === "serviceaccount") setFormat("properties");
              }}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            >
              <option value="env">Env（环境变量）</option>
              <option value="configmap">ConfigMap</option>
              <option value="secret">Secret</option>
              <option value="serviceaccount">ServiceAccount</option>
            </select>
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              格式
            </label>
            <select
              value={configType === "env" || configType === "serviceaccount" ? "properties" : format}
              onChange={(e) => setFormat(e.target.value)}
              disabled={configType === "env" || configType === "serviceaccount"}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:bg-gray-50 disabled:text-gray-500"
            >
              <option value="properties">Properties (KV)</option>
              <option value="yaml">YAML</option>
              <option value="json">JSON</option>
            </select>
            {(configType === "env" || configType === "serviceaccount") && (
              <p className="mt-1 text-xs text-gray-400">
                {configType === "serviceaccount" ? "SA 类型固定 Properties 格式（Key=注解名, Value=注解值）" : "Env 类型固定使用 Properties 格式"}
              </p>
            )}
          </div>

          {configType !== "env" && configType !== "serviceaccount" && (
            <div>
              <label className="mb-1 block text-sm font-medium text-gray-700">挂载路径</label>
              <input type="text" value={mountPath}
                onChange={(e) => setMountPath(e.target.value)}
                placeholder={`/etc/config/${name.trim() || "entry-name"}`}
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
              <p className="mt-1 text-xs text-gray-400">留空默认 /etc/config/{"{名称}"}</p>
            </div>
          )}

          <div className="flex justify-end gap-3 border-t border-gray-200 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={!name.trim() || createEntry.isPending}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {createEntry.isPending ? "创建中..." : "创建"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
