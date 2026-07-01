import { expect, test } from "@playwright/test";
import {
  createWorkspace,
  password,
  registerUser,
  routeBrowserRegistration,
} from "./helpers";

test("settings shows billing plan controls for an authenticated workspace", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `billing-${unique}@example.com`;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "Billing E2E");

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.goto("/settings");

  await expect(page.getByTestId("settings-section-nav")).toBeVisible();
  await expect(
    page
      .getByTestId("settings-section-nav")
      .getByRole("link", { name: "Billing" }),
  ).toHaveAttribute("href", "#billing");
  await expect(page.getByRole("heading", { name: "Billing" })).toBeVisible();
  await expect(page.getByText("No active plan")).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Customer Portal" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Start Checkout" }),
  ).toHaveCount(3);
  await expect(page.getByRole("heading", { name: "Starter" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Creator" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Pro" })).toBeVisible();
});

test("settings shows recent MCP activity for an authenticated user", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `mcp-activity-${unique}@example.com`;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "MCP Activity E2E");

  const mcpCall = await request.post("/mcp", {
    headers: { Authorization: `Bearer ${auth.token}` },
    data: {
      jsonrpc: "2.0",
      id: 1,
      method: "tools/call",
      params: {
        name: "list_workspaces",
        arguments: {},
      },
    },
  });
  expect(mcpCall.ok()).toBeTruthy();
  const mcpBody = await mcpCall.json();
  expect(mcpBody.error).toBeFalsy();

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.goto("/settings");

  await expect(
    page.getByRole("heading", { name: "Recent MCP Activity" }),
  ).toBeVisible();
  await expect(page.getByTestId("mcp-activity-list")).toContainText(
    "list_workspaces",
  );
  await expect(page.getByTestId("mcp-activity-list")).toContainText("success");
});

test("settings lists and revokes active web sessions", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `sessions-${unique}@example.com`;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "Sessions E2E");

  const secondLogin = await request.post("/api/v1/auth/login", {
    headers: { "User-Agent": "E2E Other Browser" },
    data: { email, password },
  });
  expect(secondLogin.ok()).toBeTruthy();

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.goto("/settings");

  await expect(
    page.getByRole("heading", { name: "Active Sessions" }),
  ).toBeVisible();
  await expect(page.getByTestId("auth-session-list")).toContainText("Current");
  await expect(page.getByTestId("auth-session-list")).toContainText(
    "E2E Other Browser",
  );

  page.on("dialog", (dialog) => dialog.accept());
  const otherSession = page
    .getByTestId("auth-session-row")
    .filter({ hasText: "E2E Other Browser" });
  await otherSession.getByRole("button", { name: "Revoke" }).click();
  await expect(page.getByTestId("auth-session-list")).not.toContainText(
    "E2E Other Browser",
  );
});

test("settings creates MCP-scoped API tokens", async ({ page, request }) => {
  const unique = Date.now().toString(36);
  const email = `mcp-token-${unique}@example.com`;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "MCP Token E2E");

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.goto("/settings");

  await expect(page.getByTestId("api-token-scope")).toContainText(
    "MCP / ChatGPT App",
  );
  await page.locator("#api-token-name").fill("ChatGPT App E2E");
  await page.getByRole("button", { name: "Create Token" }).click();

  await expect(page.getByText("Copy this token now")).toBeVisible();
  await expect(page.getByText(/op_cli_[a-f0-9]{8}_/)).toBeVisible();
  await expect(page.getByText("ChatGPT App E2E")).toBeVisible();
  await expect(page.getByText(/mcp:full/)).toBeVisible();
});

test("settings creates and accepts workspace invitations", async ({
  browser,
  baseURL,
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const adminEmail = `team-admin-${unique}@example.com`;
  const inviteEmail = `team-member-${unique}@example.com`;

  const adminAuth = await registerUser(request, adminEmail);
  await createWorkspace(request, adminAuth.token, "Team E2E");

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, adminAuth.token);
  await page.goto("/settings");

  await expect(
    page
      .getByTestId("settings-section-nav")
      .getByRole("link", { name: "Team" }),
  ).toHaveAttribute("href", "#team");
  await expect(page.getByRole("heading", { name: "Team" })).toBeVisible();

  await page.getByTestId("team-invite-email").fill(inviteEmail);
  await page.getByRole("button", { name: "Send Invite" }).click();

  await expect(page.getByTestId("team-invite-link")).toContainText(
    "/invite?token=op_inv_",
  );
  await expect(page.getByTestId("team-invitations-list")).toContainText(
    inviteEmail,
  );

  const inviteLinkText = (await page
    .getByTestId("team-invite-link")
    .textContent())!;
  const inviteURL = inviteLinkText.match(
    /https?:\/\/\S+\/invite\?token=\S+/,
  )?.[0];
  expect(inviteURL).toBeTruthy();

  const invitedAuth = await registerUser(request, inviteEmail);
  const invitedContext = await browser.newContext({ baseURL });
  const invitedPage = await invitedContext.newPage();
  await invitedPage.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, invitedAuth.token);

  const parsedInviteURL = new URL(inviteURL!);
  await invitedPage.goto(
    `${parsedInviteURL.pathname}${parsedInviteURL.search}`,
  );

  await expect(
    invitedPage.getByRole("heading", { name: "Invitation accepted" }),
  ).toBeVisible();
  await expect(invitedPage.getByText("editor access")).toBeVisible();

  await invitedPage.getByRole("button", { name: "Open Settings" }).click();
  await expect(invitedPage).toHaveURL(/\/settings#team$/);
  await expect(
    invitedPage.getByRole("heading", { name: "Team" }),
  ).toBeVisible();
  await expect(invitedPage.getByTestId("team-members-list")).toContainText(
    inviteEmail,
  );
  await invitedContext.close();
});

test("plan selection from signup starts checkout after onboarding", async ({
  page,
}) => {
  const unique = Date.now().toString(36);
  const email = `plan-signup-${unique}@example.com`;
  let checkoutBody: { workspace_id?: string; plan_id?: string } | undefined;

  await routeBrowserRegistration(page, email);
  await page.route("**/api/v1/billing/checkout", async (route) => {
    checkoutBody = JSON.parse(route.request().postData() ?? "{}");
    await route.fulfill({
      contentType: "application/json",
      json: { id: "checkout-e2e", url: "/settings?checkout=creator" },
    });
  });

  await page.goto("/register?plan=creator");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByLabel("Confirm Password").fill(password);
  await page.getByRole("button", { name: "Create Account" }).click();

  await expect(page).toHaveURL(/\/onboarding\?plan=creator/);
  await page.getByLabel("Workspace name").fill("Plan Handoff E2E");
  await page.getByRole("button", { name: "Get Started" }).click();

  await expect(page).toHaveURL(/\/settings\?checkout=creator/);
  expect(checkoutBody?.workspace_id).toBeTruthy();
  expect(checkoutBody?.plan_id).toBe("creator");
});
