import { test, expect } from "@playwright/test";

/**
 * E2E Audit: Toolbar Visibility & Displacement
 * 
 * Bug: Toolbar goes off-screen after ~0.5s (Framer Motion animation push).
 * This test captures screenshots at 0.5s and 2s, then verifies computed styles.
 */

test.describe("Toolbar Audit", () => {
  test.beforeEach(async ({ page }) => {
    // Mock auth API so ProtectedRoute allows access (intercepts before Vite proxy)
    await page.route("**/api/v1/auth/me", (route) =>
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          id: "e2e-user",
          email: "test@istok.dev",
          display_name: "E2E Tester",
          created_at: new Date().toISOString(),
        }),
      })
    );
    await page.route("**/api/v1/health", (route) =>
      route.fulfill({ status: 200, contentType: "application/json", body: '{"status":"healthy"}' })
    );
    // Inject auth tokens BEFORE React hydrates (runs before any JS on page)
    await page.addInitScript(() => {
      localStorage.setItem("auth_token", "fake-e2e-token");
      localStorage.setItem("istok_token", "fake-e2e-token");
      localStorage.setItem("istok_user", JSON.stringify({
        id: "e2e-user",
        email: "test@istok.dev",
        display_name: "E2E Tester",
        created_at: new Date().toISOString(),
      }));
    });
  });

  test("toolbar stays visible and does not displace off-screen", async ({ page }) => {
    await page.goto("/project/new", { waitUntil: "domcontentloaded" });

    // Screenshot 1: immediate (catches initial render before animation kicks)
    await page.waitForTimeout(500);
    await page.screenshot({ path: "e2e/screenshots/toolbar-0.5s.png", fullPage: false });

    // Screenshot 2: after animations settle
    await page.waitForTimeout(1500);
    await page.screenshot({ path: "e2e/screenshots/toolbar-2s.png", fullPage: false });

    // Find the toolbar header element
    const toolbar = page.locator("header").first();
    await expect(toolbar).toBeVisible();

    // Check computed styles
    const styles = await toolbar.evaluate((el) => {
      const computed = window.getComputedStyle(el);
      const rect = el.getBoundingClientRect();
      return {
        position: computed.position,
        zIndex: computed.zIndex,
        top: rect.top,
        height: rect.height,
        transform: computed.transform,
        overflow: computed.overflow,
      };
    });

    console.log("🔍 TOOLBAR COMPUTED STYLES:", JSON.stringify(styles, null, 2));

    // Assertions: toolbar must be fixed at viewport top
    expect(styles.position).toBe("fixed");
    expect(styles.top).toBeGreaterThanOrEqual(0); // At viewport top
    expect(styles.top).toBeLessThan(2); // Must be at top: 0
    expect(styles.height).toBeGreaterThan(30); // Must have visible height
    expect(styles.zIndex).not.toBe("auto"); // Must have explicit z-index

    // Check that toolbar's top is not pushed by Framer Motion transform
    if (styles.transform && styles.transform !== "none") {
      console.warn("⚠️ TOOLBAR HAS TRANSFORM:", styles.transform, "— possible Framer Motion interference");
    }
  });

  test("preview container fills remaining space below toolbar", async ({ page }) => {
    await page.goto("/project/new", { waitUntil: "domcontentloaded" });
    await page.waitForTimeout(2000);

    const toolbar = page.locator("header").first();
    const contentArea = page.locator("header + div").first();

    const toolbarBox = await toolbar.boundingBox();
    const contentBox = await contentArea.boundingBox();

    if (toolbarBox && contentBox) {
      console.log("📐 LAYOUT:", {
        toolbar: { top: toolbarBox.y, height: toolbarBox.height, bottom: toolbarBox.y + toolbarBox.height },
        content: { top: contentBox.y, height: contentBox.height },
      });

      // With position:fixed toolbar, content has pt-14 (56px) to compensate
      // Content's visible area should start around toolbar bottom
      expect(toolbarBox.y).toBeLessThan(2); // Toolbar pinned at top: 0
      expect(toolbarBox.height).toBeGreaterThanOrEqual(50); // 56px height
      expect(contentBox.height).toBeGreaterThan(100); // Content must fill space
    }
  });
});
