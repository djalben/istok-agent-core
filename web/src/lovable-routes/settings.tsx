import { createFileRoute, Outlet } from "@tanstack/react-router";
import { SettingsSidebar } from "@/components/features/SettingsSidebar";

export const Route = createFileRoute("/settings")({
  component: SettingsLayout,
  head: () => ({ meta: [{ title: "Настройки — Исток" }] }),
});

function SettingsLayout() {
  return (
    <div className="flex min-h-screen bg-background">
      <SettingsSidebar />
      <main className="min-w-0 flex-1 overflow-y-auto">
        <Outlet />
      </main>
    </div>
  );
}
