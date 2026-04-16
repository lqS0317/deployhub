"use client";

import type { Cluster } from "@/types";

interface EnvSidebarProps {
  clusters: Cluster[];
  selectedClusterId: number;
  onSelectCluster: (id: number) => void;
}

const envBadge = (env: string) => {
  const colors: Record<string, string> = {
    devnet: "bg-green-100 text-green-800",
    qanet: "bg-yellow-100 text-yellow-800",
    testnet: "bg-indigo-100 text-indigo-800",
    mainnet: "bg-red-100 text-red-800",
  };
  return (
    <span
      className={`inline-flex rounded px-1.5 py-0.5 text-[10px] font-medium ${
        colors[env] || "bg-gray-100 text-gray-700"
      }`}
    >
      {env}
    </span>
  );
};

export function EnvSidebar({
  clusters,
  selectedClusterId,
  onSelectCluster,
}: EnvSidebarProps) {
  return (
    <div className="flex h-full w-56 flex-col border-r border-gray-200 bg-gray-50">
      <div className="px-3 py-3">
        <p className="mb-2 text-xs font-medium uppercase tracking-wider text-gray-500">
          环境
        </p>
      </div>

      <div className="flex-1 overflow-y-auto px-2 py-1">
        {clusters.length === 0 ? (
          <p className="px-1 text-xs text-gray-400">暂无可用集群</p>
        ) : (
          <ul className="space-y-0.5">
            {clusters.map((c) => (
              <li key={c.id}>
                <button
                  onClick={() => onSelectCluster(c.id)}
                  className={`flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-left text-sm transition-colors ${
                    selectedClusterId === c.id
                      ? "bg-blue-50 text-blue-700 font-medium"
                      : "text-gray-700 hover:bg-gray-100"
                  }`}
                >
                  <span className="flex-1 truncate">
                    {c.display_name || c.name}
                  </span>
                  {envBadge(c.env)}
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
