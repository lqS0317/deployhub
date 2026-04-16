"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { WSClient } from "@/lib/ws-client";
import apiClient from "@/lib/api-client";

interface BuildLogViewerProps {
  buildId: string | number;
  onClose: () => void;
}

type ConnectionStatus = "connecting" | "connected" | "disconnected" | "error" | "loaded";

const STATUS_INDICATOR: Record<ConnectionStatus, { color: string; label: string }> = {
  connecting: { color: "bg-yellow-400", label: "连接中..." },
  connected: { color: "bg-green-400", label: "实时日志" },
  disconnected: { color: "bg-gray-400", label: "已断开" },
  error: { color: "bg-red-400", label: "连接错误" },
  loaded: { color: "bg-blue-400", label: "历史日志" },
};

/**
 * 清洗日志文本：
 * - 处理 \r 回车覆盖（Git/Docker 进度条），只保留每行最后一次覆盖结果
 * - 按 \n 拆行
 * - 去掉空行
 */
function cleanLogText(raw: string): string[] {
  const lines = raw.split("\n");
  const result: string[] = [];

  for (const line of lines) {
    if (line.includes("\r")) {
      const parts = line.split("\r");
      const last = parts[parts.length - 1].trim();
      if (last) result.push(last);
    } else {
      const trimmed = line.trimEnd();
      if (trimmed) result.push(trimmed);
    }
  }

  return result;
}

export function BuildLogViewer({ buildId, onClose }: BuildLogViewerProps) {
  const [lines, setLines] = useState<string[]>([]);
  const [status, setStatus] = useState<ConnectionStatus>("connecting");
  const containerRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WSClient | null>(null);
  const autoScrollRef = useRef(true);

  const scrollToBottom = useCallback(() => {
    if (autoScrollRef.current && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, []);

  const handleScroll = () => {
    if (!containerRef.current) return;
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
    autoScrollRef.current = scrollHeight - scrollTop - clientHeight < 50;
  };

  useEffect(() => {
    let cancelled = false;

    apiClient.get(`/builds/${buildId}/log`).then((res) => {
      if (cancelled) return;
      const logText = res.data?.data?.log || res.data?.log || "";
      if (logText) {
        setLines(cleanLogText(logText));
        setStatus("loaded");
      }
    }).catch(() => {});

    const ws = new WSClient({
      url: `/ws/builds/${buildId}/log`,
      onMessage: (data: unknown) => {
        const msg = data as { type?: string; chunk?: string; line?: string };
        const text = msg.chunk || msg.line || (typeof data === "string" ? data : "");
        if (text) {
          const newLines = cleanLogText(text);
          if (newLines.length > 0) {
            setLines((prev) => [...prev, ...newLines]);
          }
        }
      },
      onOpen: () => setStatus("connected"),
      onClose: () => {},
      onError: () => {},
    });

    wsRef.current = ws;
    ws.connect();

    return () => {
      cancelled = true;
      ws.disconnect();
      wsRef.current = null;
    };
  }, [buildId]);

  useEffect(() => {
    scrollToBottom();
  }, [lines, scrollToBottom]);

  const indicator = STATUS_INDICATOR[status];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60" onClick={onClose} />

      <div className="relative z-10 flex h-[80vh] w-full max-w-4xl flex-col rounded-xl bg-gray-900 shadow-2xl">
        <div className="flex items-center justify-between border-b border-gray-700 px-4 py-3">
          <div className="flex items-center gap-3">
            <h3 className="text-sm font-semibold text-gray-100">构建日志</h3>
            <span className="font-mono text-xs text-gray-400">#{String(buildId)}</span>
          </div>
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <span className={`h-2 w-2 rounded-full ${indicator.color}`} />
              <span className="text-xs text-gray-400">{indicator.label}</span>
            </div>
            <button
              onClick={onClose}
              className="rounded p-1 text-gray-400 transition-colors hover:bg-gray-700 hover:text-gray-200"
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <div
          ref={containerRef}
          onScroll={handleScroll}
          className="flex-1 overflow-y-auto p-4 font-mono text-xs leading-relaxed"
        >
          {lines.length === 0 ? (
            <div className="flex h-full items-center justify-center">
              <p className="text-gray-500">等待日志输出...</p>
            </div>
          ) : (
            lines.map((line, idx) => (
              <div
                key={idx}
                className={`whitespace-pre-wrap break-all py-px hover:bg-gray-800/50 ${
                  line.includes("[DeployHub] 构建失败") || line.includes("error")
                    ? "text-red-400"
                    : line.includes("[DeployHub] 构建成功")
                    ? "text-green-400"
                    : line.startsWith("[DeployHub]")
                    ? "text-blue-400"
                    : "text-gray-300"
                }`}
              >
                <span className="mr-3 inline-block w-10 select-none text-right text-gray-600">
                  {idx + 1}
                </span>
                {line}
              </div>
            ))
          )}
        </div>

        <div className="flex items-center justify-between border-t border-gray-700 px-4 py-2">
          <span className="text-xs text-gray-500">{lines.length} 行</span>
          <button
            onClick={scrollToBottom}
            className="text-xs text-blue-400 transition-colors hover:text-blue-300"
          >
            滚动到底部
          </button>
        </div>
      </div>
    </div>
  );
}
