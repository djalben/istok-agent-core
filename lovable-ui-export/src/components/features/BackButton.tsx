import { ArrowLeft } from "lucide-react";
import { useRouter } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface BackButtonProps {
  to?: string;
  label?: string;
  className?: string;
}

export function BackButton({ to, label = "Назад", className }: BackButtonProps) {
  const router = useRouter();
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => {
        if (to) router.navigate({ to: to as "/" });
        else router.history.back();
      }}
      className={cn(
        "group h-8 gap-1.5 px-2 text-xs text-muted-foreground hover:bg-muted/40 hover:text-foreground",
        className,
      )}
    >
      <ArrowLeft className="h-3.5 w-3.5 transition-transform group-hover:-translate-x-0.5" />
      {label}
    </Button>
  );
}
