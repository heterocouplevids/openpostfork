import { expect, test } from "@playwright/test";
import { createWorkspace, registerUser } from "./helpers";

type PostPayload = {
  workspace_id?: string;
  content?: string;
  social_account_ids?: string[];
  scheduled_at?: string;
  media_ids?: string[];
  random_delay_minutes?: number;
  [key: string]: unknown;
};

test("composer schedules a post from the suggested next slot", async ({
  page,
  request,
}) => {
  const unique = Date.now().toString(36);
  const email = `composer-scheduling-${unique}@example.com`;
  const postContent = "Schedule this launch note from the composer.";
  const suggestedDate = new Date(Date.now() + 48 * 60 * 60 * 1000)
    .toISOString()
    .slice(0, 10);
  const suggestedSlotTime = `${suggestedDate}T10:30:00Z`;
  let suggestedWorkspaceId = "";
  let draftPayload: PostPayload | undefined;
  let scheduledPayload: PostPayload | undefined;

  const auth = await registerUser(request, email);
  const workspaceBody = await createWorkspace(
    request,
    auth.token,
    "Composer Scheduling E2E",
  );

  await page.addInitScript((token) => {
    window.localStorage.setItem("token", token);
  }, auth.token);
  await page.route("**/api/v1/accounts?**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: [
        {
          id: "bluesky-main",
          slug: "bluesky-main",
          platform: "bluesky",
          account_id: "bsky-main",
          account_username: "openpost.bsky.social",
          account_avatar_url: "",
          instance_url: "",
          is_active: true,
          thread_replies_supported: true,
        },
      ],
    });
  });
  await page.route(
    "**/api/v1/posting-schedules/next-slot?**",
    async (route) => {
      const url = new URL(route.request().url());
      suggestedWorkspaceId = url.searchParams.get("workspace_id") ?? "";
      await route.fulfill({
        contentType: "application/json",
        json: {
          slot_time: suggestedSlotTime,
          message: "Next available slot found",
          slot: {
            id: "slot-e2e",
            workspace_id: workspaceBody.id,
            day_of_week: 4,
            time_of_day: "10:30",
            label: "Launch slot",
            is_active: true,
            set_id: "",
          },
        },
      });
    },
  );
  await page.route("**/api/v1/posts", async (route) => {
    if (route.request().method() === "POST") {
      const body = JSON.parse(
        route.request().postData() ?? "{}",
      ) as PostPayload;
      if (body.scheduled_at) {
        scheduledPayload = body;
      } else {
        draftPayload = body;
      }

      await route.fulfill({
        contentType: "application/json",
        json: {
          id: body.scheduled_at ? "scheduled-post" : "draft-schedule",
          workspace_id: body.workspace_id,
          content: body.content,
          status: body.scheduled_at ? "scheduled" : "draft",
          scheduled_at: body.scheduled_at ?? "",
          media: [],
          destinations: [],
        },
      });
      return;
    }

    await route.continue();
  });
  await page.route("**/api/v1/posts/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith("/variants")) {
      await route.fulfill({ status: 204 });
      return;
    }

    if (route.request().method() === "PATCH") {
      const body = JSON.parse(
        route.request().postData() ?? "{}",
      ) as PostPayload;
      if (body.scheduled_at) {
        scheduledPayload = body;
      }
      await route.fulfill({
        contentType: "application/json",
        json: {
          id: "draft-schedule",
          workspace_id: workspaceBody.id,
          content: body.content,
          status: body.scheduled_at ? "scheduled" : "draft",
          scheduled_at: body.scheduled_at ?? "",
          media: [],
          destinations: [],
        },
      });
      return;
    }

    await route.continue();
  });

  await page.goto("/");
  await page.locator("textarea").first().fill(postContent);
  await page.getByRole("button", { name: "Suggest" }).click();
  await expect.poll(() => suggestedWorkspaceId).toBe(workspaceBody.id);
  await expect(page.getByRole("button", { name: "Schedule" })).toBeEnabled();
  await page.getByRole("button", { name: "Schedule" }).click();

  await expect(page.getByText("Scheduled!")).toBeVisible();
  await expect.poll(() => scheduledPayload).toBeTruthy();

  expect(scheduledPayload).toMatchObject({
    content: postContent,
    social_account_ids: ["bluesky-main"],
    media_ids: [],
  });
  expect(scheduledPayload?.workspace_id ?? draftPayload?.workspace_id).toBe(
    workspaceBody.id,
  );
  expect(scheduledPayload?.scheduled_at).toBeTruthy();
  expect(new Date(scheduledPayload?.scheduled_at ?? "").toString()).not.toBe(
    "Invalid Date",
  );
  expect(scheduledPayload?.scheduled_at).toContain(`${suggestedDate}T`);
});
