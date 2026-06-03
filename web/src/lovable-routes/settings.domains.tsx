import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Globe, Search, ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export const Route = createFileRoute("/settings/domains")({
  component: DomainsPage,
  head: () => ({ meta: [{ title: "Домены — Исток" }] }),
});

function DomainsPage() {
  const [q, setQ] = useState("");
  return (
    <div className="mx-auto max-w-4xl px-6 py-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Домены рабочей области</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Подключайте уже купленные домены или приобретайте новые прямо из Истока.
          </p>
        </div>
        <Button className="bg-gradient-primary text-primary-foreground"><ShoppingCart /> Купить домен</Button>
      </div>

      <div className="mb-6 relative">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="Поиск по доменам..."
          className="pl-9"
        />
      </div>

      <div className="grid place-items-center rounded-2xl border border-dashed border-border/60 bg-card/20 px-6 py-20">
        <div className="grid h-14 w-14 place-items-center rounded-2xl bg-primary/10 text-primary">
          <Globe className="h-6 w-6" />
        </div>
        <h2 className="mt-4 text-base font-medium">Домены пока отсутствуют</h2>
        <p className="mt-1 max-w-md text-center text-sm text-muted-foreground">
          Купите новый домен или подключите существующий, чтобы публиковать проекты на своём адресе.
        </p>
        <Button className="mt-5 bg-gradient-primary text-primary-foreground"><ShoppingCart /> Купить домен</Button>
      </div>
    </div>
  );
}
