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

  await expect(page.getByTestId("settings-tabs")).toBeVisible();
  await page.getByRole("tab", { name: "Organization" }).click();
  await expect(page.getByRole("heading", { name: "Billing" })).toBeVisible();
  await expect(page.getByText("No active plan")).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Customer Portal" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Start Checkout" }),
  ).toHaveCount(5);
  await expect(page.getByRole("heading", { name: "Starter" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Creator" })).toBeVisible();
  await expect(page.getByRole("heading", { name: "Pro" })).toBeVisible();
  await expect(
    page.locator("#billing").getByRole("heading", { name: "Team" }),
  ).toBeVisible();
  await expect(
    page.locator("#billing").getByRole("heading", { name: "Agency" }),
  ).toBeVisible();
});

test("settings lets instance admins manage provider apps", async ({ page }) => {
  let providerApps: Array<{
    id: string;
    provider: string;
    name?: string;
    client_id: string;
    redirect_uri?: string;
    instance_url?: string;
    is_active: boolean;
    secret_configured: boolean;
    created_at: string;
    updated_at: string;
  }> = [];

  await page.addInitScript(() => {
    window.localStorage.setItem("token", "admin-settings-token");
  });
  await page.route("**/api/v1/auth/me", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: {
        id: "admin-1",
        email: "admin@example.com",
        display_name: "Admin User",
        avatar_url: "",
        is_admin: true,
        created_at: "2026-07-01T00:00:00Z",
      },
    });
  });
  await page.route("**/api/v1/workspaces**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: [
        {
          id: "ws-1",
          organization_id: "org-1",
          organization_name: "Admin Org",
          name: "Admin Settings",
          created_at: "2026-07-01T00:00:00Z",
        },
      ],
    });
  });
  await page.route("**/api/v1/organizations/org-1/billing/status", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: {
        organization_id: "org-1",
        workspace_id: "ws-1",
        status: "inactive",
        cancel_at_period_end: false,
        limits: {},
        usage: {},
        period_start: "2026-07-01T00:00:00Z",
      },
    });
  });
  await page.route("**/api/v1/workspaces/ws-1/team", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: { members: [], invitations: [], current_seats: 0 },
    });
  });
  await page.route("**/api/v1/posting-schedules?**", async (route) => {
    await route.fulfill({ contentType: "application/json", json: [] });
  });
  await page.route("**/api/v1/auth/security", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: {
        user: {
          id: "admin-1",
          email: "admin@example.com",
          display_name: "Admin User",
          avatar_url: "",
          created_at: "2026-07-01T00:00:00Z",
        },
        totp_enabled: false,
        passkeys: [],
        methods: [],
      },
    });
  });
  await page.route("**/api/v1/auth/sessions", async (route) => {
    await route.fulfill({ contentType: "application/json", json: [] });
  });
  await page.route("**/api/v1/api-tokens", async (route) => {
    await route.fulfill({ contentType: "application/json", json: [] });
  });
  await page.route("**/api/v1/mcp/activity?**", async (route) => {
    await route.fulfill({ contentType: "application/json", json: [] });
  });
  await page.route("**/api/v1/admin/provider-apps", async (route) => {
    if (route.request().method() === "GET") {
      await route.fulfill({
        contentType: "application/json",
        json: providerApps,
      });
      return;
    }
    const body = route.request().postDataJSON() as {
      provider: string;
      name?: string;
      client_id: string;
      client_secret?: string;
      redirect_uri?: string;
      instance_url?: string;
      is_active?: boolean;
    };
    const existing = providerApps.find(
      (app) =>
        app.provider === body.provider &&
        (app.instance_url ?? "") === (body.instance_url ?? ""),
    );
    const saved = {
      id: existing?.id ?? "provider-app-1",
      provider: body.provider,
      name: body.name,
      client_id: body.client_id,
      redirect_uri: body.redirect_uri,
      instance_url: body.instance_url,
      is_active: body.is_active ?? true,
      secret_configured: Boolean(
        body.client_secret || existing?.secret_configured,
      ),
      created_at: existing?.created_at ?? "2026-07-01T00:00:00Z",
      updated_at: "2026-07-01T00:00:00Z",
    };
    providerApps = existing
      ? providerApps.map((app) => (app.id === existing.id ? saved : app))
      : [...providerApps, saved];
    await route.fulfill({
      contentType: "application/json",
      json: { app: saved, existed: Boolean(existing), requires_restart: true },
    });
  });
  await page.route("**/api/v1/admin/provider-apps/*", async (route) => {
    const id = route.request().url().split("/").pop();
    providerApps = providerApps.filter((app) => app.id !== id);
    await route.fulfill({
      contentType: "application/json",
      json: { requires_restart: true },
    });
  });

  await page.goto("/settings");

  await page.getByRole("tab", { name: "Admin" }).click();
  await expect(page.getByTestId("provider-apps-settings")).toBeVisible();

  await page.locator("#provider-app-name").fill("Production X");
  await page.getByTestId("provider-app-client-id").fill("x-client-id");
  await page.getByTestId("provider-app-client-secret").fill("x-client-secret");
  await page.getByTestId("provider-app-save").click();

  await expect(page.getByTestId("provider-app-restart-required")).toBeVisible();
  await expect(page.getByTestId("provider-app-list")).toContainText(
    "X / Twitter",
  );
  await expect(page.getByTestId("provider-app-list")).toContainText(
    "Production X",
  );
  await expect(page.getByTestId("provider-app-list")).toContainText("stored");

  const providerAppRow = page.getByTestId("provider-app-row");
  await providerAppRow.getByRole("button", { name: "Edit" }).click();
  await expect(page.getByTestId("provider-app-client-id")).toHaveValue(
    "x-client-id",
  );
  await page.getByTestId("provider-app-client-id").fill("x-client-id-rotated");
  await page.getByTestId("provider-app-save").click();
  await expect(page.getByTestId("provider-app-list")).toContainText(
    "x-client-id-rotated",
  );
  await expect(page.getByTestId("provider-app-list")).toContainText("stored");

  page.on("dialog", (dialog) => dialog.accept());
  await providerAppRow.getByRole("button", { name: "Delete" }).click();
  await expect(
    page.getByText("No database-backed provider apps are configured yet"),
  ).toBeVisible();
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
  await page.getByRole("tab", { name: "Account" }).click();

  await expect(
    page.getByRole("heading", { name: "Recent MCP Activity" }),
  ).toBeVisible();
  await expect(page.getByTestId("mcp-activity-list")).toContainText(
    "list_workspaces",
  );
  await expect(page.getByTestId("mcp-activity-list")).toContainText("success");
});

test("settings account tab updates the user profile", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `profile-${unique}@example.com`;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "Profile E2E");

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.goto("/settings?tab=account");

  await expect(page.getByRole("heading", { name: "Profile" })).toBeVisible();
  await page.getByLabel("Display name").fill("Profile E2E User");
  await page.getByRole("button", { name: "Save Profile" }).click();
  await expect(page.getByText("Profile updated")).toBeVisible();

  const me = await request.get("/api/v1/auth/me", {
    headers: { Authorization: `Bearer ${auth.token}` },
  });
  expect(me.ok()).toBeTruthy();
  const meBody = await me.json();
  expect(meBody.display_name).toBe("Profile E2E User");
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
  await page.getByRole("tab", { name: "Account" }).click();

  await expect(
    page.getByRole("heading", { name: "Active Sessions" }),
  ).toBeVisible();
  await expect(page.getByTestId("auth-session-list")).toContainText("Current");
  await expect(page.getByTestId("auth-session-list")).toContainText(
    "Browser on device",
  );
  await expect(page.getByTestId("auth-session-list")).not.toContainText(
    "E2E Other Browser",
  );

  page.on("dialog", (dialog) => dialog.accept());
  const otherSession = page
    .getByTestId("auth-session-row")
    .filter({ hasText: "Browser on device" });
  await otherSession.getByRole("button", { name: "Revoke" }).click();
  await expect(page.getByTestId("auth-session-list")).not.toContainText(
    "Browser on device",
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
  await page.getByRole("tab", { name: "Account" }).click();

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
  await page.getByRole("tab", { name: "Organization" }).click();

  await expect(
    page.locator("#team").getByRole("heading", { name: "Team" }),
  ).toBeVisible();

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
  await expect(invitedPage).toHaveURL(/\/settings\?tab=organization$/);
  await expect(
    invitedPage.locator("#team").getByRole("heading", { name: "Team" }),
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
  let checkoutURL = "";

  await routeBrowserRegistration(page, email);
  await page.route("**/api/v1/**/billing/checkout", async (route) => {
    checkoutURL = route.request().url();
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
  expect(checkoutURL).toContain("/organizations/");
  expect(checkoutBody?.plan_id).toBe("creator");
});
