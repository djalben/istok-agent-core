import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2.49.1";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers":
    "authorization, x-client-info, apikey, content-type, x-supabase-client-platform, x-supabase-client-platform-version, x-supabase-client-runtime, x-supabase-client-runtime-version",
};

const ADMIN_ID = "9cab2dcf-e25d-490e-99f4-f2df9a925a02";

serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response(null, { headers: corsHeaders });
  }

  try {
    const authHeader = req.headers.get("authorization");
    if (!authHeader) {
      return new Response(JSON.stringify({ error: "Unauthorized" }), {
        status: 401, headers: { ...corsHeaders, "Content-Type": "application/json" },
      });
    }

    const supabaseUrl = Deno.env.get("SUPABASE_URL")!;
    const supabase = createClient(supabaseUrl, Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!);

    // Verify admin
    const token = authHeader.replace("Bearer ", "");
    const { data: { user } } = await createClient(
      supabaseUrl, Deno.env.get("SUPABASE_ANON_KEY")!
    ).auth.getUser(token);

    if (!user || user.id !== ADMIN_ID) {
      return new Response(JSON.stringify({ error: "Forbidden" }), {
        status: 403, headers: { ...corsHeaders, "Content-Type": "application/json" },
      });
    }

    // 1. Total users
    const { count: totalUsers } = await supabase
      .from("profiles")
      .select("*", { count: "exact", head: true });

    // 2. Growth 24h - registrations today vs yesterday
    const now = new Date();
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).toISOString();
    const yesterdayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1).toISOString();

    const { count: todayUsers } = await supabase
      .from("profiles")
      .select("*", { count: "exact", head: true })
      .gte("created_at", todayStart);

    const { count: yesterdayUsers } = await supabase
      .from("profiles")
      .select("*", { count: "exact", head: true })
      .gte("created_at", yesterdayStart)
      .lt("created_at", todayStart);

    // 3. Total successful payments
    const { data: payments } = await supabase
      .from("payments")
      .select("amount")
      .eq("status", "succeeded");

    const totalRevenue = (payments || []).reduce((sum, p) => sum + p.amount, 0);

    // 4. Total tokens used (sum of all credits spent = initial credits given minus remaining)
    const { data: allProfiles } = await supabase
      .from("profiles")
      .select("credits");
    const totalCreditsRemaining = (allProfiles || []).reduce((sum, p) => sum + p.credits, 0);

    // 5. Total projects
    const { count: totalProjects } = await supabase
      .from("projects")
      .select("*", { count: "exact", head: true });

    // 6. OpenRouter balance
    let openRouterBalance = null;
    const OPENROUTER_API_KEY = Deno.env.get("OPENROUTER_API_KEY");
    if (OPENROUTER_API_KEY) {
      try {
        const res = await fetch("https://openrouter.ai/api/v1/credits", {
          headers: { Authorization: `Bearer ${OPENROUTER_API_KEY}` },
        });
        if (res.ok) {
          const creditsData = await res.json();
          openRouterBalance = creditsData.data;
        }
      } catch (e) {
        console.error("OpenRouter credits fetch error:", e);
      }
    }

    // 7. Admin finances
    const { data: finances } = await supabase
      .from("admin_finances")
      .select("total_deposited_usd")
      .limit(1)
      .single();

    // 8. Daily token usage for burn rate (last 7 days approximation from payments/generation activity)
    // Approximate: count projects created in last 7 days as proxy for generations
    const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString();
    const { count: recentGenerations } = await supabase
      .from("projects")
      .select("*", { count: "exact", head: true })
      .gte("created_at", weekAgo);

    return new Response(JSON.stringify({
      totalUsers: totalUsers || 0,
      todayUsers: todayUsers || 0,
      yesterdayUsers: yesterdayUsers || 0,
      totalRevenue,
      totalCreditsRemaining,
      totalProjects: totalProjects || 0,
      openRouterBalance,
      totalDepositedUsd: finances?.total_deposited_usd || 0,
      recentGenerations: recentGenerations || 0,
    }), {
      headers: { ...corsHeaders, "Content-Type": "application/json" },
    });
  } catch (e) {
    console.error("admin-stats error:", e);
    return new Response(JSON.stringify({ error: e instanceof Error ? e.message : "Unknown error" }), {
      status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" },
    });
  }
});
