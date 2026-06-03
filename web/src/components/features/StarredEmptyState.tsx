import { motion } from "framer-motion";
import { Star, Sparkles, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";

export function StarredEmptyState({ onBrowse }: { onBrowse: () => void }) {
  return (
    <div className="relative mx-auto flex min-h-[70vh] max-w-3xl flex-col items-center justify-center px-6 py-20 text-center">
      <div className="relative mb-10 h-56 w-72">
        <div
          aria-hidden
          className="absolute inset-0 -z-10 rounded-full bg-gradient-radial blur-3xl"
          style={{
            background:
              "radial-gradient(closest-side, oklch(0.55 0.25 305 / 0.45), transparent 70%)",
          }}
        />
        {[
          { rotate: -18, x: -90, y: 30, delay: 0, tint: "from-violet-500 to-fuchsia-600" },
          { rotate: 8, x: -10, y: -10, delay: 0.1, tint: "from-fuchsia-500 to-pink-500" },
          { rotate: 22, x: 80, y: 40, delay: 0.2, tint: "from-indigo-500 to-violet-700" },
        ].map((c, i) => (
          <motion.div
            key={i}
            initial={{ opacity: 0, y: 20, rotate: 0 }}
            animate={{
              opacity: 1,
              y: [c.y, c.y - 8, c.y],
              rotate: c.rotate,
            }}
            transition={{
              opacity: { duration: 0.4, delay: c.delay },
              y: { duration: 4 + i, repeat: Infinity, ease: "easeInOut", delay: c.delay },
              rotate: { duration: 0.6, delay: c.delay },
            }}
            style={{ translateX: c.x }}
            className="absolute left-1/2 top-1/2 h-40 w-28 -translate-x-1/2 -translate-y-1/2"
          >
            <div
              className={`relative h-full w-full overflow-hidden rounded-2xl bg-gradient-to-br ${c.tint} shadow-2xl ring-1 ring-white/20`}
            >
              <div className="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent" />
              <Sparkles className="absolute right-2 top-2 h-3 w-3 text-white/80" />
              <Star className="absolute bottom-2 left-2 h-3 w-3 fill-white/90 text-white/90" />
              <div className="absolute bottom-3 left-2 right-2 space-y-1">
                <div className="h-1.5 w-3/4 rounded bg-white/30" />
                <div className="h-1.5 w-1/2 rounded bg-white/20" />
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      <div className="inline-flex items-center gap-1.5 rounded-full border border-border/60 bg-elevated/60 px-3 py-1 text-[11px] text-muted-foreground">
        <Star className="h-3 w-3 text-amber-300" /> Избранное
      </div>
      <h1 className="mt-4 text-2xl font-semibold tracking-tight">Здесь пока нет избранных проектов</h1>
      <p className="mt-2 max-w-md text-sm text-muted-foreground">
        Добавляйте проекты в избранное, чтобы быстро открывать их из любого рабочего пространства.
      </p>
      <Button
        onClick={onBrowse}
        className="mt-6 bg-gradient-primary text-primary-foreground shadow-glow hover:opacity-90"
      >
        Перейти к проектам <ArrowRight className="h-4 w-4" />
      </Button>
    </div>
  );
}
