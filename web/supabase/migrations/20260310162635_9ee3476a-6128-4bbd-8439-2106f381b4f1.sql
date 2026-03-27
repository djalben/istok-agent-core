
-- Add public hosting fields to projects
ALTER TABLE public.projects
  ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN slug TEXT UNIQUE DEFAULT NULL;

-- Generate slug from id for convenience
CREATE OR REPLACE FUNCTION public.generate_project_slug()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  IF NEW.slug IS NULL THEN
    NEW.slug := substr(NEW.id::text, 1, 8);
  END IF;
  RETURN NEW;
END;
$$;

CREATE TRIGGER set_project_slug
  BEFORE INSERT ON public.projects
  FOR EACH ROW EXECUTE FUNCTION public.generate_project_slug();

-- Update existing rows to have slugs
UPDATE public.projects SET slug = substr(id::text, 1, 8) WHERE slug IS NULL;

-- Allow anyone to read public projects
CREATE POLICY "Anyone can read public projects"
  ON public.projects FOR SELECT
  TO anon, authenticated
  USING (is_public = true);
