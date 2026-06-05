package application

import (
	"context"
	"regexp"
	"strings"

	"github.com/djalben/istok-agent-core/internal/ports"
)

const inspectorProviderPath = "src/components/InspectorProvider.tsx"

// inspectorProviderCode is injected into every generated React project.
// Provides visual element inspection and postMessage bridge for Point-and-Click editing.
const inspectorProviderCode = `import React, { useCallback, useEffect, useRef, useState, createContext, useContext } from 'react';

interface ElementData {
  tagName: string;
  className: string;
  id: string;
  componentName: string | null;
  textContent: string;
  rect: { top: number; left: number; width: number; height: number };
}

interface InspectorContextType {
  inspectMode: boolean;
  setInspectMode: (v: boolean) => void;
}

const InspectorContext = createContext<InspectorContextType>({
  inspectMode: false,
  setInspectMode: () => {},
});

export const useInspector = () => useContext(InspectorContext);

export const InspectorProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [inspectMode, setInspectMode] = useState(false);
  const overlayRef = useRef<HTMLDivElement | null>(null);
  const lastTarget = useRef<HTMLElement | null>(null);

  // Listen for parent frame toggling inspect mode
  useEffect(() => {
    const handler = (e: MessageEvent) => {
      if (e.data?.type === 'ISTOK_TOGGLE_INSPECT') {
        setInspectMode((prev) => !prev);
      }
      if (e.data?.type === 'ISTOK_SET_INSPECT') {
        setInspectMode(!!e.data.enabled);
      }
    };
    window.addEventListener('message', handler);
    return () => window.removeEventListener('message', handler);
  }, []);

  // Alt-key hold activates inspect mode temporarily
  useEffect(() => {
    const down = (e: KeyboardEvent) => { if (e.altKey) setInspectMode(true); };
    const up = (e: KeyboardEvent) => { if (!e.altKey) setInspectMode(false); };
    window.addEventListener('keydown', down);
    window.addEventListener('keyup', up);
    return () => {
      window.removeEventListener('keydown', down);
      window.removeEventListener('keyup', up);
    };
  }, []);

  const getElementData = useCallback((el: HTMLElement): ElementData => {
    const rect = el.getBoundingClientRect();
    return {
      tagName: el.tagName.toLowerCase(),
      className: el.className || '',
      id: el.id || '',
      componentName: el.getAttribute('data-component-name') ||
        el.closest('[data-component-name]')?.getAttribute('data-component-name') || null,
      textContent: (el.textContent || '').slice(0, 120),
      rect: { top: rect.top, left: rect.left, width: rect.width, height: rect.height },
    };
  }, []);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!inspectMode) return;
    const target = e.target as HTMLElement;
    if (target === lastTarget.current) return;
    lastTarget.current = target;

    if (!overlayRef.current) {
      const div = document.createElement('div');
      div.id = 'istok-inspector-overlay';
      div.style.cssText =
        'position:fixed;pointer-events:none;border:2px solid #3b82f6;' +
        'background:rgba(59,130,246,0.08);z-index:99999;transition:all 0.1s ease;' +
        'border-radius:4px;';
      document.body.appendChild(div);
      overlayRef.current = div;
    }

    const rect = target.getBoundingClientRect();
    const ov = overlayRef.current;
    ov.style.top = rect.top + 'px';
    ov.style.left = rect.left + 'px';
    ov.style.width = rect.width + 'px';
    ov.style.height = rect.height + 'px';
    ov.style.display = 'block';
  }, [inspectMode]);

  const handleMouseLeave = useCallback(() => {
    if (overlayRef.current) overlayRef.current.style.display = 'none';
    lastTarget.current = null;
  }, []);

  const handleClick = useCallback((e: React.MouseEvent) => {
    if (!inspectMode) return;
    e.preventDefault();
    e.stopPropagation();

    const target = e.target as HTMLElement;
    const data = getElementData(target);

    window.parent.postMessage({
      type: 'ISTOK_ELEMENT_CLICKED',
      elementData: data,
    }, '*');
  }, [inspectMode, getElementData]);

  return (
    <InspectorContext.Provider value={{ inspectMode, setInspectMode }}>
      <div
        onMouseMove={handleMouseMove}
        onMouseLeave={handleMouseLeave}
        onClick={handleClick}
        style={{ minHeight: '100vh' }}
      >
        {children}
      </div>
    </InspectorContext.Provider>
  );
};

export default InspectorProvider;
`

// injectInspectorProvider adds the InspectorProvider file to a generated file set
// and patches the main App entry point to wrap content in <InspectorProvider>.
// Only applies to multi-file React projects (not single index.html).
func injectInspectorProvider(ctx context.Context, files map[string]string) {
	if len(files) < 3 {
		return // single-file project, skip injection
	}

	// Check if it's a React project (has .tsx files)
	hasReact := false
	for path := range files {
		if len(path) > 4 && path[len(path)-4:] == ".tsx" {
			hasReact = true

			break
		}
	}
	if !hasReact {
		return
	}

	// 1. Inject the InspectorProvider file
	files[inspectorProviderPath] = inspectorProviderCode
	ports.LoggerFromContext(ctx).InfoContext(ctx, "inspector provider injected", "path", inspectorProviderPath)

	// 2. Mount the provider in the app entry so element clicks are actually intercepted.
	mountInspectorProvider(ctx, files)
}

// appRenderRe matches a self-closing <App /> render call in the entry point.
var appRenderRe = regexp.MustCompile(`<App\s*/>`)

// mountInspectorProvider wraps <App /> in the React entry with <InspectorProvider>
// and adds the import. DEFENSIVE: no-op if no recognizable entry or render shape is
// found, so a valid project's render is never broken (worst case: inspect via Alt only).
func mountInspectorProvider(ctx context.Context, files map[string]string) {
	for _, entry := range []string{"src/main.tsx", "src/index.tsx", "src/main.jsx", "src/index.jsx"} {
		code, ok := files[entry]
		if !ok || code == "" {
			continue
		}
		if strings.Contains(code, "InspectorProvider") {
			return // already wrapped — nothing to do
		}
		if !appRenderRe.MatchString(code) {
			return // unknown render shape — leave untouched (safe)
		}
		patched := appRenderRe.ReplaceAllString(code, "<InspectorProvider><App /></InspectorProvider>")
		patched = "import InspectorProvider from './components/InspectorProvider';\n" + patched
		files[entry] = patched
		ports.LoggerFromContext(ctx).InfoContext(ctx, "inspector provider mounted", "entry", entry)

		return
	}
}
