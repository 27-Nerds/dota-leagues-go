import { cp, mkdir, readdir } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
const root = new URL('../', import.meta.url);
const destination = new URL('../public/', root);
await mkdir(destination, { recursive: true });
// Copy only frontend build artifacts; preserve downloaded league/team assets.
for (const entry of await readdir(new URL('dist/', root))) {
  await cp(new URL('dist/' + entry, root), new URL(entry, destination), { recursive: true });
}
console.log('Frontend deployed to ' + fileURLToPath(destination));
