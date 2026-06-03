import { createFileRoute, notFound } from "@tanstack/react-router";
import { TopBar } from "@/components/features/TopBar";
import { BuilderShell } from "@/components/features/BuilderShell";
import { mockProjects } from "@/lib/mockData";

export const Route = createFileRoute("/builder/$id")({
  loader: ({ params }) => {
    const project = mockProjects.find((p) => p.id === params.id);
    if (!project) throw notFound();
    return { project };
  },
  head: ({ loaderData }) => ({
    meta: [
      { title: `${loaderData?.project.name ?? "Проект"} · Исток` },
      { name: "description", content: loaderData?.project.description ?? "Рабочая область сборщика Истока." },
    ],
  }),
  notFoundComponent: () => {
    const { id } = Route.useParams();
    return (
      <div className="flex h-screen flex-col bg-background">
        <TopBar showBack />
        <div className="grid flex-1 place-items-center px-6 text-center">
          <div>
            <h1 className="text-2xl font-semibold">Проект не найден</h1>
            <p className="mt-2 text-sm text-muted-foreground">Проект с идентификатором «{id}» не существует.</p>
          </div>
        </div>
      </div>
    );
  },
  errorComponent: ({ error }) => (
    <div className="flex h-screen flex-col bg-background">
      <TopBar showBack />
      <div className="grid flex-1 place-items-center px-6 text-center">
        <p className="text-sm text-muted-foreground">{error.message}</p>
      </div>
    </div>
  ),
  component: ProjectBuilder,
});

function ProjectBuilder() {
  const { project } = Route.useLoaderData();
  return <BuilderShell mode="project" projectName={project.name} />;
}
