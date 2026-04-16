const WS_BASE_URL = process.env.NEXT_PUBLIC_WS_URL || "";

function resolveWsBase(): string {
  if (WS_BASE_URL) return WS_BASE_URL;
  if (typeof window === "undefined") return "";
  const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${window.location.host}`;
}

// 心跳间隔（毫秒）
const HEARTBEAT_INTERVAL = 30000;
// 重连延迟基数（毫秒）
const RECONNECT_BASE_DELAY = 1000;
// 最大重连延迟（毫秒）
const MAX_RECONNECT_DELAY = 30000;

export type WSMessageHandler = (data: unknown) => void;

export interface WSClientOptions {
  url: string;
  onMessage: WSMessageHandler;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: (error: Event) => void;
}

// WebSocket 客户端，支持自动重连和心跳
export class WSClient {
  private ws: WebSocket | null = null;
  private options: WSClientOptions;
  private reconnectAttempts = 0;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private isClosed = false;

  constructor(options: WSClientOptions) {
    this.options = options;
  }

  connect(): void {
    this.isClosed = false;
    const token = typeof window !== "undefined" ? localStorage.getItem("access_token") : null;
    const url = `${resolveWsBase()}${this.options.url}${token ? `?token=${token}` : ""}`;

    this.ws = new WebSocket(url);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.startHeartbeat();
      this.options.onOpen?.();
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.options.onMessage(data);
      } catch {
        this.options.onMessage(event.data);
      }
    };

    this.ws.onclose = () => {
      this.stopHeartbeat();
      this.options.onClose?.();
      if (!this.isClosed) {
        this.reconnect();
      }
    };

    this.ws.onerror = (error) => {
      this.options.onError?.(error);
    };
  }

  // 断开连接
  disconnect(): void {
    this.isClosed = true;
    this.stopHeartbeat();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: "ping" }));
      }
    }, HEARTBEAT_INTERVAL);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  // 指数退避重连
  private reconnect(): void {
    const delay = Math.min(
      RECONNECT_BASE_DELAY * Math.pow(2, this.reconnectAttempts),
      MAX_RECONNECT_DELAY
    );
    this.reconnectAttempts++;
    this.reconnectTimer = setTimeout(() => {
      this.connect();
    }, delay);
  }
}
