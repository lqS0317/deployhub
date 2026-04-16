"use client";

import { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { WSClient } from "@/lib/ws-client";

interface PodInfo {
  name: string;
  status: string;
  ready: boolean;
  containers: string[];
  created_at: string;
}

interface Props {
  deploymentId: number;
}

export function PodLogPanel({ deploymentId }: Props) {
  const [selectedPod, setSelectedPod] = useState("");
  const [selectedContainer, setSelectedContainer] = useState("");
  const [lines, setLines] = useState<string[]>([]);
  const [error, setError] = useState("");
  const [connected, setConnected] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WSClient | null>(null);

  const { data: pods = [], isLoading } = useQuery({
    queryKey: ["deploy-pods", deploymentId],
    queryFn: async () => {
      const res = await apiClient.get<{ items: PodInfo[] }>(`/deployments/${deploymentId}/pods`);
      return res.data.items ?? [];
    },
  });

  // 选择 Pod 后自动选第一个容器
  const currentPod = pods.find((p) => p.name === selectedPod);
  const containers = currentPod?.containers ?? [];

  useEffect(() => {
    if (pods.length > 0 && !selectedPod) {
      setSelectedPod(pods[0].name);
      if (pods[0].containers?.length > 0) {
        setSelectedContainer(pods[0].containers[0]);
      }
    }
  }, [pods, selectedPod]);

  useEffect(() => {
    if (currentPod && containers.length > 0 && !containers.includes(selectedContainer)) {
      setSelectedContainer(containers[0]);
    }
  }, [currentPod, containers, selectedContainer]);

  // 连接 WS
  const connectWS = () => {
    if (wsRef.current) {
      wsRef.current.disconnect();
    }
    setLines([]);
    setError("");

    if (!selectedPod || !selectedContainer) return;

    const token = localStorage.getItem("access_token") || "";
    const wsUrl = `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.hostname}:8080/ws/deployments/${deploymentId}/pod-logs?token=${token}&pod=${selectedPod}&container=${selectedContainer}&tail=200`;

    const ws = new WebSocket(wsUrl);
    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);
    ws.onerror = () => setError("WebSocket 连接失败");
    ws.onmessage = (e) => {
      const text = e.data;
      if (text.startsWith("[ERROR]")) {
        setError(text);
      } else {
        const newLines = text.split("\n").filter((l: string) => l);
        setLines((prev) => [...prev, ...newLines]);
      }
    };

    wsRef.current = { disconnect: () => ws.close() } as WSClient;
  };

  useEffect(() => {
    return () => {
      if (wsRef.current) wsRef.current.disconnect();
    };
  }, []);

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [lines]);

  return (
    <div className="rounded-lg border border-gray-200 bg-gray-900 overflow-hidden">
      <div className="flex items-center gap-3 border-b border-gray-700 px-4 py-2">
        <h4 className="text-sm font-medium text-gray-100">容器日志</h4>

        {isLoading ? (
          <span className="text-xs text-gray-400">加载 Pod 列表...</span>
        ) : pods.length === 0 ? (
          <span className="text-xs text-yellow-400">无 Pod</span>
        ) : (
          <>
            <select value={selectedPod} onChange={(e) => setSelectedPod(e.target.value)}
              className="rounded border border-gray-600 bg-gray-800 px-2 py-1 text-xs text-gray-200">
              {pods.map((p) => (
                <option key={p.name} value={p.name}>
                  {p.name} ({p.status}{p.ready ? "" : " ⚠"})
                </option>
              ))}
            </select>
            <select value={selectedContainer} onChange={(e) => setSelectedContainer(e.target.value)}
              className="rounded border border-gray-600 bg-gray-800 px-2 py-1 text-xs text-gray-200">
              {containers.map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
            <button onClick={connectWS}
              className="rounded bg-blue-600 px-2 py-1 text-xs text-white hover:bg-blue-700">
              {connected ? "重连" : "查看日志"}
            </button>
          </>
        )}

        {connected && (
          <span className="flex items-center gap-1 text-xs text-green-400">
            <span className="h-1.5 w-1.5 rounded-full bg-green-400" />实时
          </span>
        )}
      </div>

      {error && (
        <div className="border-b border-red-800 bg-red-900/50 px-4 py-2 text-xs text-red-300">
          {error}
        </div>
      )}

      <div ref={containerRef} className="h-80 overflow-y-auto p-3 font-mono text-xs leading-relaxed">
        {lines.length === 0 && !error ? (
          <p className="text-gray-500">选择 Pod 和容器后点击「查看日志」</p>
        ) : (
          lines.map((line, idx) => (
            <div key={idx} className="whitespace-pre-wrap break-all text-gray-300 py-px hover:bg-gray-800/50">
              {line}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
