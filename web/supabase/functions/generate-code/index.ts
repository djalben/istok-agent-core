import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2.49.1";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers":
    "authorization, x-client-info, apikey, content-type, x-supabase-client-platform, x-supabase-client-platform-version, x-supabase-client-runtime, x-supabase-client-runtime-version",
};

const SYSTEM_PROMPT = `Ты — Исток AI, эксперт-архитектор по веб-разработке. Ты ВСЕГДА разбиваешь код на логические файлы.

ФОРМАТ ОТВЕТА — СТРОГО JSON в блоке \`\`\`json:

\`\`\`json
{
  "index.html": "<!DOCTYPE html>...",
  "styles.css": "body { ... }",
  "script.js": "// code"
}
\`\`\`

ПРАВИЛА:
1. Ответ ОБЯЗАТЕЛЬНО содержит блок \`\`\`json с объектом файлов. Ничего кроме JSON-блока.
2. Минимум: index.html + styles.css + script.js. Для сложных проектов добавляй файлы.
3. В index.html подключай: <link rel="stylesheet" href="styles.css"> и <script src="script.js"></script>
4. Шрифты Inter/Montserrat через Google Fonts.
5. Тёмный современный дизайн, адаптивная вёрстка, CSS-переменные.
6. Интерфейс на русском, код комментируй на английском.
7. Каждый файл полноценный — никаких TODO/плейсхолдеров.
8. Контекст РФ: даты ДД.ММ.ГГГГ, ₽, +7, VK/Telegram/Yandex.`;

serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response(null, { headers: corsHeaders });
  }

  try {
    const { messages } = await req.json();

    const authHeader = req.headers.get("authorization");
    if (!authHeader) {
      return new Response(
        JSON.stringify({ error: "Требуется авторизация" }),
        { status: 401, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    const supabaseUrl = Deno.env.get("SUPABASE_URL")!;
    const supabaseServiceKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!;
    const supabase = createClient(supabaseUrl, supabaseServiceKey);

    const token = authHeader.replace("Bearer ", "");
    const { data: { user }, error: userError } = await createClient(
      supabaseUrl,
      Deno.env.get("SUPABASE_ANON_KEY")!
    ).auth.getUser(token);

    if (userError || !user) {
      return new Response(
        JSON.stringify({ error: "Неверный токен авторизации" }),
        { status: 401, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    const { data: profile, error: profileError } = await supabase
      .from("profiles")
      .select("credits")
      .eq("id", user.id)
      .single();

    if (profileError || !profile) {
      return new Response(
        JSON.stringify({ error: "Профиль не найден" }),
        { status: 404, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    if (profile.credits <= 0) {
      return new Response(
        JSON.stringify({ error: "Недостаточно кредитов. Пожалуйста, пополните баланс." }),
        { status: 403, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    const OPENROUTER_API_KEY = Deno.env.get("OPENROUTER_API_KEY");
    if (!OPENROUTER_API_KEY) {
      throw new Error("OPENROUTER_API_KEY is not configured");
    }

    const response = await fetch(
      "https://openrouter.ai/api/v1/chat/completions",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${OPENROUTER_API_KEY}`,
          "Content-Type": "application/json",
          "HTTP-Referer": "https://istok-ai.lovable.app",
          "X-Title": "Istok AI",
        },
        body: JSON.stringify({
          model: "anthropic/claude-3.5-sonnet",
          messages: [
            { role: "system", content: SYSTEM_PROMPT },
            ...messages,
          ],
        }),
      }
    );

    if (!response.ok) {
      if (response.status === 429) {
        return new Response(
          JSON.stringify({ error: "Слишком много запросов. Попробуйте позже." }),
          { status: 429, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }
      if (response.status === 402) {
        return new Response(
          JSON.stringify({ error: "Необходимо пополнить баланс." }),
          { status: 402, headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }
      const text = await response.text();
      console.error("OpenRouter error:", response.status, text);
      return new Response(
        JSON.stringify({ error: "Ошибка генерации кода" }),
        { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
      );
    }

    const data = await response.json();
    const content = data.choices?.[0]?.message?.content || "";
    const usage = data.usage;

    const tokensUsed = usage
      ? (usage.prompt_tokens || 0) + (usage.completion_tokens || 0)
      : Math.ceil(content.length / 4) + Math.ceil(JSON.stringify(messages).length / 4);

    const newCredits = Math.max(0, profile.credits - tokensUsed);
    await supabase
      .from("profiles")
      .update({ credits: newCredits })
      .eq("id", user.id);

    // Extract JSON files structure
    const jsonMatch = content.match(/```json\s*\n([\s\S]*?)```/);
    if (jsonMatch) {
      try {
        const files = JSON.parse(jsonMatch[1].trim());
        if (typeof files === "object" && files !== null && !Array.isArray(files)) {
          return new Response(
            JSON.stringify({ files, tokensUsed, creditsRemaining: newCredits }),
            { headers: { ...corsHeaders, "Content-Type": "application/json" } }
          );
        }
      } catch (e) {
        console.error("JSON parse error:", e, "Raw:", jsonMatch[1].substring(0, 200));
      }
    }

    try {
      const files = JSON.parse(content.trim());
      if (typeof files === "object" && files !== null && !Array.isArray(files)) {
        return new Response(
          JSON.stringify({ files, tokensUsed, creditsRemaining: newCredits }),
          { headers: { ...corsHeaders, "Content-Type": "application/json" } }
        );
      }
    } catch { /* not pure JSON */ }

    const htmlMatch = content.match(/```html\s*\n([\s\S]*?)```/);
    const generatedCode = htmlMatch ? htmlMatch[1].trim() : content;

    return new Response(
      JSON.stringify({ files: { "index.html": generatedCode }, tokensUsed, creditsRemaining: newCredits }),
      { headers: { ...corsHeaders, "Content-Type": "application/json" } }
    );
  } catch (e) {
    console.error("generate-code error:", e);
    return new Response(
      JSON.stringify({ error: e instanceof Error ? e.message : "Unknown error" }),
      { status: 500, headers: { ...corsHeaders, "Content-Type": "application/json" } }
    );
  }
});
