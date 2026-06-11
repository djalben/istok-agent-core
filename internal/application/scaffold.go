package application

import "maps"

// scaffoldFiles — жёстко зашитый, гарантированный React 18 + Vite + Tailwind каркас.
// Сидируется в хранилище файлов сессии ДО запуска Кодера, чтобы фундаментальная
// React-оболочка (entry point, mount, configs) существовала всегда — даже если LLM
// исчерпает контекст на больших генерациях (100+ файлов) и забудет создать main.tsx.
// Кодер перезаписывает src/App.tsx своим layout'ом; остальные файлы остаются как есть.
var scaffoldFiles = map[string]string{
	"index.html": `<!doctype html>
<html lang="ru">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Istok App</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
`,

	"src/main.tsx": `import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./index.css";

const rootEl = document.getElementById("root");
if (rootEl) {
  createRoot(rootEl).render(
    <StrictMode>
      <App />
    </StrictMode>,
  );
}
`,

	"src/App.tsx": `export default function App() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-zinc-950 text-zinc-100">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-tight">Загрузка приложения…</h1>
        <p className="mt-2 text-zinc-400">Кодер Истока наполняет интерфейс.</p>
      </div>
    </div>
  );
}
`,

	"src/index.css": `@tailwind base;
@tailwind components;
@tailwind utilities;
`,

	"vite.config.ts": `import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
`,

	"tailwind.config.ts": `import type { Config } from "tailwindcss";

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {},
  },
  plugins: [],
} satisfies Config;
`,

	"postcss.config.js": `export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
};
`,

	"tsconfig.json": `{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"]
}
`,
}

// ScaffoldFiles возвращает СВЕЖУЮ копию каркаса (чтобы вызывающий мог безопасно
// мутировать/перезаписывать файлы, не затрагивая общий шаблон).
func ScaffoldFiles() map[string]string {
	return maps.Clone(scaffoldFiles)
}
