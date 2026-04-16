"use client";

interface YamlPreviewProps {
  yaml: string;
  loading?: boolean;
}

export function YamlPreview({ yaml, loading }: YamlPreviewProps) {
  if (loading) {
    return (
      <div className="flex items-center justify-center rounded-lg bg-gray-900 p-6">
        <div className="inline-block h-5 w-5 animate-spin rounded-full border-2 border-green-400 border-t-transparent" />
        <span className="ml-2 text-sm text-gray-400">加载预览中...</span>
      </div>
    );
  }

  return (
    <pre className="max-h-[400px] overflow-auto rounded-lg bg-gray-900 p-4 text-sm leading-relaxed text-green-400 font-mono">
      {yaml || "暂无预览内容"}
    </pre>
  );
}
