"use client";

import { createContext, useContext, useState, useCallback, useRef } from "react";

interface ToastItem {
  id: number;
  message: string;
  type: "success" | "error" | "warning";
}

interface ToastContextType {
  toast: (message: string, type?: ToastItem["type"]) => void;
}

const ToastContext = createContext<ToastContextType>({ toast: () => {} });

export function useToast() {
  return useContext(ToastContext);
}

// 全局引用，供非组件代码（如 axios 拦截器）调用
let globalToast: ToastContextType["toast"] | null = null;

export function showToast(message: string, type: ToastItem["type"] = "error") {
  if (globalToast) {
    globalToast(message, type);
  }
}

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<ToastItem[]>([]);
  const counter = useRef(0);

  const toast = useCallback((message: string, type: ToastItem["type"] = "error") => {
    const id = ++counter.current;
    setItems((prev) => [...prev, { id, message, type }]);
    setTimeout(() => {
      setItems((prev) => prev.filter((t) => t.id !== id));
    }, 4000);
  }, []);

  globalToast = toast;

  const remove = (id: number) => {
    setItems((prev) => prev.filter((t) => t.id !== id));
  };

  const colorMap = {
    success: "bg-green-600",
    error: "bg-red-600",
    warning: "bg-yellow-500 text-black",
  };

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <div className="fixed top-4 right-4 z-[9999] flex flex-col gap-2 pointer-events-none">
        {items.map((item) => (
          <div
            key={item.id}
            className={`pointer-events-auto flex items-center gap-2 rounded-lg px-4 py-3 text-sm font-medium text-white shadow-lg animate-in slide-in-from-right ${colorMap[item.type]}`}
            style={{ animation: "slideIn 0.2s ease-out" }}
          >
            <span className="flex-1">{item.message}</span>
            <button onClick={() => remove(item.id)} className="opacity-70 hover:opacity-100 text-lg leading-none">
              ×
            </button>
          </div>
        ))}
      </div>
      <style jsx global>{`
        @keyframes slideIn {
          from { transform: translateX(100%); opacity: 0; }
          to { transform: translateX(0); opacity: 1; }
        }
      `}</style>
    </ToastContext.Provider>
  );
}
