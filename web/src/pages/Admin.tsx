import { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Loader2,
  Users,
  TrendingUp,
  DollarSign,
  Flame,
  BarChart3,
  CreditCard,
  ArrowUpRight,
  ArrowDownRight,
} from "lucide-react";
import MainLayout from "@/components/layout/MainLayout";

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Admin Dashboard
//  Anthropic Direct + Replicate-only stack (OpenRouter удалён).
//  TODO: подключить Go API endpoints для реальной статистики.
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

const ADMIN_ID = "9cab2dcf-e25d-490e-99f4-f2df9a925a02";
const AVG_TOKENS_PER_GENERATION = 4000;
/** Anthropic Claude 3.7 Sonnet pricing: $3/MTok input, $15/MTok output. Avg blended ≈ $9/MTok. */
const COST_PER_TOKEN_USD = 0.000009;
const USD_TO_RUB = 90;

interface AdminStats {
  totalUsers: number;
  todayUsers: number;
  yesterdayUsers: number;
  totalRevenue: number;
  totalCreditsRemaining: number;
  totalProjects: number;
  totalDepositedUsd: number;
  recentGenerations: number;
}

const Admin = () => {
  const { user, loading: authLoading } = useAuth();
  const navigate = useNavigate();
  const [stats, setStats] = useState<AdminStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!authLoading && (!user || user.id !== ADMIN_ID)) {
      navigate("/", { replace: true });
    }
  }, [user, authLoading, navigate]);

  const fetchStats = useCallback(async () => {
    try {
      // TODO: Загрузка статистики через Go API (/api/v1/admin/stats)
      setStats({
        totalUsers: 0,
        todayUsers: 0,
        yesterdayUsers: 0,
        totalRevenue: 0,
        totalCreditsRemaining: 0,
        totalProjects: 0,
        totalDepositedUsd: 0,
        recentGenerations: 0,
      });
    } catch (e) {
      console.error("Failed to fetch admin stats:", e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (user?.id === ADMIN_ID) {
      fetchStats();
    }
  }, [user, fetchStats]);

  if (authLoading || loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!user || user.id !== ADMIN_ID || !stats) return null;

  // ── Computed KPIs ──────────────────────────────────
  const growthPercent =
    stats.yesterdayUsers > 0
      ? Math.round(((stats.todayUsers - stats.yesterdayUsers) / stats.yesterdayUsers) * 100)
      : stats.todayUsers > 0
        ? 100
        : 0;

  const revenueAfterTax = stats.totalRevenue * 0.9;

  const dailyGenerations = stats.recentGenerations > 0 ? stats.recentGenerations / 7 : 0;
  const dailyCostUsd = dailyGenerations * AVG_TOKENS_PER_GENERATION * COST_PER_TOKEN_USD;
  const dailyCostRub = dailyCostUsd * USD_TO_RUB;
  const monthlyCostRub = dailyCostRub * 30;
  const netProfit = revenueAfterTax - monthlyCostRub;
  const roi = monthlyCostRub > 0 ? ((netProfit / monthlyCostRub) * 100).toFixed(1) : "∞";

  return (
    <MainLayout>
      <div className="max-w-7xl mx-auto px-4 md:px-6 py-6 md:py-8 space-y-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Админ-панель</h1>
            <p className="text-muted-foreground mt-1">Сводка системы Исток AI</p>
          </div>
          <button
            onClick={() => navigate("/")}
            className="text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            ← На главную
          </button>
        </div>

        {/* ── KPI Grid ── */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Пользователи
              </CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalUsers}</div>
              <div className="flex items-center text-xs mt-1">
                {growthPercent >= 0 ? (
                  <ArrowUpRight className="h-3 w-3 text-green-500 mr-1" />
                ) : (
                  <ArrowDownRight className="h-3 w-3 text-destructive mr-1" />
                )}
                <span className={growthPercent >= 0 ? "text-green-500" : "text-destructive"}>
                  {growthPercent >= 0 ? "+" : ""}
                  {growthPercent}%
                </span>
                <span className="text-muted-foreground ml-1">
                  за 24ч (сегодня: {stats.todayUsers}, вчера: {stats.yesterdayUsers})
                </span>
              </div>
            </CardContent>
          </Card>

          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Проекты
              </CardTitle>
              <BarChart3 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalProjects}</div>
              <p className="text-xs text-muted-foreground mt-1">Всего сгенерировано</p>
            </CardContent>
          </Card>

          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Расход (Anthropic)
              </CardTitle>
              <Flame className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">${dailyCostUsd.toFixed(3)}</div>
              <p className="text-xs text-muted-foreground mt-1">
                ~{dailyGenerations.toFixed(1)} генераций/день · Claude 3.7
              </p>
            </CardContent>
          </Card>

          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Прирост за 7д
              </CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.recentGenerations}</div>
              <p className="text-xs text-muted-foreground mt-1">генераций</p>
            </CardContent>
          </Card>
        </div>

        {/* ── Financial Section ── */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Юнит-экономика (мес.)
              </CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent className="space-y-3">
              <Row label="Выручка (payments)" value={`${stats.totalRevenue.toLocaleString("ru-RU")} ₽`} />
              <Row label="Выручка после налога (×0.9)" value={`${revenueAfterTax.toLocaleString("ru-RU")} ₽`} />
              <Row
                label="Затраты на токены (Anthropic)"
                value={`-${monthlyCostRub.toFixed(0)} ₽`}
                tone="negative"
              />
              <div className="h-px bg-border my-2" />
              <Row
                label="Чистая прибыль"
                value={`${netProfit >= 0 ? "+" : ""}${netProfit.toFixed(0)} ₽`}
                tone={netProfit >= 0 ? "positive" : "negative"}
                bold
              />
              <Row label="ROI" value={`${roi}%`} tone="primary" bold />
            </CardContent>
          </Card>

          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Кредиты пользователей
              </CardTitle>
              <CreditCard className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent className="space-y-3">
              <Row
                label="Остаток у всех пользователей"
                value={`${stats.totalCreditsRemaining.toLocaleString("ru-RU")} токенов`}
              />
              <Row label="Генераций за 7 дней" value={`${stats.recentGenerations}`} />
              <Row label="Внесено пользователями" value={`$${stats.totalDepositedUsd}`} />
            </CardContent>
          </Card>
        </div>
      </div>
    </MainLayout>
  );
};

interface RowProps {
  label: string;
  value: string;
  tone?: "neutral" | "positive" | "negative" | "primary";
  bold?: boolean;
}

const Row = ({ label, value, tone = "neutral", bold = false }: RowProps) => {
  const toneClass =
    tone === "positive"
      ? "text-green-500"
      : tone === "negative"
        ? "text-destructive"
        : tone === "primary"
          ? "text-primary"
          : "text-foreground";
  return (
    <div className="flex justify-between items-center">
      <span className={`text-sm ${bold ? "font-medium" : "text-muted-foreground"}`}>{label}</span>
      <span className={`font-mono ${bold ? "font-bold text-lg" : "font-semibold"} ${toneClass}`}>
        {value}
      </span>
    </div>
  );
};

export default Admin;
