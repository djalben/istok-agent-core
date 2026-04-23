/**
 * ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 *  Instant Feedback — Sandpack-ready Preview Component
 *  Показывает превью фронтенда мгновенно с моками API.
 *  Когда Sandpack подключен — рендерит React/JS live.
 *  До подключения — использует iframe с inline HTML.
 * ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 */

import { useMemo } from "react";
import type { ProjectFiles } from "./WorkspacePreview";

// API Mock — подставляется в preview пока реальный бэкенд деплоится
const API_MOCK_SCRIPT = `
<script data-istok-mock>
(function() {
  // Mock fetch для превью — возвращает заглушки вместо реальных API вызовов
  const originalFetch = window.fetch;
  window.fetch = function(url, opts) {
    if (typeof url === 'string' && url.includes('/api/')) {
      console.log('[MOCK] Intercepted API call:', url);
      
      // Health check
      if (url.includes('/health')) {
        return Promise.resolve(new Response(JSON.stringify({
          status: 'healthy', service: 'istok-mock', uptime: '0s'
        }), { headers: { 'Content-Type': 'application/json' } }));
      }
      
      // Generate — return mock SSE
      if (url.includes('/generate')) {
        const stream = new ReadableStream({
          start(controller) {
            const enc = new TextEncoder();
            controller.enqueue(enc.encode('event: status\\ndata: {"agent":"system","status":"started","message":"Mock preview active","progress":100}\\n\\n'));
            controller.enqueue(enc.encode('event: done\\ndata: {"message":"Mock complete"}\\n\\n'));
            controller.close();
          }
        });
        return Promise.resolve(new Response(stream, {
          headers: { 'Content-Type': 'text/event-stream' }
        }));
      }
      
      // Default mock
      return Promise.resolve(new Response(JSON.stringify({ mock: true }), {
        headers: { 'Content-Type': 'application/json' }
      }));
    }
    return originalFetch.apply(this, arguments);
  };
  console.log('[ISTOK] API Mock active — preview mode');
})();
</script>
`;

interface SandpackPreviewProps {
  files: ProjectFiles;
  useMocks?: boolean;
  className?: string;
}

/**
 * SandpackPreview — компонент мгновенного превью.
 * Phase 1: iframe + API mocks (текущая реализация)
 * Phase 2: Sandpack integration (TODO: npm install @codesandbox/sandpack-react)
 */
export default function SandpackPreview({ files, useMocks = true, className = "" }: SandpackPreviewProps) {
  const previewHtml = useMemo(() => {
    let html = files["index.html"] || "<html><body><h1>No preview</h1></body></html>";
    
    // Inject API mocks before closing </head> or </body>
    if (useMocks) {
      if (html.includes("</head>")) {
        html = html.replace("</head>", API_MOCK_SCRIPT + "</head>");
      } else if (html.includes("</body>")) {
        html = html.replace("</body>", API_MOCK_SCRIPT + "</body>");
      }
    }

    return html;
  }, [files, useMocks]);

  const srcDoc = useMemo(() => {
    return previewHtml;
  }, [previewHtml]);

  return (
    <iframe
      srcDoc={srcDoc}
      title="Instant Preview"
      className={`w-full h-full border-0 bg-white ${className}`}
      sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
    />
  );
}

/**
 * TODO Phase 2: Replace iframe with Sandpack when dependency is added.
 * 
 * import { SandpackProvider, SandpackPreview } from "@codesandbox/sandpack-react";
 * 
 * <SandpackProvider
 *   files={Object.fromEntries(
 *     Object.entries(files).map(([name, code]) => [`/${name}`, code])
 *   )}
 *   template="vanilla"
 *   theme="dark"
 * >
 *   <SandpackPreview showNavigator={false} />
 * </SandpackProvider>
 */
