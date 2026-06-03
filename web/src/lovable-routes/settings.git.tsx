import { createFileRoute } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/settings/git")({
  component: GitPage,
  head: () => ({ meta: [{ title: "Git — Исток" }] }),
});

function GitHubLogo({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} fill="currentColor">
      <path d="M12 .5C5.65.5.5 5.65.5 12c0 5.08 3.29 9.39 7.86 10.91.58.11.79-.25.79-.56v-2.1c-3.2.69-3.87-1.37-3.87-1.37-.53-1.34-1.29-1.7-1.29-1.7-1.05-.72.08-.71.08-.71 1.16.08 1.77 1.19 1.77 1.19 1.03 1.77 2.71 1.26 3.37.96.1-.75.4-1.26.73-1.55-2.55-.29-5.24-1.28-5.24-5.7 0-1.26.45-2.29 1.19-3.1-.12-.29-.52-1.47.11-3.06 0 0 .97-.31 3.18 1.18a11 11 0 0 1 5.79 0c2.2-1.49 3.17-1.18 3.17-1.18.63 1.59.23 2.77.11 3.06.74.81 1.19 1.84 1.19 3.1 0 4.43-2.7 5.4-5.27 5.69.41.36.78 1.06.78 2.14v3.17c0 .31.21.68.8.56C20.21 21.38 23.5 17.07 23.5 12 23.5 5.65 18.35.5 12 .5z"/>
    </svg>
  );
}
function GitLabLogo({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className}>
      <path fill="#E24329" d="m23.6 9.6-.03-.09L20.41.92a.83.83 0 0 0-.79-.55.84.84 0 0 0-.8.6l-2.13 6.55H7.32L5.18.97a.83.83 0 0 0-.8-.6.84.84 0 0 0-.8.55L.42 9.51l-.03.09a5.9 5.9 0 0 0 1.95 6.78l.01.01.03.02 4.83 3.62 2.4 1.81 1.45 1.1a.97.97 0 0 0 1.18 0l1.46-1.1 2.39-1.81 4.86-3.64.02-.01a5.9 5.9 0 0 0 1.95-6.78z"/>
      <path fill="#FC6D26" d="m23.6 9.6-.03-.09a10.7 10.7 0 0 0-4.27 1.92L12 17.05c2.42 1.83 4.53 3.42 4.53 3.42l4.86-3.64.02-.01a5.9 5.9 0 0 0 2.19-7.22z"/>
      <path fill="#FCA326" d="M7.47 20.47 9.86 22.28l1.45 1.1a.97.97 0 0 0 1.18 0l1.46-1.1 2.39-1.81S9.87 17.05 7.47 20.47z"/>
      <path fill="#FC6D26" d="M4.7 11.42A10.7 10.7 0 0 0 .43 9.5a5.9 5.9 0 0 0 2.18 7.22l.01.01.03.02 4.83 3.62c.4-.42 2.51-2.01 4.52-3.42z"/>
    </svg>
  );
}

function Card({ logo, name, desc }: { logo: React.ReactNode; name: string; desc: string }) {
  return (
    <div className="flex items-center gap-4 rounded-xl border border-border/60 bg-card/40 p-5">
      <div className="grid h-12 w-12 place-items-center rounded-xl bg-elevated">{logo}</div>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium">{name}</p>
        <p className="text-xs text-muted-foreground">{desc}</p>
      </div>
      <Button variant="outline" size="sm">Подключить</Button>
    </div>
  );
}

function GitPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-semibold">Git</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Двусторонняя синхронизация проектов с вашим Git-провайдером. Подключите аккаунт, чтобы автоматически коммитить изменения.
        </p>
      </div>

      <div className="space-y-3">
        <Card
          logo={<GitHubLogo className="h-6 w-6 text-foreground" />}
          name="GitHub"
          desc="Синхронизация с публичными и приватными репозиториями."
        />
        <Card
          logo={<GitLabLogo className="h-7 w-7" />}
          name="GitLab"
          desc="Поддержка GitLab.com и self-hosted установок."
        />
      </div>
    </div>
  );
}
