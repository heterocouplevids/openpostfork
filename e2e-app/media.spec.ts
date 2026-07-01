import { expect, test } from "@playwright/test";
import { createWorkspace, registerUser } from "./helpers";

const tinyPNG = Buffer.from(
  "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
  "base64",
);

test("media library uploads and lists a local media file", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `media-library-${unique}@example.com`;

  const auth = await registerUser(request, email);
  const workspaceBody = await createWorkspace(
    request,
    auth.token,
    "Media Library E2E",
  );

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.goto("/media");

  await expect(
    page.getByRole("heading", { name: "Media Library" }),
  ).toBeVisible();
  await expect(page.getByText("No media found")).toBeVisible();

  await page.getByRole("button", { name: "Upload" }).first().click();
  await expect(
    page.getByRole("dialog", { name: "Upload Media" }),
  ).toBeVisible();
  await page.locator("#file-upload").setInputFiles({
    name: "launch-card.png",
    mimeType: "image/png",
    buffer: tinyPNG,
  });
  await page
    .getByRole("dialog")
    .getByRole("button", { name: "Upload" })
    .click();

  await expect(page.getByText("File uploaded successfully")).toBeVisible();
  await expect(page.getByText("launch-card.png")).toBeVisible();
  await expect(
    page.locator("span").filter({ hasText: /^Unused$/ }),
  ).toBeVisible();
  await expect(page.getByText("1 file")).toBeVisible();

  const media = await request.get(
    `/api/v1/media?workspace_id=${workspaceBody.id}`,
    {
      headers: { Authorization: `Bearer ${auth.token}` },
    },
  );
  expect(media.ok()).toBeTruthy();
  const mediaBody = await media.json();
  expect(mediaBody.total).toBe(1);
  expect(mediaBody.media[0]).toMatchObject({
    original_filename: "launch-card.png",
    mime_type: "image/png",
    usage_count: 0,
    can_delete: true,
    processing_status: "ready",
  });
});
