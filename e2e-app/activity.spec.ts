import { expect, test } from "@playwright/test";

type JobFixture = {
  id: string;
  type: string;
  status: string;
  run_at: string;
  attempts: number;
  max_attempts: number;
};

function jobFixture(index: number): JobFixture {
  const runAt = new Date(Date.UTC(2026, 6, 1, 12, index, 0));
  return {
    id: `job-${index.toString().padStart(2, "0")}`,
    type: `publish_post_${index.toString().padStart(2, "0")}`,
    status: "pending",
    run_at: runAt.toISOString(),
    attempts: 0,
    max_attempts: 3,
  };
}

test("activity jobs tab loads additional pages", async ({ page }) => {
  const jobs = Array.from({ length: 55 }, (_, index) => jobFixture(index));

  await page.addInitScript(() => {
    window.localStorage.setItem("token", "activity-token");
  });
  await page.route("**/api/v1/auth/me", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: {
        id: "user-1",
        email: "activity@example.com",
        is_admin: false,
        created_at: "2026-07-01T00:00:00Z",
      },
    });
  });
  await page.route("**/api/v1/workspaces**", async (route) => {
    await route.fulfill({
      contentType: "application/json",
      json: [
        { id: "ws-1", name: "Activity", created_at: "2026-07-01T00:00:00Z" },
      ],
    });
  });
  await page.route("**/api/v1/posts**", async (route) => {
    await route.fulfill({ contentType: "application/json", json: [] });
  });
  await page.route("**/api/v1/jobs**", async (route) => {
    const url = new URL(route.request().url());
    const limit = Number(url.searchParams.get("limit") ?? "50");
    const offset = Number(url.searchParams.get("offset") ?? "0");
    const pageJobs = jobs.slice(offset, offset + limit);
    const nextOffset = offset + pageJobs.length;

    await route.fulfill({
      contentType: "application/json",
      headers: {
        "X-Total-Count": jobs.length.toString(),
        "X-Limit": limit.toString(),
        "X-Offset": offset.toString(),
        "X-Next-Offset": nextOffset.toString(),
        "X-Has-More": (nextOffset < jobs.length).toString(),
      },
      json: pageJobs,
    });
  });

  await page.goto("/activity");
  await page.getByRole("tab", { name: "Jobs" }).click();

  await expect(page.getByTestId("job-row")).toHaveCount(50);
  await expect(page.getByTestId("jobs-load-more")).toBeVisible();
  await page.getByTestId("jobs-load-more").click();

  await expect(page.getByTestId("job-row")).toHaveCount(55);
  await expect(page.getByTestId("jobs-load-more")).toHaveCount(0);
});
