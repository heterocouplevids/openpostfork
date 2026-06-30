import { expect, test } from "@playwright/test";

test("accounts page shows configured and unavailable providers", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `accounts-${unique}@example.com`;
  const password = "password-1234";

  const register = await request.post("/api/v1/auth/register", {
    data: { email, password },
  });
  expect(register.ok()).toBeTruthy();
  const auth = await register.json();
  expect(auth.token).toBeTruthy();

  const workspace = await request.post("/api/v1/workspaces", {
    headers: { Authorization: `Bearer ${auth.token}` },
    data: { name: "Provider Availability E2E" },
  });
  expect(workspace.ok()).toBeTruthy();

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.route("**/api/v1/accounts/providers", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: [
        {
          platform: "bluesky",
          display_name: "Bluesky",
          auth_mode: "app_password",
          configured: true,
        },
        {
          platform: "x",
          display_name: "X (Twitter)",
          auth_mode: "oauth",
          configured: false,
        },
        {
          platform: "mastodon",
          display_name: "Mastodon",
          auth_mode: "oauth_oob",
          configured: false,
        },
        {
          platform: "linkedin",
          display_name: "LinkedIn",
          auth_mode: "oauth",
          configured: false,
        },
        {
          platform: "threads",
          display_name: "Threads",
          auth_mode: "oauth",
          configured: false,
        },
      ],
    });
  });
  await page.goto("/accounts");

  await expect(
    page.getByRole("heading", { name: "Connect a Platform" }),
  ).toBeVisible();
  await expect(page.getByTestId("provider-card-bluesky")).toContainText(
    "Post to Bluesky",
  );
  await expect(
    page
      .getByTestId("provider-card-bluesky")
      .getByRole("button", { name: "Connect" }),
  ).toBeEnabled();

  for (const platform of ["x", "mastodon", "linkedin", "threads"]) {
    await expect(page.getByTestId(`provider-card-${platform}`)).toContainText(
      "Not configured",
    );
    await expect(
      page
        .getByTestId(`provider-card-${platform}`)
        .getByRole("button", { name: "Unavailable" }),
    ).toBeDisabled();
  }
});
