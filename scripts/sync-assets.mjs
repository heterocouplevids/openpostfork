import { cp, mkdir, rm } from "node:fs/promises";
import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, "..");
const source = path.join(root, "assets");
const frontendRoot = path.join(root, "frontend");

const targets = [
  path.join(frontendRoot, "static", "assets"),
  path.join(root, "docs-site", "public", "assets"),
];

const brandIconSource = path.join(source, "brand", "icon.svg");
const capacitorAssetsTarget = path.join(frontendRoot, "assets");

if (!existsSync(source)) {
  console.error("Missing assets/ directory");
  process.exit(1);
}

for (const target of targets) {
  await rm(target, { recursive: true, force: true });
  await mkdir(path.dirname(target), { recursive: true });
  await cp(source, target, { recursive: true });
  console.log(`Synced assets -> ${path.relative(root, target)}`);
}

if (!existsSync(brandIconSource)) {
  console.error("Missing brand icon at assets/brand/icon.svg");
  process.exit(1);
}

await mkdir(capacitorAssetsTarget, { recursive: true });
await cp(brandIconSource, path.join(capacitorAssetsTarget, "logo.svg"));
console.log("Prepared frontend/assets/logo.svg for Capacitor asset generation");
