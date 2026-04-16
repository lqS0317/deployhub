"use client";

import { useState } from "react";

interface TemplateEditorProps {
  templateContent: string;
  onChange: (content: string) => void;
  onSave: () => void;
  onPreview?: () => void;
  saving?: boolean;
}

// 配置模板编辑器组件，支持 Go 模板语法高亮编辑
export function TemplateEditor({
  templateContent,
  onChange,
  onSave,
  onPreview,
  saving = false,
}: TemplateEditorProps) {
  const [lineCount, setLineCount] = useState(1);

  const handleChange = (value: string) => {
    onChange(value);
    setLineCount(value.split("\n").length);
  };

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-gray-700">模板内容</span>
        <div className="flex gap-2">
          {onPreview && (
            <button
              onClick={onPreview}
              className="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 transition-colors"
            >
              预览渲染
            </button>
          )}
          <button
            onClick={onSave}
            disabled={saving}
            className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            {saving ? "保存中..." : "保存"}
          </button>
        </div>
      </div>
      <div className="relative rounded-lg border border-gray-300 bg-gray-900 overflow-hidden">
        {/* 行号区域 */}
        <div className="absolute left-0 top-0 bottom-0 w-10 bg-gray-800 pointer-events-none select-none">
          <div className="pt-3 pr-2 text-right">
            {Array.from({ length: Math.max(lineCount, 1) }, (_, i) => (
              <div
                key={i}
                className="text-xs leading-5 text-gray-500 font-mono"
              >
                {i + 1}
              </div>
            ))}
          </div>
        </div>
        <textarea
          value={templateContent}
          onChange={(e) => handleChange(e.target.value)}
          className="w-full min-h-[300px] resize-y bg-transparent pl-12 pr-4 py-3 text-sm leading-5 text-green-400 font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-inset"
          placeholder="# 输入 Go 模板内容&#10;apiVersion: v1&#10;kind: ConfigMap&#10;metadata:&#10;  name: {{ .Name }}"
          spellCheck={false}
        />
      </div>
      <p className="text-xs text-gray-500">
        支持 Go 模板语法：{"{{ .VariableName }}"} 引用变量，{"{{ range }}"}{" "}
        循环遍历
      </p>
    </div>
  );
}
