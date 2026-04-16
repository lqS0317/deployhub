"use client";

import { useEffect, useRef, useState } from "react";
import { WSClient } from "@/lib/ws-client";

interface Props {
  deploymentId: number;
}

export function DeployLogPanel({ deploymentId }: Props) {
  const [lines, setLines] = useState<string[]>([]);
  const [connected, setConnected] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const autoScrollRef = useRef(true);

  useEffect(() => {
    const ws = new WSClient({
      url: `/ws/deployments/${deploymentId}/progress`,
      onMessage: (data: unknown) => {
        const msg = typeof data === "string" ? data : JSON.stringify(data);
        setLines((prev) => [...prev, msg]);
      },
      onOpen: () => setConnected(true),
      onClose: () => setConnected(false),
    });
    ws.connect();
    return () => ws.disconnect();
  }, [deploymentId]);

  useEffect(() => {
    if (autoScrollRef.current && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [lines]);

  return (
    <div className="rounded-lg border border-gray-200 bg-gray-900 overflow-hidden">
      <div className="flex items-center justify-between border-b border-gray-700 px-4 py-2">
        <h4 className="text-sm font-medium text-gray-100">部署事件日志</h4>
        <span className={`flex items-center gap-1.5 text-xs ${connected ? "text-green-400" : "text-gray-500"}`}>
          <span className={`h-1.5 w-1.5 rounded-full ${connected ? "bg-green-400" : "bg-gray-500"}`} />
          {connected ? "已连接" : "未连接"}
        </span>
      </div>
      <div
        ref={containerRef}
        onScroll={() => {
          if (!containerRef.current) return;
          const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
          autoScrollRef.current = scrollHeight - scrollTop - clientHeight < 50;
        }}
        className="h-64 overflow-y-auto p-3 font-mono text-xs leading-relaxed"
      >
        {lines.length === 0 ? (
          <p className="text-gray-500">等待事件...</p>
        ) : (
          lines.map((line, idx) => (
            <div key={idx} className="whitespace-pre-wrap break-all text-gray-300 py-px">
              <span className="mr-2 text-gray-600 select-none">{idx + 1}</span>
              {line}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
