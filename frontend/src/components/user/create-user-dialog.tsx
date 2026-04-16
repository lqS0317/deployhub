"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { showToast } from "@/components/ui/toast";

interface Props {
  onClose: () => void;
}

export function CreateUserDialog({ onClose }: Props) {
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("member");
  const qc = useQueryClient();

  const createUser = useMutation({
    mutationFn: async () => {
      const res = await apiClient.post("/users", { username, email, password, role });
      return res.data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["users"] });
      showToast("用户创建成功", "success");
      onClose();
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim() || !email.trim() || !password) return;
    createUser.mutate();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative z-10 w-full max-w-md rounded-xl bg-white p-6 shadow-2xl">
        <h2 className="text-lg font-semibold mb-4">创建用户</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">用户名 *</label>
            <input type="text" value={username} onChange={(e) => setUsername(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">邮箱 *</label>
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">密码 *</label>
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="至少 6 位"
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">角色</label>
            <select value={role} onChange={(e) => setRole(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500">
              <option value="member">成员</option>
              <option value="admin">管理员</option>
            </select>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className="rounded-lg border border-gray-300 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">取消</button>
            <button type="submit" disabled={createUser.isPending} className="rounded-lg bg-blue-600 px-4 py-2 text-sm text-white hover:bg-blue-700 disabled:opacity-50">
              {createUser.isPending ? "创建中..." : "创建用户"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
