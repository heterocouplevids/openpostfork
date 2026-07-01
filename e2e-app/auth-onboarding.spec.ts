import { expect, test } from "@playwright/test";
import {
  createWorkspace,
  password,
  registerUser,
  routeBrowserRegistration,
} from "./helpers";

test("registration routes first-time users through onboarding", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `auth-onboarding-${unique}@example.com`;
  const workspaceName = "Launch Workspace E2E";

  await routeBrowserRegistration(page, email);
  await page.goto("/register");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByLabel("Confirm Password").fill(password);
  await page.getByRole("button", { name: "Create Account" }).click();

  await expect(page).toHaveURL(/\/onboarding$/);
  await expect(
    page.getByRole("heading", { name: "Welcome to OpenPost" }),
  ).toBeVisible();
  await page.getByLabel("Workspace name").fill(workspaceName);
  await page.getByRole("button", { name: "Get Started" }).click();

  await expect(page).toHaveURL(/\/$/);
  await expect(page.locator("textarea").first()).toBeVisible();

  const token = await page.evaluate(() => window.localStorage.getItem("token"));
  if (!token) throw new Error("missing auth token after onboarding");

  const workspaces = await request.get("/api/v1/workspaces", {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(workspaces.ok()).toBeTruthy();
  const workspaceBody = await workspaces.json();
  expect(workspaceBody).toEqual(
    expect.arrayContaining([expect.objectContaining({ name: workspaceName })]),
  );
});

test("login honors same-origin redirects for existing workspaces", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `auth-login-${unique}@example.com`;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, "Login Redirect E2E");

  await page.goto(`/login?redirect=${encodeURIComponent("/settings#billing")}`);
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Sign In" }).click();

  await expect(page).toHaveURL(/\/settings#billing$/);
  await expect(page.getByRole("heading", { name: "Billing" })).toBeVisible();
});
