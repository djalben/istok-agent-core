import { createFileRoute } from "@tanstack/react-router";
import { BuilderShell } from "@/components/features/BuilderShell";

export const Route = createFileRoute("/builder/")({
  head: () => ({
    meta: [
      { title: "Новый проект · Исток" },
      { name: "description", content: "Начните новый full-stack проект из одного промта. Команда агентов Истока спланирует, соберёт и запустит его." },
    ],
  }),
  component: NewBuilder,
});

function NewBuilder() {
  return <BuilderShell mode="clean" projectName="Без названия" />;
}
