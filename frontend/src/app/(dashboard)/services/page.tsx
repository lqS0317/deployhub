"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useServices, useDeleteService } from "@/hooks/use-services";
import { ServiceCreateDialog, ServiceFormDialog } from "@/components/service/service-create-dialog";
import type { Service } from "@/types";

export default function ServicesPage() {
  const router = useRouter();
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [showCreate, setShowCreate] = useState(false);
  const [editingService, setEditingService] = useState<Service | null>(null);
  const pageSize = 20;

  const { data, isLoading } = useServices({ page, page_size: pageSize, search });
  const deleteService = useDeleteService();

  const services: Service[] = data?.items ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / pageSize);

  // 删除服务确认
  const handleDelete = (e: React.MouseEvent, id: number) => {
    e.stopPropagation();
    if (window.confirm("确定要删除该服务吗？此操作不可撤销。")) {
      deleteService.mutate(id);
    }
  };

  return (
    <div className="space-y-6">
      {/* 页面标题和操作栏 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">服务管理</h1>
          <p className="mt-1 text-sm text-gray-500">管理所有微服务的生命周期</p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
        >
          创建服务
        </button>
      </div>

      {/* 搜索框 */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <input
            type="text"
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
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
      </div>

      {/* 服务列表表格 */}
      <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                名称
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                类型
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                语言
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500">
                Git 仓库
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium uppercase tracking-wider text-gray-500">
                操作
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {isLoading ? (
              <tr>
                <td colSpan={5} className="px-6 py-12 text-center">
                  <div className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
                  <p className="mt-2 text-sm text-gray-500">加载中...</p>
                </td>
              </tr>
            ) : services.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-6 py-12 text-center text-sm text-gray-500">
                  暂无服务数据
                </td>
              </tr>
            ) : (
              services.map((svc) => (
                <tr
                  key={svc.id}
                  onClick={() => router.push(`/services/${svc.id}`)}
                  className="cursor-pointer transition-colors hover:bg-gray-50"
                >
                  <td className="whitespace-nowrap px-6 py-4">
                    <div className="flex items-center gap-3">
                      <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
                        <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
                        </svg>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-gray-900">{svc.name}</p>
                        {svc.display_name && (
                          <p className="text-xs text-gray-500">{svc.display_name}</p>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {svc.service_type ? (
                      <span className="inline-flex rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700">
                        {svc.service_type}
                      </span>
                    ) : "-"}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {svc.language ? (
                      <span className="inline-flex items-center gap-1 text-xs">
                        <span className="font-medium">{svc.language}</span>
                        {svc.language_version && <span className="text-gray-400">{svc.language_version}</span>}
                      </span>
                    ) : "-"}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-sm text-gray-600">
                    {svc.git_repo?.name || "-"}
                  </td>
                  <td className="whitespace-nowrap px-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-3">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setEditingService(svc);
                        }}
                        className="text-sm text-blue-600 transition-colors hover:text-blue-800"
                      >
                        编辑
                      </button>
                      <button
                        onClick={(e) => handleDelete(e, svc.id)}
                        className="text-sm text-red-600 transition-colors hover:text-red-800"
                      >
                        删除
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* 分页控件 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-gray-500">
            共 {total} 条记录，第 {page}/{totalPages} 页
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              上一页
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="rounded-lg border border-gray-300 px-3 py-1.5 text-sm text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              下一页
            </button>
          </div>
        </div>
      )}

      {/* 创建服务弹窗 */}
      {showCreate && <ServiceCreateDialog onClose={() => setShowCreate(false)} />}

      {/* 编辑服务弹窗 */}
      {editingService && (
        <ServiceFormDialog
          service={editingService}
          onClose={() => setEditingService(null)}
        />
      )}
    </div>
  );
}
