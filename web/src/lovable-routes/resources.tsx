import { createFileRoute } from "@tanstack/react-router";
import { DashboardLayout } from "@/components/features/DashboardLayout";
import { ResourcesGrid } from "@/components/features/ResourcesGrid";
import { BackButton } from "@/components/features/BackButton";

export const Route = createFileRoute("/resources")({
  head: () => ({
    meta: [
      { title: "Ресурсы — шаблоны Истока" },
      { name: "description", content: "Изучите готовые шаблоны для быстрого старта проектов в Истоке." },
      { property: "og:title", content: "Ресурсы — шаблоны Истока" },
      { property: "og:description", content: "Готовые шаблоны: SaaS-дашборды, портфолио, маркетплейсы и многое другое." },
    ],
  }),
  component: ResourcesPage,
});

function ResourcesPage() {
  return (
    <DashboardLayout active="resources">
      <div className="px-6 pt-4">
        <BackButton to="/" />
      </div>
      <ResourcesGrid />
    </DashboardLayout>
  );
}
