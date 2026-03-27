import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";
import { supabase } from "@/integrations/supabase/client";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Loader2, Users, TrendingUp, DollarSign, Flame, BarChart3, CreditCard, Wallet, ArrowUpRight, ArrowDownRight } from "lucide-react";

const ADMIN_ID = "9cab2dcf-e25d-490e-99f4-f2df9a925a02";
const AVG_TOKENS_PER_GENERATION = 4000;
const COST_PER_TOKEN_USD = 0.000015; // ~$15/1M tokens for Claude 3.5 Sonnet via OpenRouter
const USD_TO_RUB = 90;

interface AdminStats {
  totalUsers: number;
  todayUsers: number;
  yesterdayUsers: number;
  totalRevenue: number;
  totalCreditsRemaining: number;
  totalProjects: number;
  openRouterBalance: { total_credits: number; total_usage: number } | null;
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

  useEffect(() => {
    if (user?.id === ADMIN_ID) {
      fetchStats();
    }
  }, [user]);

  const fetchStats = async () => {
    try {
      const { data, error } = await supabase.functions.invoke("admin-stats");
      if (error) throw error;
      setStats(data);
    } catch (e) {
      console.error("Failed to fetch admin stats:", e);
    } finally {
      setLoading(false);
    }
  };

  if (authLoading || loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!user || user.id !== ADMIN_ID || !stats) return null;

  // Calculations
  const growthPercent = stats.yesterdayUsers > 0
    ? Math.round(((stats.todayUsers - stats.yesterdayUsers) / stats.yesterdayUsers) * 100)
    : stats.todayUsers > 0 ? 100 : 0;

  const revenueAfterTax = stats.totalRevenue * 0.9;
  const orBalance = stats.openRouterBalance;
  const orRemainingUsd = orBalance ? (orBalance.total_credits - orBalance.total_usage) : 0;
  const tokenCostRub = orBalance ? orBalance.total_usage * USD_TO_RUB : 0;
  const netProfit = revenueAfterTax - tokenCostRub;
  const roi = tokenCostRub > 0 ? ((netProfit / tokenCostRub) * 100).toFixed(1) : "∞";

  // Burn rate: days remaining at current pace
  const dailyGenerations = stats.recentGenerations > 0 ? stats.recentGenerations / 7 : 0;
  const dailyCostUsd = dailyGenerations * AVG_TOKENS_PER_GENERATION * COST_PER_TOKEN_USD;
  const burnDays = dailyCostUsd > 0 ? Math.floor(orRemainingUsd / dailyCostUsd) : 999;

  const burnColor = burnDays < 7 ? "text-destructive" : burnDays < 30 ? "text-yellow-500" : "text-green-500";

  return (
    <div className="min-h-screen bg-background text-foreground p-6 md:p-10">
      <div className="max-w-7xl mx-auto space-y-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Админ-панель</h1>
            <p className="text-muted-foreground mt-1">Сводка системы Исток AI</p>
          </div>
          <button onClick={() => navigate("/")} className="text-sm text-muted-foreground hover:text-foreground transition-colors">
            ← На главную
          </button>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {/* Users */}
          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Пользователи</CardTitle>
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
                  {growthPercent >= 0 ? "+" : ""}{growthPercent}%
                </span>
                <span className="text-muted-foreground ml-1">за 24ч (сегодня: {stats.todayUsers}, вчера: {stats.yesterdayUsers})</span>
              </div>
            </CardContent>
          </Card>

          {/* Projects */}
          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Проекты</CardTitle>
              <BarChart3 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.totalProjects}</div>
              <p className="text-xs text-muted-foreground mt-1">Всего сгенерировано</p>
            </CardContent>
          </Card>

          {/* OpenRouter Balance */}
          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Баланс OpenRouter</CardTitle>
              <Wallet className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                ${orRemainingUsd.toFixed(2)}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                Потрачено: ${orBalance?.total_usage?.toFixed(2) || "0.00"}
              </p>
            </CardContent>
          </Card>

          {/* Burn Rate */}
          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Burn Rate</CardTitle>
              <Flame className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className={`text-2xl font-bold ${burnColor}`}>
                {burnDays >= 999 ? "∞" : `${burnDays} дн.`}
              </div>
              <p className="text-xs text-muted-foreground mt-1">
                ~{dailyGenerations.toFixed(1)} генераций/день · ${dailyCostUsd.toFixed(3)}/день
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Financial Section */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {/* Revenue & Profit */}
          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Юнит-экономика</CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Выручка (payments)</span>
                <span className="font-mono font-semibold">{stats.totalRevenue.toLocaleString("ru-RU")} ₽</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Выручка после налога (×0.9)</span>
                <span className="font-mono font-semibold">{revenueAfterTax.toLocaleString("ru-RU")} ₽</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Затраты на токены (в ₽)</span>
                <span className="font-mono font-semibold text-destructive">-{tokenCostRub.toFixed(0)} ₽</span>
              </div>
              <div className="h-px bg-border my-2" />
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">Чистая прибыль</span>
                <span className={`font-mono font-bold text-lg ${netProfit >= 0 ? "text-green-500" : "text-destructive"}`}>
                  {netProfit >= 0 ? "+" : ""}{netProfit.toFixed(0)} ₽
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">ROI</span>
                <span className="font-mono font-bold text-primary">{roi}%</span>
              </div>
            </CardContent>
          </Card>

          {/* Credits Overview */}
          <Card className="border-border/50 bg-card">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Кредиты пользователей</CardTitle>
              <CreditCard className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Остаток у всех пользователей</span>
                <span className="font-mono font-semibold">{stats.totalCreditsRemaining.toLocaleString("ru-RU")} токенов</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Генераций за 7 дней</span>
                <span className="font-mono font-semibold">{stats.recentGenerations}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-muted-foreground">Внесено на OpenRouter</span>
                <span className="font-mono font-semibold">${stats.totalDepositedUsd}</span>
              </div>

              {/* Burn Rate Visual */}
              <div className="mt-4 pt-3 border-t border-border">
                <div className="flex items-center gap-2 mb-2">
                  <Flame className={`h-4 w-4 ${burnColor}`} />
                  <span className="text-sm font-medium">Прогноз исчерпания баланса</span>
                </div>
                <div className="w-full bg-secondary rounded-full h-3 overflow-hidden">
                  <div
                    className={`h-full rounded-full transition-all duration-500 ${
                      burnDays < 7 ? "bg-destructive" : burnDays < 30 ? "bg-yellow-500" : "bg-green-500"
                    }`}
                    style={{ width: `${Math.min(100, (burnDays / 90) * 100)}%` }}
                  />
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  {burnDays >= 999 ? "Нет активности — баланс не расходуется" : `Хватит примерно на ${burnDays} дней при текущем темпе`}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
};

export default Admin;
