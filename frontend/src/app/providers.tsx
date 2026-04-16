"use client";

import { QueryClient, QueryClientProvider, MutationCache } from "@tanstack/react-query";
import { useState } from "react";
import { ToastProvider, showToast } from "@/components/ui/toast";

function extractErrorMessage(error: unknown): string {
  if (error && typeof error === "object" && "response" in error) {
    const resp = (error as { response?: { data?: { error?: { message?: string } }; status?: number } }).response;
    if (resp?.data?.error?.message) return resp.data.error.message;
    if (resp?.status === 403) return "权限不足，需要管理员角色";
    if (resp?.status === 409) return "资源已存在";
    if (resp?.status === 400) return "请求参数错误";
  }
  if (error instanceof Error) return error.message;
  return "操作失败";
}

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000,
            retry: 1,
          },
        },
        mutationCache: new MutationCache({
          onError: (error) => {
            showToast(extractErrorMessage(error), "error");
          },
        }),
      })
  );

  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>{children}</ToastProvider>
    </QueryClientProvider>
  );
}
