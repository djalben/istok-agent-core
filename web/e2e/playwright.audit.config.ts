import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: ".",
  testMatch: "toolbar-audit.spec.ts",
  timeout: 30000,
  use: {
    baseURL: "http://localhost:8080",
    headless: true,
    viewport: { width: 1280, height: 720 },
    screenshot: "on",
  },
  reporter: "list",
});
