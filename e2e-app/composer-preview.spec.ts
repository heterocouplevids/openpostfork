import { expect, test } from "@playwright/test";
import { createWorkspace, registerUser } from "./helpers";

test("composer renders account-specific provider previews", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `composer-preview-${unique}@example.com`;

  const auth = await registerUser(request, email);
  const workspaceBody = await createWorkspace(
    request,
    auth.token,
    "Composer Preview E2E",
  );

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.route("**/api/v1/accounts?**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: [
        {
          id: "instagram-main",
          slug: "instagram-main",
          platform: "instagram",
          account_id: "ig-main",
          account_username: "openpost_main",
          account_avatar_url: "https://cdn.example/main.jpg",
          instance_url: "",
          is_active: true,
          thread_replies_supported: false,
        },
        {
          id: "instagram-studio",
          slug: "instagram-studio",
          platform: "instagram",
          account_id: "ig-studio",
          account_username: "openpost_studio",
          account_avatar_url: "https://cdn.example/studio.jpg",
          instance_url: "",
          is_active: true,
          thread_replies_supported: false,
        },
      ],
    });
  });
  await page.route("**/api/v1/posts", async (route) => {
    if (route.request().method() === "POST") {
      await route.fulfill({
        contentType: "application/json",
        json: {
          id: "draft-preview",
          workspace_id: workspaceBody.id,
          content: "Launch update",
          status: "draft",
          scheduled_at: "",
          media: [],
          destinations: [],
        },
      });
      return;
    }
    await route.continue();
  });

  await page.goto("/");
  await page.locator("textarea").first().fill("Launch update");

  await expect(page.locator('[data-testid="instagram-preview"]')).toHaveCount(
    2,
  );
  await expect(
    page.locator('[data-testid="instagram-preview"]').nth(0),
  ).toContainText("@openpost_main");
  await expect(
    page.locator('[data-testid="instagram-preview"]').nth(1),
  ).toContainText("@openpost_studio");
});
