"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

// 系统设置页面：默认重定向到集群管理
export default function SettingsPage() {
  const router = useRouter();

  useEffect(() => {
    router.replace("/settings/clusters");
  }, [router]);

  return null;
}
