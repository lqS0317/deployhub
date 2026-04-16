"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useService, useServiceMembers, useAddMember, useRemoveMember } from "@/hooks/use-services";
import { useBuilds } from "@/hooks/use-builds";
import { useDeployments } from "@/hooks/use-deployments";
import { TriggerBuildDialog } from "@/components/build/trigger-build-dialog";
import { DeployDialog } from "@/components/deploy/deploy-dialog";
import type { ServiceMember, Build, Deployment } from "@/types";

// 构建状态徽章颜色
const BUILD_STATUS_COLORS: Record<string, string> = {
  success: "bg-green-100 text-green-700",
  succeeded: "bg-green-100 text-green-700",
  building: "bg-blue-100 text-blue-700",
  pending: "bg-yellow-100 text-yellow-700",
  failed: "bg-red-100 text-red-700",
  cancelled: "bg-gray-100 text-gray-600",
};

// 部署状态徽章颜色
const DEPLOY_STATUS_COLORS: Record<string, string> = {
  success: "bg-green-100 text-green-700",
  succeeded: "bg-green-100 text-green-700",
  deploying: "bg-blue-100 text-blue-700",
  pending_approval: "bg-yellow-100 text-yellow-700",
  approved: "bg-blue-100 text-blue-700",
  failed: "bg-red-100 text-red-700",
  rejected: "bg-red-100 text-red-700",
  rolled_back: "bg-orange-100 text-orange-700",
  expired: "bg-gray-100 text-gray-600",
};

type Tab = "info" | "members" | "builds" | "deployments";

export default function ServiceDetailPage() {
  const params = useParams();
  const router = useRouter();
  const serviceId = params.id as string;

  const [activeTab, setActiveTab] = useState<Tab>("info");
  const [showBuildDialog, setShowBuildDialog] = useState(false);
  const [showDeployDialog, setShowDeployDialog] = useState(false);
  const [newMemberUserId, setNewMemberUserId] = useState("");
  const [newMemberRole, setNewMemberRole] = useState("developer");

  const { data: service, isLoading: serviceLoading } = useService(serviceId);
  const { data: membersData } = useServiceMembers(serviceId);
  const { data: buildsData } = useBuilds({ service_id: serviceId, page: 1, page_size: 5 });
  const { data: deploymentsData } = useDeployments({ service_id: serviceId, page: 1, page_size: 5 });
  const addMember = useAddMember(serviceId);
  const removeMember = useRemoveMember(serviceId);

  const members: ServiceMember[] = Array.isArray(membersData) ? membersData : (membersData as any)?.items ?? [];
  const builds: Build[] = buildsData?.items ?? [];
  const deployments: Deployment[] = deploymentsData?.items ?? [];

  const tabs: { key: Tab; label: string }[] = [
    { key: "info", label: "基本信息" },
    { key: "members", label: "成员管理" },
    { key: "builds", label: "构建历史" },
    { key: "deployments", label: "部署历史" },
  ];

  // 添加成员
  const handleAddMember = () => {
    if (!newMemberUserId.trim()) return;
    addMember.mutate(
      { user_id: Number(newMemberUserId), role: newMemberRole },
      { onSuccess: () => setNewMemberUserId("") }
    );
  };

  // 移除成员
  const handleRemoveMember = (memberId: number) => {
    if (window.confirm("确定要移除该成员吗？")) {
      removeMember.mutate(memberId);
    }
  };

  if (serviceLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
      </div>
    );
  }

  if (!service) {
    return (
      <div className="py-20 text-center text-gray-500">
        服务不存在或已被删除
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* 返回 + 标题 */}
      <div className="flex items-center gap-4">
        <button
          onClick={() => router.push("/services")}
          className="rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
        >
          <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div className="flex-1">
          <h1 className="text-2xl font-bold text-gray-900">{service.display_name || service.name}</h1>
          <p className="mt-1 text-sm text-gray-500">{service.name}</p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => setShowBuildDialog(true)}
            className="rounded-lg border border-blue-600 px-4 py-2 text-sm font-medium text-blue-600 transition-colors hover:bg-blue-50"
          >
            触发构建
          </button>
          <button
            onClick={() => setShowDeployDialog(true)}
            className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700"
          >
            发起部署
          </button>
        </div>
      </div>

      {/* 信息概览卡片 */}
      <div className="grid grid-cols-2 gap-4 rounded-lg border border-gray-200 bg-white p-6 md:grid-cols-4">
        <InfoItem label="集群" value={service.cluster?.display_name || service.cluster?.name || "-"} />
        <InfoItem label="命名空间" value={service.namespace || "-"} />
        <InfoItem label="Git 仓库" value={service.git_repo?.name || "-"} />
        <InfoItem label="端口" value={service.port ? String(service.port) : "-"} />
      </div>

      {/* Tab 切换 */}
      <div className="border-b border-gray-200">
        <nav className="flex gap-6">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`border-b-2 pb-3 text-sm font-medium transition-colors ${
                activeTab === tab.key
                  ? "border-blue-600 text-blue-600"
                  : "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab 内容 */}
      {activeTab === "info" && (
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <DetailRow label="服务名称" value={service.name} />
            <DetailRow label="显示名称" value={service.display_name || "-"} />
            <DetailRow label="描述" value={service.description || "-"} />
            <DetailRow label="Dockerfile 路径" value={service.dockerfile_path || "-"} />
            <DetailRow label="镜像路径" value={service.image_repo || "-"} />
            <DetailRow label="副本数" value={service.replicas ? String(service.replicas) : "-"} />
            <DetailRow label="创建时间" value={service.created_at ? new Date(service.created_at).toLocaleString("zh-CN") : "-"} />
            <DetailRow label="更新时间" value={service.updated_at ? new Date(service.updated_at).toLocaleString("zh-CN") : "-"} />
          </dl>
        </div>
      )}

      {activeTab === "members" && (
        <div className="space-y-4">
          {/* 添加成员表单 */}
          <div className="flex items-end gap-3 rounded-lg border border-gray-200 bg-white p-4">
            <div className="flex-1">
              <label className="mb-1 block text-sm font-medium text-gray-700">用户 ID</label>
              <input
                type="text"
                value={newMemberUserId}
                onChange={(e) => setNewMemberUserId(e.target.value)}
                placeholder="输入用户 ID"
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div className="w-36">
              <label className="mb-1 block text-sm font-medium text-gray-700">角色</label>
              <select
                value={newMemberRole}
                onChange={(e) => setNewMemberRole(e.target.value)}
                className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="developer">Developer</option>
                <option value="viewer">Viewer</option>
                <option value="owner">Owner</option>
              </select>
            </div>
            <button
              onClick={handleAddMember}
              disabled={addMember.isPending}
              className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
            >
              添加
            </button>
          </div>

          {/* 成员列表 */}
          <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">用户</th>
                  <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">角色</th>
                  <th className="px-6 py-3 text-right text-xs font-medium uppercase text-gray-500">操作</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {members.length === 0 ? (
                  <tr>
                    <td colSpan={3} className="px-6 py-8 text-center text-sm text-gray-500">暂无成员</td>
                  </tr>
                ) : (
                  members.map((m) => (
                    <tr key={m.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 text-sm text-gray-900">{m.user?.username || String(m.user_id)}</td>
                      <td className="px-6 py-4">
                        <span className="inline-flex rounded-full bg-blue-50 px-2.5 py-0.5 text-xs font-medium text-blue-700">
                          {m.role}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        <button
                          onClick={() => handleRemoveMember(m.id)}
                          className="text-sm text-red-600 hover:text-red-800"
                        >
                          移除
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === "builds" && (
        <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">分支</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">提交</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">状态</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">时间</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {builds.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-6 py-8 text-center text-sm text-gray-500">暂无构建记录</td>
                </tr>
              ) : (
                builds.map((b) => (
                  <tr
                    key={b.id}
                    onClick={() => router.push(`/builds?id=${b.id}`)}
                    className="cursor-pointer hover:bg-gray-50"
                  >
                    <td className="px-6 py-4 text-sm text-gray-900">{b.git_branch || "-"}</td>
                    <td className="px-6 py-4 font-mono text-sm text-gray-600">
                      {b.git_commit ? b.git_commit.slice(0, 7) : "-"}
                    </td>
                    <td className="px-6 py-4">
                      <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${BUILD_STATUS_COLORS[b.status] || "bg-gray-100 text-gray-600"}`}>
                        {b.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {b.created_at ? new Date(b.created_at).toLocaleString("zh-CN") : "-"}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {activeTab === "deployments" && (
        <div className="overflow-hidden rounded-lg border border-gray-200 bg-white">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">镜像版本</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">副本数</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">状态</th>
                <th className="px-6 py-3 text-left text-xs font-medium uppercase text-gray-500">时间</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {deployments.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-6 py-8 text-center text-sm text-gray-500">暂无部署记录</td>
                </tr>
              ) : (
                deployments.map((d) => (
                  <tr
                    key={d.id}
                    onClick={() => router.push(`/deployments/${d.id}`)}
                    className="cursor-pointer hover:bg-gray-50"
                  >
                    <td className="px-6 py-4 font-mono text-sm text-gray-900">{d.image_tag || "-"}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">{d.replicas ?? "-"}</td>
                    <td className="px-6 py-4">
                      <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${DEPLOY_STATUS_COLORS[d.status] || "bg-gray-100 text-gray-600"}`}>
                        {d.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-500">
                      {d.created_at ? new Date(d.created_at).toLocaleString("zh-CN") : "-"}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      {showBuildDialog && (
        <TriggerBuildDialog
          defaultServiceId={serviceId}
          onClose={() => setShowBuildDialog(false)}
        />
      )}

      {showDeployDialog && (
        <DeployDialog
          defaultServiceId={serviceId}
          onClose={() => setShowDeployDialog(false)}
        />
      )}
    </div>
  );
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-xs text-gray-500">{label}</p>
      <p className="mt-1 text-sm font-medium text-gray-900">{value}</p>
    </div>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-sm text-gray-500">{label}</dt>
      <dd className="mt-1 text-sm font-medium text-gray-900">{value}</dd>
    </div>
  );
}
