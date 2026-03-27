import { useEffect, useState, useMemo } from "react";
import { useParams } from "react-router-dom";
// import { supabase } from "@/integrations/supabase/client"; // Не используется - переход на Go Auth
import { Loader2 } from "lucide-react";
import { codeToFiles } from "@/components/WorkspacePreview";

function buildFullHtml(code: string): string {
  const files = codeToFiles(code);
  const html = files["index.html"] || "";
  if (Object.keys(files).length === 1) return html;

  let result = html;
  for (const [name, content] of Object.entries(files)) {
    if (name.endsWith(".css")) {
      const re = new RegExp(`<link[^>]*href=["']${name.replace(".", "\\.")}["'][^>]*/?>`, "gi");
      result = re.test(result)
        ? result.replace(re, `<style>${content}</style>`)
        : result.replace("</head>", `<style>${content}</style>\n</head>`);
    }
  }
  for (const [name, content] of Object.entries(files)) {
    if (name.endsWith(".js") || name.endsWith(".ts")) {
      const re = new RegExp(`<script[^>]*src=["']${name.replace(".", "\\.")}["'][^>]*>\\s*</script>`, "gi");
      result = re.test(result)
        ? result.replace(re, `<script>${content}</script>`)
        : result.replace("</body>", `<script>${content}</script>\n</body>`);
    }
  }
  return result;
}

const ViewProject = () => {
  const { projectId } = useParams<{ projectId: string }>();
  const [code, setCode] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!projectId) return;
    (async () => {
      let { data, error: err } = await supabase
        .from("projects")
        .select("code, is_public")
        .eq("slug", projectId)
        .eq("is_public", true)
        .maybeSingle();
      if (!data) {
        const res = await supabase
          .from("projects")
          .select("code, is_public")
          .eq("id", projectId)
          .eq("is_public", true)
          .maybeSingle();
        data = res.data;
        err = res.error;
      }
      if (err || !data) {
        setError("Проект не найден или не опубликован");
      } else {
        setCode(data.code);
      }
      setLoading(false);
    })();
  }, [projectId]);

  const previewHtml = useMemo(() => (code ? buildFullHtml(code) : ""), [code]);

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-background">
        <Loader2 size={28} className="text-primary animate-spin" />
      </div>
    );
  }

  if (error || !code) {
    return (
      <div className="h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <p className="text-muted-foreground text-lg">{error || "Проект не найден"}</p>
          <a href="/" className="text-primary text-sm mt-4 inline-block hover:underline">← На главную</a>
        </div>
      </div>
    );
  }

  return (
    <iframe title="Published project" srcDoc={previewHtml} className="w-full h-screen border-0" sandbox="allow-scripts" />
  );
};

export default ViewProject;
