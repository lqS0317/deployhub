"use client";

import { useState } from "react";
import type { Cluster } from "@/types";

interface KeyValuePair {
  key: string;
  value: string;
}

interface EnvValuesEditorProps {
  clusters: Cluster[];
  configType: "configmap" | "secret";
  values: Record<string, KeyValuePair[]>;
  onSave: (clusterId: string, pairs: KeyValuePair[]) => void;
  saving?: boolean;
}

// 按集群管理环境变量键值对编辑器
export function EnvValuesEditor({
  clusters,
  configType,
  values,
  onSave,
  saving = false,
}: EnvValuesEditorProps) {
  const [selectedCluster, setSelectedCluster] = useState<string>(
    clusters[0]?.id?.toString() || ""
  );
  const [pairs, setPairs] = useState<KeyValuePair[]>(
    values[selectedCluster] || [{ key: "", value: "" }]
  );
  const [showValues, setShowValues] = useState<Record<number, boolean>>({});

  const isSecret = configType === "secret";

  const handleClusterChange = (clusterId: string) => {
    setSelectedCluster(clusterId);
    setPairs(values[clusterId] || [{ key: "", value: "" }]);
    setShowValues({});
  };

  const updatePair = (index: number, field: "key" | "value", val: string) => {
    const updated = [...pairs];
    updated[index] = { ...updated[index], [field]: val };
    setPairs(updated);
  };

  const addRow = () => {
    setPairs([...pairs, { key: "", value: "" }]);
  };

  const removeRow = (index: number) => {
    setPairs(pairs.filter((_, i) => i !== index));
  };

  const toggleShowValue = (index: number) => {
    setShowValues((prev) => ({ ...prev, [index]: !prev[index] }));
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <label className="text-sm font-medium text-gray-700">目标集群</label>
          <select
            value={selectedCluster}
            onChange={(e) => handleClusterChange(e.target.value)}
            className="rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            {clusters.map((c) => (
              <option key={c.id} value={c.id}>
                {c.display_name || c.name}
              </option>
            ))}
          </select>
        </div>
        <button
          onClick={() => onSave(selectedCluster, pairs)}
          disabled={saving}
          className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 transition-colors"
        >
          {saving ? "保存中..." : "保存当前集群配置"}
        </button>
      </div>

      {/* 键值对表格 */}
      <div className="rounded-lg border border-gray-200 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-4 py-2.5 text-left text-xs font-medium text-gray-500 uppercase">
                键
              </th>
              <th className="px-4 py-2.5 text-left text-xs font-medium text-gray-500 uppercase">
                值
              </th>
              <th className="w-20 px-4 py-2.5 text-center text-xs font-medium text-gray-500 uppercase">
                操作
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {pairs.map((pair, index) => (
              <tr key={index} className="hover:bg-gray-50">
                <td className="px-4 py-2">
                  <input
                    type="text"
                    value={pair.key}
                    onChange={(e) => updatePair(index, "key", e.target.value)}
                    placeholder="KEY_NAME"
                    className="w-full rounded border border-gray-300 px-2 py-1 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                  />
                </td>
                <td className="px-4 py-2">
                  <div className="flex items-center gap-2">
                    <input
                      type={isSecret && !showValues[index] ? "password" : "text"}
                      value={pair.value}
                      onChange={(e) =>
                        updatePair(index, "value", e.target.value)
                      }
                      placeholder={isSecret ? "••••••" : "value"}
                      className="w-full rounded border border-gray-300 px-2 py-1 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                    />
                    {isSecret && (
                      <button
                        onClick={() => toggleShowValue(index)}
                        className="shrink-0 text-xs text-gray-500 hover:text-gray-700"
                        title={showValues[index] ? "隐藏" : "显示"}
                      >
                        {showValues[index] ? "🙈" : "👁"}
                      </button>
                    )}
                  </div>
                </td>
                <td className="px-4 py-2 text-center">
                  <button
                    onClick={() => removeRow(index)}
                    className="text-red-500 hover:text-red-700 text-sm"
                    title="删除"
                  >
                    ✕
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <button
        onClick={addRow}
        className="self-start rounded-md border border-dashed border-gray-300 px-3 py-1.5 text-sm text-gray-600 hover:border-gray-400 hover:text-gray-800 transition-colors"
      >
        + 添加键值对
      </button>
    </div>
  );
}
