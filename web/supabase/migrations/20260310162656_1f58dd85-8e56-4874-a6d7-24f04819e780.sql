
-- Fix search_path for generate_project_slug
CREATE OR REPLACE FUNCTION public.generate_project_slug()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY INVOKER
SET search_path = public
AS $$
BEGIN
  IF NEW.slug IS NULL THEN
    NEW.slug := substr(NEW.id::text, 1, 8);
  END IF;
  RETURN NEW;
END;
$$;
