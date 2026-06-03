import { useEffect, useState, type ReactNode } from "react";
import { useNavigate } from "@tanstack/react-router";
import { Menu, Sparkles } from "lucide-react";
import {
  DashboardSidebar, type DashboardSection,
} from "@/components/features/DashboardSidebar";
import { CommandPalette } from "@/components/features/CommandPalette";
import { ConnectorsModal } from "@/components/features/ConnectorsModal";
import { ReferralModal } from "@/components/features/ReferralModal";
import { TooltipProvider } from "@/components/ui/tooltip";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";

interface DashboardLayoutProps {
  active: DashboardSection;
  onSelectSection?: (s: DashboardSection) => void;
  children: ReactNode;
}

export function DashboardLayout({ active, onSelectSection, children }: DashboardLayoutProps) {
  const navigate = useNavigate();
  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [paletteOpen, setPaletteOpen] = useState(false);
  const [connectorsOpen, setConnectorsOpen] = useState(false);
  const [referralOpen, setReferralOpen] = useState(false);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const mod = e.metaKey || e.ctrlKey;
      if (mod && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setPaletteOpen((v) => !v);
      }
      if (mod && e.key.toLowerCase() === "b") {
        e.preventDefault();
        setCollapsed((v) => !v);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  const handleSelect = (s: DashboardSection) => {
    setMobileOpen(false);
    if (s === "connectors") {
      setConnectorsOpen(true);
      return;
    }
    if (s === "resources") {
      navigate({ to: "/resources" });
      return;
    }
    if (active === "resources") {
      navigate({ to: "/" });
    }
    onSelectSection?.(s);
  };

  const sidebarProps = {
    active,
    onSelect: handleSelect,
    onOpenSearch: () => { setPaletteOpen(true); setMobileOpen(false); },
    onSelectProject: (id: string) => { setMobileOpen(false); navigate({ to: "/builder/$id", params: { id } }); },
    onShareLovable: () => { setReferralOpen(true); setMobileOpen(false); },
  };

  return (
    <TooltipProvider delayDuration={150}>
      <div className="flex min-h-screen bg-background">
        <DashboardSidebar
          {...sidebarProps}
          collapsed={collapsed}
          onToggle={() => setCollapsed((v) => !v)}
        />

        <div className="flex min-w-0 flex-1 flex-col">
          {/* Mobile header */}
          <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-border/60 bg-background/80 px-4 backdrop-blur-xl md:hidden">
            <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
              <SheetTrigger asChild>
                <Button variant="ghost" size="icon" className="h-9 w-9" aria-label="Открыть меню">
                  <Menu className="h-5 w-5" />
                </Button>
              </SheetTrigger>
              <SheetContent side="left" className="w-[280px] p-0">
                <DashboardSidebar
                  {...sidebarProps}
                  collapsed={false}
                  onToggle={() => setMobileOpen(false)}
                  mobile
                />
              </SheetContent>
            </Sheet>
            <div className="flex items-center gap-2">
              <div className="grid h-7 w-7 place-items-center rounded-md bg-gradient-primary shadow-glow">
                <Sparkles className="h-3.5 w-3.5 text-primary-foreground" />
              </div>
              <span className="text-sm font-semibold">Исток</span>
            </div>
            <div className="w-9" />
          </header>

          <main className="min-w-0 flex-1">{children}</main>
        </div>

        <CommandPalette open={paletteOpen} onOpenChange={setPaletteOpen} />
        <ConnectorsModal open={connectorsOpen} onOpenChange={setConnectorsOpen} />
        <ReferralModal open={referralOpen} onOpenChange={setReferralOpen} />
      </div>
    </TooltipProvider>
  );
}
