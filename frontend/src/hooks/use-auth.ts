"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import apiClient from "@/lib/api-client";
import type { User, LoginResponse, RegisterResponse } from "@/types";

// queryKey 常量
const AUTH_KEYS = {
  me: ["auth", "me"] as const,
};

/** 用户登录 */
export function useLogin() {
  const queryClient = useQueryClient();
  const router = useRouter();

  return useMutation({
    mutationFn: async (data: { username: string; password: string }) => {
      const res = await apiClient.post<LoginResponse>("/auth/login", data);
      return res.data;
    },
    onSuccess: (data) => {
      localStorage.setItem("access_token", data.access_token);
      queryClient.invalidateQueries({ queryKey: AUTH_KEYS.me });
      router.push("/services");
    },
  });
}

/** 用户注册 */
export function useRegister() {
  return useMutation({
    mutationFn: async (data: {
      username: string;
      email: string;
      password: string;
    }) => {
      const res = await apiClient.post<RegisterResponse>(
        "/auth/register",
        data
      );
      return res.data;
    },
  });
}

/** 获取当前用户信息 */
export function useMe() {
  return useQuery({
    queryKey: AUTH_KEYS.me,
    queryFn: async () => {
      const res = await apiClient.get<User>("/auth/me");
      return res.data;
    },
    retry: false,
    enabled:
      typeof window !== "undefined" &&
      !!localStorage.getItem("access_token"),
  });
}

/** 登出：清除 token 并调用后端 + 跳转 */
export function useLogout() {
  const queryClient = useQueryClient();
  const router = useRouter();

  return useMutation({
    mutationFn: async () => {
      try {
        await apiClient.post("/auth/logout");
      } catch {
        // 忽略后端错误，确保前端 token 清除
      }
      localStorage.removeItem("access_token");
    },
    onSuccess: () => {
      queryClient.clear();
      router.push("/login");
    },
  });
}

/** 更新用户资料（昵称、手机号） */
export function useUpdateProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { nickname?: string; phone?: string }) => {
      const res = await apiClient.put("/auth/profile", data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: AUTH_KEYS.me });
    },
  });
}

/** 修改密码 */
export function useChangePassword() {
  return useMutation({
    mutationFn: async (data: {
      old_password: string;
      new_password: string;
    }) => {
      const res = await apiClient.put("/auth/password", data);
      return res.data;
    },
  });
}

/** 上传头像 */
export function useUploadAvatar() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData();
      formData.append("avatar", file);
      const res = await apiClient.post("/auth/avatar", formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: AUTH_KEYS.me });
    },
  });
}
