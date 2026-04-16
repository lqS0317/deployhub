"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api-client";
import { useService } from "@/hooks/use-services";
import { EnvSidebar } from "@/components/config/env-sidebar";
import { EntryList } from "@/components/config/entry-list";
import { EntryDetail } from "@/components/config/entry-detail";
import { CreateEntryDialog } from "@/components/config/create-entry-dialog";
import type { Cluster } from "@/types";

export default function ServiceConfigPage() {
  const params = useParams();
  const router = useRouter();
  const serviceId = Number(params.serviceId);

  const { data: service, isLoading: loadingService } = useService(serviceId);

  const { data: clustersData } = useQuery({
    queryKey: ["clusters"],
    queryFn: async () => {
      const res = await apiClient.get("/clusters");
      return res.data;
    },
  });
  const clusters: Cluster[] = clustersData?.items ?? [];

  const [selectedClusterId, setSelectedClusterId] = useState<number>(0);
  const [selectedEntryId, setSelectedEntryId] = useState<number | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  useEffect(() => {
    if (clusters.length > 0 && selectedClusterId === 0) {
      setSelectedClusterId(clusters[0].id);
    }
  }, [clusters, selectedClusterId]);

  // 切换集群时重置已选条目
  const handleSelectCluster = (id: number) => {
    setSelectedClusterId(id);
    setSelectedEntryId(null);
  };

  if (loadingService) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <div className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-blue-600 border-t-transparent" />
          <p className="mt-2 text-sm text-gray-500">加载中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="-m-6 flex h-[calc(100vh-0px)]">
      {/* 左侧边栏 */}
      <EnvSidebar
        clusters={clusters}
        selectedClusterId={selectedClusterId}
        onSelectCluster={handleSelectCluster}
      />

      {/* 右侧内容 */}
      <div className="flex flex-1 flex-col overflow-hidden">
        {/* 面包屑 */}
        <div className="flex items-center gap-2 border-b border-gray-200 bg-white px-6 py-2.5">
          <button
            onClick={() => router.push("/configs")}
            className="text-sm text-blue-600 hover:text-blue-800"
          >
            配置中心
          </button>
          <span className="text-gray-400">/</span>
          <span className="text-sm font-medium text-gray-900">
            {service?.display_name || service?.name || `服务 #${serviceId}`}
          </span>
        </div>

        {/* 内容区域 */}
        {!selectedClusterId ? (
          <div className="flex flex-1 items-center justify-center">
            <p className="text-sm text-gray-400">
              {clusters.length === 0
                ? "暂无可用集群，请先在设置中添加集群"
                : "请选择环境"}
            </p>
          </div>
        ) : selectedEntryId ? (
          <EntryDetail
            entryId={selectedEntryId}
            serviceId={serviceId}
            clusterId={selectedClusterId}
            onBack={() => setSelectedEntryId(null)}
          />
        ) : (
          <div className="flex-1 overflow-y-auto p-6">
            <EntryList
              serviceId={serviceId}
              clusterId={selectedClusterId}
              selectedEntryId={selectedEntryId}
              onSelect={setSelectedEntryId}
              onCreateClick={() => setShowCreateDialog(true)}
            />
          </div>
        )}
      </div>

      {/* 新建条目弹窗 */}
      <CreateEntryDialog
        open={showCreateDialog}
        onClose={() => setShowCreateDialog(false)}
        serviceId={serviceId}
        clusterId={selectedClusterId}
      />
    </div>
  );
}
