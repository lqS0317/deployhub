"use client";

import { useState, useMemo } from "react";
import type { ConfigVersion } from "@/types";

interface VersionDiffProps {
  versions: ConfigVersion[];
}

// 配置版本对比组件，左右分栏显示差异
export function VersionDiff({ versions }: VersionDiffProps) {
  const [leftVersion, setLeftVersion] = useState<string>(
    versions[1]?.id?.toString() || ""
  );
  const [rightVersion, setRightVersion] = useState<string>(
    versions[0]?.id?.toString() || ""
  );

  const leftContent =
    versions.find((v) => v.id?.toString() === leftVersion)?.rendered_content || "";
  const rightContent =
    versions.find((v) => v.id?.toString() === rightVersion)?.rendered_content || "";

  // 简单逐行差异对比
  const diffLines = useMemo(() => {
    const leftLines = leftContent.split("\n");
    const rightLines = rightContent.split("\n");
    const maxLen = Math.max(leftLines.length, rightLines.length);

    const result: Array<{
      left: string;
      right: string;
      type: "same" | "added" | "removed" | "changed";
    }> = [];

    for (let i = 0; i < maxLen; i++) {
      const l = leftLines[i] ?? "";
      const r = rightLines[i] ?? "";
      if (l === r) {
        result.push({ left: l, right: r, type: "same" });
      } else if (!l && r) {
        result.push({ left: "", right: r, type: "added" });
      } else if (l && !r) {
        result.push({ left: l, right: "", type: "removed" });
      } else {
        result.push({ left: l, right: r, type: "changed" });
      }
    }
    return result;
  }, [leftContent, rightContent]);

  const lineStyle = (type: string, side: "left" | "right") => {
    if (type === "same") return "";
    if (type === "added" && side === "right") return "bg-green-100 text-green-800";
    if (type === "removed" && side === "left") return "bg-red-100 text-red-800";
    if (type === "changed") {
      return side === "left"
        ? "bg-red-50 text-red-700"
        : "bg-green-50 text-green-700";
    }
    return "";
  };

  if (versions.length < 2) {
    return (
      <div className="text-sm text-gray-500 py-4 text-center">
        至少需要两个版本才能进行对比
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      {/* 版本选择器 */}
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2">
          <label className="text-sm text-gray-600">旧版本</label>
          <select
            value={leftVersion}
            onChange={(e) => setLeftVersion(e.target.value)}
            className="rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-blue-500 focus:outline-none"
          >
            {versions.map((v) => (
              <option key={v.id} value={v.id}>
                v{v.version} - {new Date(v.created_at).toLocaleString("zh-CN")}
              </option>
            ))}
          </select>
        </div>
        <span className="text-gray-400">→</span>
        <div className="flex items-center gap-2">
          <label className="text-sm text-gray-600">新版本</label>
          <select
            value={rightVersion}
            onChange={(e) => setRightVersion(e.target.value)}
            className="rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-blue-500 focus:outline-none"
          >
            {versions.map((v) => (
              <option key={v.id} value={v.id}>
                v{v.version} - {new Date(v.created_at).toLocaleString("zh-CN")}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* 差异对比区域 */}
      <div className="grid grid-cols-2 gap-0 rounded-lg border border-gray-300 overflow-hidden">
        {/* 左栏标题 */}
        <div className="bg-red-50 px-3 py-2 border-b border-r border-gray-300">
          <span className="text-xs font-medium text-red-700">旧版本</span>
        </div>
        <div className="bg-green-50 px-3 py-2 border-b border-gray-300">
          <span className="text-xs font-medium text-green-700">新版本</span>
        </div>

        {/* 左栏内容 */}
        <div className="border-r border-gray-300 font-mono text-xs overflow-x-auto">
          {diffLines.map((line, i) => (
            <div
              key={i}
              className={`flex px-2 py-0.5 min-h-[20px] ${lineStyle(line.type, "left")}`}
            >
              <span className="w-8 shrink-0 text-right text-gray-400 pr-2 select-none">
                {line.left !== undefined ? i + 1 : ""}
              </span>
              <span className="whitespace-pre">{line.left}</span>
            </div>
          ))}
        </div>

        {/* 右栏内容 */}
        <div className="font-mono text-xs overflow-x-auto">
          {diffLines.map((line, i) => (
            <div
              key={i}
              className={`flex px-2 py-0.5 min-h-[20px] ${lineStyle(line.type, "right")}`}
            >
              <span className="w-8 shrink-0 text-right text-gray-400 pr-2 select-none">
                {line.right !== undefined ? i + 1 : ""}
              </span>
              <span className="whitespace-pre">{line.right}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
