import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers":
    "authorization, x-client-info, apikey, content-type, x-supabase-client-platform, x-supabase-client-platform-version, x-supabase-client-runtime, x-supabase-client-runtime-version",
};

serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response(null, { headers: corsHeaders });
  }

  try {
    const body = await req.json();
    const event = body?.event;
    const paymentData = body?.object;

    if (!event || !paymentData) {
      return new Response(
        JSON.stringify({ error: "Invalid webhook payload" }),
        { status: 400, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    const supabase = createClient(
      Deno.env.get("SUPABASE_URL")!,
      Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!
    );

    const externalId = paymentData.id;
    const status = paymentData.status; // "succeeded", "canceled", "waiting_for_capture"

    if (event === "payment.succeeded") {
      // Update payment status
      const { data: payment, error: paymentError } = await supabase
        .from("payments")
        .update({ status: "succeeded" })
        .eq("external_id", externalId)
        .select("user_id, amount")
        .single();

      if (paymentError) {
        console.error("Payment update error:", paymentError);
        return new Response(
          JSON.stringify({ error: "Failed to update payment" }),
          { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }

      // Activate subscription
      if (payment) {
        const periodEnd = new Date();
        periodEnd.setMonth(periodEnd.getMonth() + 1);

        await supabase
          .from("subscriptions")
          .upsert({
            user_id: payment.user_id,
            status: "active",
            plan_type: "business",
            current_period_end: periodEnd.toISOString(),
            updated_at: new Date().toISOString(),
          }, { onConflict: "user_id" });
      }
    } else if (event === "payment.canceled") {
      await supabase
        .from("payments")
        .update({ status: "canceled" })
        .eq("external_id", externalId);
    }

    return new Response(
      JSON.stringify({ success: true }),
      { headers: { ...corsHeaders, "Content-Type": "application/json" } }
    );
  } catch (e) {
    console.error("Webhook error:", e);
    return new Response(
      JSON.stringify({ error: e instanceof Error ? e.message : "Unknown error" }),
      { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
    );
  }
});
