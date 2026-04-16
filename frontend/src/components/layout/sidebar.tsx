"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, useRef, useEffect } from "react";
import { useMe, useLogout } from "@/hooks/use-auth";

const navItems = [
  { href: "/services", label: "服务管理", icon: "□" },
  { href: "/builds", label: "构建中心", icon: "⚙" },
  { href: "/deployments", label: "发布管理", icon: "▶" },
  { href: "/configs", label: "配置中心", icon: "☰" },
  { href: "/routes", label: "路由中心", icon: "⇌" },
  { href: "/plugins", label: "插件中心", icon: "⊞" },
  { href: "/approvals", label: "审批中心", icon: "✓" },
  { href: "/notifications", label: "通知中心", icon: "🔔" },
  { href: "/settings", label: "系统设置", icon: "⚡" },
];

export function Sidebar() {
  const pathname = usePathname();
  const { data: user } = useMe();
  const logout = useLogout();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const displayName = user?.nickname || user?.username || "用户";
  const avatarText = displayName.charAt(0).toUpperCase();

  return (
    <aside className="flex h-screen w-60 flex-col border-r border-gray-200 bg-white">
      <div className="flex h-14 items-center border-b border-gray-200 px-4">
        <h1 className="text-lg font-bold text-gray-900">DeployHub</h1>
      </div>
      <nav className="flex-1 overflow-y-auto p-3">
        <ul className="space-y-1">
          {navItems.map((item) => {
            const isActive = pathname.startsWith(item.href);
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  className={`flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                    isActive
                      ? "bg-blue-50 text-blue-700"
                      : "text-gray-600 hover:bg-gray-100 hover:text-gray-900"
                  }`}
                >
                  <span className="text-base">{item.icon}</span>
                  {item.label}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      {/* 用户信息区域 */}
      <div className="relative border-t border-gray-200 p-3" ref={dropdownRef}>
        <button
          onClick={() => setDropdownOpen(!dropdownOpen)}
          className="flex w-full items-center gap-3 rounded-md px-2 py-2 text-sm hover:bg-gray-100 transition-colors"
        >
          {user?.avatar ? (
            <img
              src={user.avatar}
              alt="头像"
              className="h-8 w-8 rounded-full object-cover"
            />
          ) : (
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-white text-sm font-medium">
              {avatarText}
            </div>
          )}
          <div className="flex-1 text-left">
            <div className="font-medium text-gray-900 truncate">{displayName}</div>
            <div className="text-xs text-gray-500 truncate">{user?.email}</div>
          </div>
          <svg className="h-4 w-4 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clipRule="evenodd" />
          </svg>
        </button>

        {dropdownOpen && (
          <div className="absolute bottom-full left-3 right-3 mb-1 rounded-md border border-gray-200 bg-white py-1 shadow-lg z-50">
            <Link
              href="/profile"
              onClick={() => setDropdownOpen(false)}
              className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
            >
              个人资料
            </Link>
            <button
              onClick={() => {
                setDropdownOpen(false);
                logout.mutate();
              }}
              className="block w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-gray-100"
            >
              退出登录
            </button>
          </div>
        )}
      </div>
    </aside>
  );
}
