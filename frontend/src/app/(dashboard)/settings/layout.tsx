"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

// 系统设置 Tab 导航栏配置
const settingsTabs = [
  { href: "/settings/clusters", label: "集群管理" },
  { href: "/settings/git-repos", label: "Git 仓库" },
  { href: "/settings/registries", label: "镜像仓库" },
  { href: "/settings/users", label: "用户管理" },
  { href: "/settings/groups", label: "组管理" },
  { href: "/settings/notifications", label: "通知渠道" },
  { href: "/settings/system", label: "系统配置" },
];

// 系统设置布局：顶部 Tab 导航 + 子页面内容
export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const pathname = usePathname();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">系统设置</h1>
        <p className="mt-1 text-sm text-gray-500">管理平台基础设施与账号</p>
      </div>

      {/* Tab 导航 */}
      <div className="flex gap-1 border-b border-gray-200">
        {settingsTabs.map((tab) => {
          const isActive = pathname === tab.href;
          return (
            <Link
              key={tab.href}
              href={tab.href}
              className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
                isActive
                  ? "border-blue-600 text-blue-600"
                  : "border-transparent text-gray-500 hover:text-gray-700"
              }`}
            >
              {tab.label}
            </Link>
          );
        })}
      </div>

      {children}
    </div>
  );
}
