"use client";

import { useState, useEffect } from "react";

interface CodeEditorProps {
  content: string;
  onChange: (val: string) => void;
  format: "yaml" | "json";
  canEdit: boolean;
  saving?: boolean;
  onSave?: () => void;
}

export function CodeEditor({
  content,
  onChange,
  format,
  canEdit,
  saving,
  onSave,
}: CodeEditorProps) {
  const [localContent, setLocalContent] = useState(content);

  useEffect(() => {
    setLocalContent(content);
  }, [content]);

  const handleChange = (val: string) => {
    setLocalContent(val);
    onChange(val);
  };

  const lineCount = localContent.split("\n").length;

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium uppercase tracking-wider text-gray-500">
          {format.toUpperCase()} 编辑器
        </span>
        {canEdit && onSave && (
          <button
            onClick={onSave}
            disabled={saving}
            className="rounded-md bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? "保存中..." : "保存草稿"}
          </button>
        )}
      </div>
      <div className="overflow-hidden rounded-lg border border-gray-200">
        <div className="flex">
          {/* 行号 */}
          <div className="select-none bg-gray-50 px-3 py-3 text-right">
            {Array.from({ length: Math.max(lineCount, 20) }, (_, i) => (
              <div
                key={i}
                className="text-xs leading-[1.625rem] text-gray-400 font-mono"
              >
                {i + 1}
              </div>
            ))}
          </div>
          {/* 编辑区 */}
          <textarea
            value={localContent}
            onChange={(e) => handleChange(e.target.value)}
            readOnly={!canEdit}
            spellCheck={false}
            className="flex-1 resize-none bg-white px-3 py-3 font-mono text-sm leading-[1.625rem] text-gray-900 focus:outline-none"
            style={{ minHeight: `${Math.max(lineCount, 20) * 26}px` }}
          />
        </div>
      </div>
    </div>
  );
}
