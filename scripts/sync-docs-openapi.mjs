import { cp, mkdir } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(scriptDir, '..');
const source = path.join(root, 'frontend', 'openapi.json');

const targets = [
	path.join(root, 'docs-site', '.generated', 'openapi.json'),
	path.join(root, 'docs-site', 'public', 'openapi.json'),
];

for (const target of targets) {
	await mkdir(path.dirname(target), { recursive: true });
	await cp(source, target);
	console.log(`Synced OpenAPI spec -> ${path.relative(root, target)}`);
}

const cliDocs = path.join(root, 'docs-site', 'reference', 'cli.md');
const result = spawnSync('go', ['run', './cmd/openpost-docs', cliDocs], {
	cwd: path.join(root, 'cli'),
	stdio: 'inherit',
});
if (result.status !== 0) {
	throw new Error(`Failed to generate CLI reference docs with exit code ${result.status}`);
}
console.log(`Generated CLI reference -> ${path.relative(root, cliDocs)}`);
