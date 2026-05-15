import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import openapiTS, { astToString, COMMENT_HEADER } from "openapi-typescript";
import swagger2openapi from "swagger2openapi";

const scriptDir = dirname(fileURLToPath(import.meta.url));
const frontendDir = resolve(scriptDir, "..");
const repoRoot = resolve(frontendDir, "..");
const inputPath = resolve(repoRoot, "backend", "docs", "swagger.json");
const outputPath = resolve(frontendDir, "app", "api", "ims-api.d.ts");

const swagger = JSON.parse(await readFile(inputPath, "utf8"));
const converted = await swagger2openapi.convertObj(swagger, {
  patch: true,
  warnOnly: true,
});

const ast = await openapiTS(converted.openapi);
const source = `${COMMENT_HEADER}${astToString(ast)}\n`;

await mkdir(dirname(outputPath), { recursive: true });
await writeFile(outputPath, source, "utf8");

console.log(
  `Generated ${relative(frontendDir, outputPath)} from ${relative(frontendDir, inputPath)}`,
);
