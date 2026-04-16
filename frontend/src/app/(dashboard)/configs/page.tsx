"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useServices } from "@/hooks/use-services";
import type { Service } from "@/types";

export default function ConfigCenterPage() {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const { data, isLoading } = useServices({ page: 1, page_size: 200 });

  const services: Service[] = (data?.items ?? []).filter(
    (s) =>
      !search ||
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.display_name?.toLowerCase().includes(search.toLowerCase())
  );

  const deployTypeBadge = (type: string) => {
    const styles =
      type === "helm"
        ? "bg-purple-100 text-purple-800"
        : "bg-blue-100 text-blue-800";
    return (
      <span
        className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${styles}`}
      >
        {type}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">配置中心</h1>
        <p className="mt-1 text-sm text-gray-500">
          Apollo 风格配置管理，按服务查看和编辑配置集
        </p>
      </div>

      <div className="relative max-w-md">
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="按名称搜索服务..."
          className="w-full rounded-lg border border-gray-300 bg-white px-4 py-2 pl-10 text-sm text-gray-900 placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
        <svg
          className="absolute left-3 top-2.5 h-4 w-4 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      </div>

      {isLoading ? (
        <div className="py-12 text-center">
          <div className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
          <p className="mt-2 text-sm text-gray-500">加载中...</p>
        </div>
      ) : services.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center">
          <p className="text-gray-500">
            {search ? "未找到匹配的服务" : "暂无服务"}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {services.map((svc) => (
            <div
              key={svc.id}
              onClick={() => router.push(`/configs/${svc.id}`)}
              className="cursor-pointer rounded-lg border border-gray-200 bg-white p-5 transition-all hover:border-blue-300 hover:shadow-md"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
                    <svg
                      className="h-5 w-5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                      />
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                    </svg>
                  </div>
                  <div>
                    <p className="text-sm font-semibold text-gray-900">
                      {svc.name}
                    </p>
                    {svc.display_name && (
                      <p className="text-xs text-gray-500">
                        {svc.display_name}
                      </p>
                    )}
                  </div>
                </div>
                {deployTypeBadge(svc.deploy_type)}
              </div>
              {svc.description && (
                <p className="mt-3 text-xs text-gray-500 line-clamp-2">
                  {svc.description}
                </p>
              )}
              <div className="mt-3 flex items-center gap-3 text-xs text-gray-400">
                {svc.language && <span>{svc.language}</span>}
                {svc.git_repo?.name && (
                  <>
                    <span>·</span>
                    <span>{svc.git_repo.name}</span>
                  </>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
