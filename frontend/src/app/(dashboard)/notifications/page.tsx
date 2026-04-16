"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { showToast } from "@/components/ui/toast";
import type { Notification, NotificationChannel, Service } from "@/types";

type MainTab = "inbox" | "rules" | "logs";
type ReadFilter = "all" | "unread" | "read";

const EVENT_LABELS: Record<string, string> = {
  all: "All（默认）",
  build_success: "构建成功", build_failed: "构建失败", build_cancelled: "构建取消",
  deploy_success: "部署成功", deploy_failed: "部署失败", deploy_cancelled: "部署取消",
  approval_pending: "待审批", pod_unhealthy: "Pod 异常", rollback_triggered: "回滚触发",
};
const EVENT_TYPES = ["all", "build_success", "build_failed", "build_cancelled", "deploy_success", "deploy_failed", "deploy_cancelled", "approval_pending", "pod_unhealthy", "rollback_triggered"];
const LOG_STATUS: Record<string, { label: string; color: string }> = {
  sent: { label: "成功", color: "bg-green-100 text-green-700" },
  failed: { label: "失败", color: "bg-red-100 text-red-700" },
};

interface NotifRule { id: number; channel_id: number; event_type: string; enabled: boolean; channel?: NotificationChannel }
interface SvcRule { id: number; service_id: number; channel_id: number; enabled: boolean; channel?: NotificationChannel; service?: Service }
interface NotifLog { id: number; service_id: number; channel_id: number; event_type: string; title: string; content: string; status: string; error_msg?: string; created_at: string }

export default function NotificationsPage() {
  const queryClient = useQueryClient();
  const [mainTab, setMainTab] = useState<MainTab>("inbox");

  // ========== 站内通知 ==========
  const [readFilter, setReadFilter] = useState<ReadFilter>("all");
  const { data: notifsData, isLoading: notifsLoading } = useQuery({
    queryKey: ["notifications", readFilter],
    queryFn: async () => {
      const params: Record<string, string> = {};
      if (readFilter === "unread") params.is_read = "false";
      if (readFilter === "read") params.is_read = "true";
      return (await apiClient.get("/notifications", { params })).data;
    },
    enabled: mainTab === "inbox",
  });
  const notifications: Notification[] = notifsData?.items ?? [];
  const unreadCount = notifications.filter((n) => !n.is_read).length;
  const markRead = useMutation({ mutationFn: (id: number) => apiClient.put(`/notifications/${id}/read`), onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notifications"] }) });
  const markAllRead = useMutation({ mutationFn: () => apiClient.put("/notifications/read-all"), onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notifications"] }) });

  // ========== 渠道列表（规则页共用） ==========
  const { data: channelsData } = useQuery({
    queryKey: ["notification-channels"],
    queryFn: () => apiClient.get("/notification-channels").then((r) => r.data),
    enabled: mainTab === "rules",
  });
  const channels: NotificationChannel[] = channelsData?.items ?? [];

  // ========== 全局规则 ==========
  const { data: rulesData } = useQuery({
    queryKey: ["notification-rules"],
    queryFn: () => apiClient.get("/notification-rules").then((r) => r.data),
    enabled: mainTab === "rules",
  });
  const globalRules: NotifRule[] = rulesData?.items ?? [];
  const upsertGlobal = useMutation({ mutationFn: (d: { channel_id: number; event_type: string; enabled: boolean }) => apiClient.post("/notification-rules", d), onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notification-rules"] }) });
  const deleteGlobal = useMutation({ mutationFn: (id: number) => apiClient.delete(`/notification-rules/${id}`), onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notification-rules"] }) });
  const [newEvt, setNewEvt] = useState(""); const [newCh, setNewCh] = useState("");

  // ========== 服务级规则 ==========
  const { data: svcRulesData } = useQuery({
    queryKey: ["service-notification-rules"],
    queryFn: () => apiClient.get("/service-notification-rules").then((r) => r.data),
    enabled: mainTab === "rules",
  });
  const svcRules: SvcRule[] = svcRulesData?.items ?? [];
  const { data: servicesData } = useQuery({
    queryKey: ["services"],
    queryFn: () => apiClient.get("/services").then((r) => r.data),
    enabled: mainTab === "rules",
  });
  const services: Service[] = servicesData?.items ?? [];
  const upsertSvc = useMutation({ mutationFn: (d: { service_id: number; channel_id: number; enabled: boolean }) => apiClient.post("/service-notification-rules", d), onSuccess: () => queryClient.invalidateQueries({ queryKey: ["service-notification-rules"] }) });
  const deleteSvc = useMutation({ mutationFn: (id: number) => apiClient.delete(`/service-notification-rules/${id}`), onSuccess: () => queryClient.invalidateQueries({ queryKey: ["service-notification-rules"] }) });
  const [newSvcId, setNewSvcId] = useState(""); const [newSvcCh, setNewSvcCh] = useState("");

  // ========== 发送记录 ==========
  const [logPage, setLogPage] = useState(1);
  const { data: logsData } = useQuery({
    queryKey: ["notification-logs", logPage],
    queryFn: () => apiClient.get("/notification-logs", { params: { page: logPage, page_size: 20 } }).then((r) => r.data),
    enabled: mainTab === "logs",
  });
  const logs: NotifLog[] = logsData?.items ?? [];
  const logTotal = logsData?.total ?? 0;

  const mainTabs: { key: MainTab; label: string }[] = [
    { key: "inbox", label: "站内通知" },
    { key: "rules", label: "通知规则" },
    { key: "logs", label: "发送记录" },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">通知中心</h1>
        {mainTab === "inbox" && (
          <button onClick={() => markAllRead.mutate()} disabled={unreadCount === 0}
            className="rounded-md border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50">
            全部标为已读
          </button>
        )}
      </div>

      <div className="flex gap-1 border-b border-gray-200">
        {mainTabs.map((t) => (
          <button key={t.key} onClick={() => setMainTab(t.key)}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${mainTab === t.key ? "border-blue-600 text-blue-600" : "border-transparent text-gray-500 hover:text-gray-700"}`}>
            {t.label}
            {t.key === "inbox" && unreadCount > 0 && <span className="ml-1.5 rounded-full bg-red-500 px-1.5 py-0.5 text-xs text-white">{unreadCount}</span>}
          </button>
        ))}
      </div>

      {/* ========== 站内通知 ========== */}
      {mainTab === "inbox" && (
        <>
          <div className="flex gap-1">
            {(["all", "unread", "read"] as ReadFilter[]).map((f) => (
              <button key={f} onClick={() => setReadFilter(f)}
                className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${readFilter === f ? "bg-blue-100 text-blue-700" : "bg-gray-100 text-gray-600 hover:bg-gray-200"}`}>
                {{ all: "全部", unread: "未读", read: "已读" }[f]}
              </button>
            ))}
          </div>
          {notifsLoading ? <div className="py-12 text-center text-gray-500">加载中...</div> : notifications.length === 0 ? (
            <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center"><p className="text-gray-500">暂无通知</p></div>
          ) : (
            <div className="space-y-2">
              {notifications.map((n) => (
                <div key={n.id} onClick={() => { if (!n.is_read) markRead.mutate(n.id); }}
                  className={`flex cursor-pointer items-start gap-3 rounded-lg border p-4 transition-colors ${n.is_read ? "border-gray-200 bg-white hover:bg-gray-50" : "border-blue-200 bg-blue-50/50 hover:bg-blue-50"}`}>
                  <div className="mt-1 shrink-0">{!n.is_read && <span className="block h-2 w-2 rounded-full bg-blue-500" />}{n.is_read && <span className="block h-2 w-2" />}</div>
                  <div className="flex-1 min-w-0">
                    <p className={`text-sm ${n.is_read ? "text-gray-600" : "text-gray-900 font-medium"}`}>{n.title}</p>
                    {n.content && <p className="mt-1 text-sm text-gray-500 truncate">{n.content}</p>}
                    <p className="mt-1 text-xs text-gray-400">{new Date(n.created_at).toLocaleString("zh-CN")}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {/* ========== 通知规则 ========== */}
      {mainTab === "rules" && (
        <div className="space-y-6">
          {/* 全局规则 */}
          <div className="rounded-lg border border-gray-200 bg-white">
            <div className="border-b border-gray-200 px-4 py-3">
              <h3 className="text-sm font-semibold text-gray-900">全局通知规则</h3>
              <p className="text-xs text-gray-500 mt-0.5">优先级: 服务级绑定 &gt; 按事件自定义 &gt; All 默认渠道</p>
            </div>
            <div className="divide-y divide-gray-100">
              {EVENT_TYPES.map((evt) => {
                const matched = globalRules.filter((r) => r.event_type === evt);
                return (
                  <div key={evt} className="flex items-center justify-between px-4 py-3">
                    <div className="flex items-center gap-2">
                      <span className={`text-sm font-medium w-28 ${evt === "all" ? "text-blue-700" : "text-gray-900"}`}>{EVENT_LABELS[evt] || evt}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      {matched.map((r) => (
                        <span key={r.id} className="inline-flex items-center gap-1 rounded-full bg-blue-50 px-2 py-0.5 text-xs text-blue-700">
                          {r.channel?.name || `#${r.channel_id}`}
                          <button onClick={() => deleteGlobal.mutate(r.id)} className="text-blue-400 hover:text-red-500 ml-0.5">×</button>
                        </span>
                      ))}
                      {matched.length === 0 && <span className="text-xs text-gray-400">{evt === "all" ? "未设置默认" : "跟随 All"}</span>}
                    </div>
                  </div>
                );
              })}
            </div>
            <div className="border-t border-gray-200 px-4 py-3 flex items-center gap-2">
              <select value={newEvt} onChange={(e) => setNewEvt(e.target.value)} className="rounded border border-gray-300 px-2 py-1.5 text-sm">
                <option value="">事件类型</option>
                {EVENT_TYPES.map((e) => <option key={e} value={e}>{EVENT_LABELS[e] || e}</option>)}
              </select>
              <select value={newCh} onChange={(e) => setNewCh(e.target.value)} className="rounded border border-gray-300 px-2 py-1.5 text-sm">
                <option value="">选择渠道</option>
                {channels.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
              <button onClick={() => {
                if (!newEvt || !newCh) return;
                upsertGlobal.mutate({ channel_id: Number(newCh), event_type: newEvt, enabled: true }, {
                  onSuccess: () => { setNewEvt(""); setNewCh(""); showToast("规则已添加", "success"); },
                });
              }} disabled={!newEvt || !newCh} className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-40">添加</button>
            </div>
          </div>

          {/* 服务级规则 */}
          <div className="rounded-lg border border-gray-200 bg-white">
            <div className="border-b border-gray-200 px-4 py-3">
              <h3 className="text-sm font-semibold text-gray-900">服务级通知规则</h3>
              <p className="text-xs text-gray-500 mt-0.5">为服务绑定专用渠道，该服务所有事件都走此渠道（覆盖全局规则）</p>
            </div>
            {svcRules.length > 0 && (
              <table className="w-full">
                <thead><tr className="bg-gray-50 border-b border-gray-200">
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">服务</th>
                  <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">渠道</th>
                  <th className="px-4 py-2 text-right text-xs font-medium text-gray-500">操作</th>
                </tr></thead>
                <tbody className="divide-y divide-gray-100">
                  {svcRules.map((r) => (
                    <tr key={r.id} className="hover:bg-gray-50">
                      <td className="px-4 py-2.5 text-sm text-gray-900">{r.service?.name || `#${r.service_id}`}</td>
                      <td className="px-4 py-2.5 text-sm text-blue-700">{r.channel?.name || `#${r.channel_id}`}</td>
                      <td className="px-4 py-2.5 text-right">
                        <button onClick={() => deleteSvc.mutate(r.id)} className="text-xs text-red-500 hover:text-red-700">删除</button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
            <div className="border-t border-gray-200 px-4 py-3 flex items-center gap-2">
              <select value={newSvcId} onChange={(e) => setNewSvcId(e.target.value)} className="rounded border border-gray-300 px-2 py-1.5 text-sm">
                <option value="">选择服务</option>
                {services.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
              <select value={newSvcCh} onChange={(e) => setNewSvcCh(e.target.value)} className="rounded border border-gray-300 px-2 py-1.5 text-sm">
                <option value="">选择渠道</option>
                {channels.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
              </select>
              <button onClick={() => {
                if (!newSvcId || !newSvcCh) return;
                upsertSvc.mutate({ service_id: Number(newSvcId), channel_id: Number(newSvcCh), enabled: true }, {
                  onSuccess: () => { setNewSvcId(""); setNewSvcCh(""); showToast("服务规则已添加", "success"); },
                });
              }} disabled={!newSvcId || !newSvcCh} className="rounded bg-blue-600 px-3 py-1.5 text-sm text-white hover:bg-blue-700 disabled:opacity-40">添加</button>
            </div>
          </div>
        </div>
      )}

      {/* ========== 发送记录 ========== */}
      {mainTab === "logs" && (
        <div className="rounded-lg border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead><tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">时间</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">事件</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">标题</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">错误</th>
            </tr></thead>
            <tbody className="divide-y divide-gray-100">
              {logs.length === 0 ? (
                <tr><td colSpan={5} className="px-4 py-12 text-center text-sm text-gray-500">暂无发送记录</td></tr>
              ) : logs.map((l) => {
                const st = LOG_STATUS[l.status];
                return (
                  <tr key={l.id} className="hover:bg-gray-50">
                    <td className="px-4 py-3 text-xs text-gray-500 whitespace-nowrap">{new Date(l.created_at).toLocaleString("zh-CN")}</td>
                    <td className="px-4 py-3 text-xs"><span className="rounded bg-gray-100 px-1.5 py-0.5">{EVENT_LABELS[l.event_type] || l.event_type}</span></td>
                    <td className="px-4 py-3 text-sm text-gray-900 truncate max-w-[300px]" title={l.title}>{l.title}</td>
                    <td className="px-4 py-3"><span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${st?.color || "bg-gray-100"}`}>{st?.label || l.status}</span></td>
                    <td className="px-4 py-3 text-xs text-red-500 truncate max-w-[200px]" title={l.error_msg}>{l.error_msg || "-"}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
          {logTotal > 20 && (
            <div className="flex items-center justify-between border-t border-gray-200 px-4 py-2">
              <span className="text-xs text-gray-500">共 {logTotal} 条</span>
              <div className="flex gap-2">
                <button onClick={() => setLogPage((p) => Math.max(1, p - 1))} disabled={logPage <= 1} className="rounded border px-2 py-1 text-xs disabled:opacity-40">上一页</button>
                <button onClick={() => setLogPage((p) => p + 1)} disabled={logPage * 20 >= logTotal} className="rounded border px-2 py-1 text-xs disabled:opacity-40">下一页</button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
