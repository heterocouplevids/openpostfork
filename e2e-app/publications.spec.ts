import { expect, test } from "@playwright/test";

test("publications page sends a source publication into the composer", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `publications-${unique}@example.com`;
  const password = "password-1234";

  const register = await request.post("/api/v1/auth/register", {
    data: { email, password },
  });
  expect(register.ok()).toBeTruthy();
  const auth = await register.json();
  expect(auth.token).toBeTruthy();

  const workspace = await request.post("/api/v1/workspaces", {
    headers: { Authorization: `Bearer ${auth.token}` },
    data: { name: "Publications E2E" },
  });
  expect(workspace.ok()).toBeTruthy();
  const workspaceBody = await workspace.json();

  const sourceContent =
    "Launch the agentic scheduler flow with CLI, MCP, and app handoff.";
  const publication = await request.post("/api/v1/publications", {
    headers: { Authorization: `Bearer ${auth.token}` },
    data: {
      workspace_id: workspaceBody.id,
      title: "Agent launch brief",
      source_content: sourceContent,
      source_url: "https://example.com/launch",
      goal: "launch",
      audience: "operators",
    },
  });
  expect(publication.ok()).toBeTruthy();
  const publicationBody = await publication.json();

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.goto("/publications");

  await expect(
    page.getByRole("heading", { name: "Publications" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: /Agent launch brief/ }),
  ).toBeVisible();
  await expect(
    page
      .getByRole("complementary")
      .getByRole("heading", { name: "Agent launch brief" }),
  ).toBeVisible();
  await expect(
    page.getByRole("complementary").getByText(sourceContent),
  ).toBeVisible();

  await page.getByRole("button", { name: "Compose" }).click();
  await expect(page).toHaveURL(new RegExp(`publication=${publicationBody.id}`));
  await expect(page.getByText("Source publication")).toBeVisible();
  await expect(page.getByText("Agent launch brief")).toBeVisible();
  await expect(page.locator("textarea").first()).toHaveValue(sourceContent);
});
