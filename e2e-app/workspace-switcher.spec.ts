import { expect, test } from "@playwright/test";
import { createWorkspace, registerUser } from "./helpers";

test("sidebar footer switches between workspaces", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `workspace-switcher-${unique}@example.com`;
  const firstName = `Launch ${unique}`;
  const secondName = `Client ${unique}`;

  const auth = await registerUser(request, email);
  await createWorkspace(request, auth.token, firstName);
  await createWorkspace(request, auth.token, secondName);

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.goto("/");

  const workspaceNames = [firstName, secondName];
  const workspaceNamePattern = new RegExp(workspaceNames.join("|"));
  const workspaceButton = page
    .getByRole("button", { name: workspaceNamePattern })
    .first();
  await expect(workspaceButton).toBeVisible();
  const buttonText = await workspaceButton.innerText();
  const activeWorkspace = workspaceNames.find((name) =>
    buttonText.includes(name),
  );
  expect(activeWorkspace).toBeTruthy();
  const nextWorkspace =
    activeWorkspace === firstName ? secondName : firstName;

  await workspaceButton.click();
  await expect(page.getByText("Switch workspace")).toBeVisible();
  await page.getByRole("menuitem", { name: new RegExp(nextWorkspace) }).click();

  await expect(workspaceButton).toContainText(nextWorkspace);
});
