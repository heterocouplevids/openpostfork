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
