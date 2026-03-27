import Editor from "@monaco-editor/react";

interface CodeEditorProps {
  code: string;
  onChange: (code: string) => void;
  language?: string;
}

const getLanguage = (filename?: string): string => {
  if (!filename) return "html";
  if (filename.endsWith(".html")) return "html";
  if (filename.endsWith(".css")) return "css";
  if (filename.endsWith(".js") || filename.endsWith(".jsx")) return "javascript";
  if (filename.endsWith(".ts") || filename.endsWith(".tsx")) return "typescript";
  if (filename.endsWith(".json")) return "json";
  if (filename.endsWith(".md")) return "markdown";
  return "html";
};

const CodeEditor = ({ code, onChange, language }: CodeEditorProps) => {
  return (
    <Editor
      height="100%"
      language={language ?? "html"}
      value={code}
      onChange={(value) => onChange(value ?? "")}
      theme="vs-dark"
      options={{
        fontSize: 13,
        lineHeight: 22,
        minimap: { enabled: false },
        scrollBeyondLastLine: false,
        padding: { top: 12, bottom: 12 },
        wordWrap: "on",
        tabSize: 2,
        automaticLayout: true,
        fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace",
        fontLigatures: true,
        renderLineHighlight: "line",
        cursorBlinking: "smooth",
        smoothScrolling: true,
        bracketPairColorization: { enabled: true },
      }}
    />
  );
};

export { getLanguage };
export default CodeEditor;
