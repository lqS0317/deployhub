"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { Approval } from "@/types";

type ApprovalStatus = "pending" | "approved" | "rejected" | "all";

// 审批中心页面：查看审批列表、通过/拒绝审批操作
export default function ApprovalsPage() {
  const queryClient = useQueryClient();
  const [statusFilter, setStatusFilter] = useState<ApprovalStatus>("pending");
  const [commentInputs, setCommentInputs] = useState<Record<number, string>>({});
  const [actioningId, setActioningId] = useState<number | null>(null);

  // 获取审批列表
  const { data: approvalsData, isLoading } = useQuery({
    queryKey: ["approvals", statusFilter],
    queryFn: async () => {
      const params = statusFilter !== "all" ? { status: statusFilter } : {};
      const res = await apiClient.get("/approvals", { params });
      return res.data;
    },
  });
  const approvals: Approval[] = approvalsData?.items ?? [];

  // 通过审批
  const approveMutation = useMutation({
    mutationFn: async ({ id, comment }: { id: number; comment: string }) => {
      const res = await apiClient.post(`/approvals/${id}/approve`, { comment });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["approvals"] });
      setActioningId(null);
    },
  });

  // 拒绝审批
  const rejectMutation = useMutation({
    mutationFn: async ({ id, comment }: { id: number; comment: string }) => {
      const res = await apiClient.post(`/approvals/${id}/reject`, { comment });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["approvals"] });
      setActioningId(null);
    },
  });

  const statusBadge = (status: string) => {
    const styles: Record<string, string> = {
      pending: "bg-yellow-100 text-yellow-800",
      approved: "bg-green-100 text-green-800",
      rejected: "bg-red-100 text-red-800",
    };
    const labels: Record<string, string> = {
      pending: "待审批",
      approved: "已通过",
      rejected: "已拒绝",
    };
    return (
      <span
        className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${
          styles[status] || "bg-gray-100 text-gray-800"
        }`}
      >
        {labels[status] || status}
      </span>
    );
  };

  const filterTabs: { key: ApprovalStatus; label: string }[] = [
    { key: "pending", label: "待审批" },
    { key: "approved", label: "已通过" },
    { key: "rejected", label: "已拒绝" },
    { key: "all", label: "全部" },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">审批中心</h1>
        <p className="mt-1 text-sm text-gray-500">审核部署与配置变更请求</p>
      </div>

      {/* 状态筛选 */}
      <div className="flex gap-1 border-b border-gray-200">
        {filterTabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setStatusFilter(tab.key)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              statusFilter === tab.key
                ? "border-blue-600 text-blue-600"
                : "border-transparent text-gray-500 hover:text-gray-700"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* 审批列表 */}
      {isLoading ? (
        <div className="text-center py-12 text-gray-500">加载中...</div>
      ) : approvals.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center">
          <p className="text-gray-500">暂无审批记录</p>
        </div>
      ) : (
        <div className="rounded-lg border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead>
              <tr className="bg-gray-50 border-b border-gray-200">
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  关联部署
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  申请人
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  审批人
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  状态
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  创建时间
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  操作
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {approvals.map((approval) => (
                <tr key={approval.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-4 py-3 text-sm text-gray-900">
                    {approval.deployment?.service?.name || `部署 #${approval.deployment_id}`}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">
                    {approval.requester?.username || approval.requester_id}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-600">
                    {approval.approver?.username || approval.approver_id || "-"}
                  </td>
                  <td className="px-4 py-3">{statusBadge(approval.status)}</td>
                  <td className="px-4 py-3 text-sm text-gray-500">
                    {new Date(approval.created_at).toLocaleString("zh-CN")}
                  </td>
                  <td className="px-4 py-3">
                    {approval.status === "pending" ? (
                      <div className="flex flex-col gap-2">
                        {actioningId === approval.id ? (
                          <>
                            <input
                              type="text"
                              value={commentInputs[approval.id] || ""}
                              onChange={(e) =>
                                setCommentInputs({
                                  ...commentInputs,
                                  [approval.id]: e.target.value,
                                })
                              }
                              placeholder="输入审批意见..."
                              className="rounded border border-gray-300 px-2 py-1 text-sm focus:border-blue-500 focus:outline-none"
                            />
                            <div className="flex gap-2">
                              <button
                                onClick={() =>
                                  approveMutation.mutate({
                                    id: approval.id,
                                    comment: commentInputs[approval.id] || "",
                                  })
                                }
                                disabled={approveMutation.isPending}
                                className="rounded bg-green-600 px-2.5 py-1 text-xs font-medium text-white hover:bg-green-700 disabled:opacity-50"
                              >
                                通过
                              </button>
                              <button
                                onClick={() =>
                                  rejectMutation.mutate({
                                    id: approval.id,
                                    comment: commentInputs[approval.id] || "",
                                  })
                                }
                                disabled={rejectMutation.isPending}
                                className="rounded bg-red-600 px-2.5 py-1 text-xs font-medium text-white hover:bg-red-700 disabled:opacity-50"
                              >
                                拒绝
                              </button>
                              <button
                                onClick={() => setActioningId(null)}
                                className="rounded border border-gray-300 px-2.5 py-1 text-xs text-gray-600 hover:bg-gray-50"
                              >
                                取消
                              </button>
                            </div>
                          </>
                        ) : (
                          <button
                            onClick={() => setActioningId(approval.id)}
                            className="rounded bg-blue-600 px-2.5 py-1 text-xs font-medium text-white hover:bg-blue-700"
                          >
                            审批
                          </button>
                        )}
                      </div>
                    ) : (
                      <span className="text-xs text-gray-400">
                        {approval.comment || "-"}
                      </span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
