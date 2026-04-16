"use client";

import { useState, useEffect } from "react";
import {
  useCreateRouteEntry,
  useUpdateRouteEntry,
} from "@/hooks/use-route-entries";
import { showToast } from "@/components/ui/toast";
import { ServiceForm } from "./service-form";
import { IngressForm } from "./ingress-form";
import { IngressRouteForm } from "./ingressroute-form";
import { ApisixRouteForm } from "./apisixroute-form";
import type { RouteEntry } from "@/types";

interface CreateRouteDialogProps {
  open: boolean;
  onClose: () => void;
  resourceType: string;
  editEntry?: RouteEntry | null;
}

const EMPTY_CONFIGS: Record<string, unknown> = {
  service: { type: "ClusterIP", selector: {}, ports: [] },
  ingress: { ingressClassName: "", annotations: {}, tls: [], rules: [] },
  ingressroute: {
    entryPoints: ["websecure"],
    routes: [],
    tls: {},
  },
  apisixroute: { rules: [] },
};

const TYPE_LABELS: Record<string, string> = {
  service: "Service",
  ingress: "Ingress",
  ingressroute: "IngressRoute",
  apisixroute: "ApisixRoute",
};

export function CreateRouteDialog({
  open,
  onClose,
  resourceType,
  editEntry,
}: CreateRouteDialogProps) {
  const isEdit = !!editEntry;
  const initialConfig = editEntry?.config || EMPTY_CONFIGS[resourceType] || {};
  const [name, setName] = useState(editEntry?.name || "");
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const [config, setConfig] = useState<any>(initialConfig);

  const createEntry = useCreateRouteEntry();
  const updateEntry = useUpdateRouteEntry();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    const payload = {
      name: name.trim(),
      resource_type: resourceType,
      config,
    };

    if (isEdit) {
      updateEntry.mutate(
        { id: editEntry.id, ...payload },
        {
          onSuccess: () => {
            showToast("路由已更新", "success");
            onClose();
          },
          onError: () => showToast("更新失败", "error"),
        }
      );
    } else {
      createEntry.mutate(payload, {
        onSuccess: () => {
          showToast("路由已创建", "success");
          onClose();
        },
        onError: () => showToast("创建失败", "error"),
      });
    }
  };

  if (!open) return null;

  const isPending = createEntry.isPending || updateEntry.isPending;
  const label = TYPE_LABELS[resourceType] || resourceType;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/50" onClick={onClose} />
      <div className="relative w-full max-w-3xl max-h-[90vh] overflow-y-auto rounded-lg bg-white shadow-xl">
        <div className="sticky top-0 z-10 flex items-center justify-between border-b border-gray-200 bg-white px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">
            {isEdit ? `编辑 ${label}` : `新建 ${label}`}
          </h3>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">
              名称 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={`例如: my-${resourceType}`}
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div className="border-t border-gray-200 pt-4">
            {resourceType === "service" && (
              <ServiceForm key={`${editEntry?.id}-${editEntry?.updated_at}`} value={config} onChange={setConfig} />
            )}
            {resourceType === "ingress" && (
              <IngressForm key={`${editEntry?.id}-${editEntry?.updated_at}`} value={config} onChange={setConfig} />
            )}
            {resourceType === "ingressroute" && (
              <IngressRouteForm key={`${editEntry?.id}-${editEntry?.updated_at}`} value={config} onChange={setConfig} />
            )}
            {resourceType === "apisixroute" && (
              <ApisixRouteForm key={`${editEntry?.id}-${editEntry?.updated_at}`} value={config} onChange={setConfig} />
            )}
          </div>

          <div className="flex justify-end gap-3 border-t border-gray-200 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="rounded-md border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={!name.trim() || isPending}
              className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
            >
              {isPending
                ? isEdit
                  ? "更新中..."
                  : "创建中..."
                : isEdit
                  ? "更新"
                  : "创建"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
