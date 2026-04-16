"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import type { NotificationChannel } from "@/types";

interface ChannelForm {
  name: string;
  type: string;
  webhook_url: string;
}

const emptyForm: ChannelForm = { name: "", type: "dingtalk", webhook_url: "" };

export default function NotificationChannelsPage() {
  const queryClient = useQueryClient();
  const [showDialog, setShowDialog] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<ChannelForm>(emptyForm);
  const [testingId, setTestingId] = useState<number | null>(null);
  const [testResult, setTestResult] = useState<Record<number, { ok: boolean; msg: string }>>({});

  const { data: channelsData, isLoading } = useQuery({
    queryKey: ["notification-channels"],
    queryFn: () => apiClient.get("/notification-channels").then((r) => r.data),
  });
  const channels: NotificationChannel[] = channelsData?.items ?? [];

  const createMutation = useMutation({
    mutationFn: (data: ChannelForm) => apiClient.post("/notification-channels", data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ["notification-channels"] }); closeDialog(); },
  });
  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: ChannelForm }) => apiClient.put(`/notification-channels/${id}`, data),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ["notification-channels"] }); closeDialog(); },
  });
  const deleteMutation = useMutation({
    mutationFn: (id: number) => apiClient.delete(`/notification-channels/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notification-channels"] }),
  });

  const testWebhook = async (id: number) => {
    setTestingId(id);
    try {
      const res = await apiClient.post(`/notification-channels/${id}/test`);
      setTestResult((prev) => ({ ...prev, [id]: { ok: true, msg: res.data?.message || "成功" } }));
    } catch {
      setTestResult((prev) => ({ ...prev, [id]: { ok: false, msg: "失败" } }));
    } finally {
      setTestingId(null);
    }
  };

  const openCreate = () => { setEditingId(null); setForm(emptyForm); setShowDialog(true); };
  const openEdit = (ch: NotificationChannel) => { setEditingId(ch.id); setForm({ name: ch.name, type: ch.type, webhook_url: ch.webhook_url }); setShowDialog(true); };
  const closeDialog = () => { setShowDialog(false); setEditingId(null); setForm(emptyForm); };
  const handleSubmit = () => { editingId ? updateMutation.mutate({ id: editingId, data: form }) : createMutation.mutate(form); };

  const typeBadge = (t: string) => {
    const styles: Record<string, string> = { dingtalk: "bg-blue-100 text-blue-800", feishu: "bg-purple-100 text-purple-800", slack: "bg-yellow-100 text-yellow-800" };
    const labels: Record<string, string> = { dingtalk: "钉钉", feishu: "飞书", wechat: "企业微信", slack: "Slack", webhook: "Webhook" };
    return <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${styles[t] || "bg-gray-100 text-gray-800"}`}>{labels[t] || t}</span>;
  };

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <button onClick={openCreate} className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">添加通知渠道</button>
      </div>

      {isLoading ? <div className="py-12 text-center text-gray-500">加载中...</div> : channels.length === 0 ? (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center"><p className="text-gray-500">暂无通知渠道</p></div>
      ) : (
        <div className="rounded-lg border border-gray-200 overflow-hidden">
          <table className="w-full">
            <thead><tr className="bg-gray-50 border-b border-gray-200">
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">名称</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">类型</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">Webhook URL</th>
              <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr></thead>
            <tbody className="divide-y divide-gray-100">
              {channels.map((ch) => (
                <tr key={ch.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-sm font-medium text-gray-900">{ch.name}</td>
                  <td className="px-4 py-3">{typeBadge(ch.type)}</td>
                  <td className="px-4 py-3 text-sm text-gray-600 font-mono truncate max-w-[250px]">{ch.webhook_url}</td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button onClick={() => testWebhook(ch.id)} disabled={testingId === ch.id} className="text-xs text-blue-600 hover:text-blue-800 disabled:opacity-50">{testingId === ch.id ? "发送中..." : "测试"}</button>
                      {testResult[ch.id] && <span className={`text-xs ${testResult[ch.id].ok ? "text-green-600" : "text-red-600"}`}>{testResult[ch.id].msg}</span>}
                      <button onClick={() => openEdit(ch)} className="text-xs text-gray-600 hover:text-gray-800">编辑</button>
                      <button onClick={() => { if (confirm("确定删除？")) deleteMutation.mutate(ch.id); }} className="text-xs text-red-600 hover:text-red-800">删除</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {showDialog && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="absolute inset-0 bg-black/50" onClick={closeDialog} />
          <div className="relative w-full max-w-lg rounded-lg bg-white shadow-xl">
            <div className="flex items-center justify-between border-b border-gray-200 px-6 py-4">
              <h3 className="text-lg font-semibold text-gray-900">{editingId ? "编辑通知渠道" : "添加通知渠道"}</h3>
              <button onClick={closeDialog} className="text-gray-400 hover:text-gray-600">✕</button>
            </div>
            <div className="px-6 py-4 space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">渠道名称</label>
                <input type="text" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="例如: 运维告警群"
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">渠道类型</label>
                <select value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })}
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500">
                  <option value="dingtalk">钉钉</option><option value="feishu">飞书</option><option value="wechat">企业微信</option><option value="slack">Slack</option><option value="webhook">自定义 Webhook</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Webhook URL</label>
                <input type="text" value={form.webhook_url} onChange={(e) => setForm({ ...form, webhook_url: e.target.value })} placeholder="https://..."
                  className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm font-mono focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
              </div>
            </div>
            <div className="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
              <button onClick={closeDialog} className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50">取消</button>
              <button onClick={handleSubmit} disabled={!form.name || !form.webhook_url || createMutation.isPending || updateMutation.isPending}
                className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50">
                {createMutation.isPending || updateMutation.isPending ? "保存中..." : editingId ? "保存" : "添加"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
