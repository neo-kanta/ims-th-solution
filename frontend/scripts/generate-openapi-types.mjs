import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

import openapiTS, { astToString, COMMENT_HEADER } from "openapi-typescript";
import swagger2openapi from "swagger2openapi";

const scriptDir = dirname(fileURLToPath(import.meta.url));
const frontendDir = resolve(scriptDir, "..");
const repoRoot = resolve(frontendDir, "..");
const v1Path = resolve(repoRoot, "backend", "docs", "swagger.json");
const v2Path = resolve(repoRoot, "backend", "docs", "v2", "v2_swagger.json");
const outputPath = resolve(frontendDir, "app", "api", "ims-api.d.ts");

async function toOpenApi3(path) {
  const swagger = JSON.parse(await readFile(path, "utf8"));
  const converted = await swagger2openapi.convertObj(swagger, {
    patch: true,
    warnOnly: true,
  });
  return converted.openapi;
}

const v1 = await toOpenApi3(v1Path);
const v2 = await toOpenApi3(v2Path);

// Portfolio V2 (docs/api/portfolio-v2-api-ddd.md) is generated as its own
// Swagger spec with its own basePath (/api/v2) — Swagger 2.0 only supports
// one basePath per spec, and V2 doesn't share V1's /api/v1 mount. See
// backend/cmd/server/swagger_v2_docs.go for the "why". Both specs' paths
// and schemas are merged into one generated types file here so the
// frontend keeps a single typed client surface; app/api/openapi.ts's
// useOpenApiClientV2() points its request base URL at /api/v2 to match
// how these V2 path keys are relative (no basePath prefix baked in, same
// as V1's keys are relative to /api/v1).
const merged = {
  ...v1,
  paths: { ...v1.paths, ...v2.paths },
  components: {
    ...v1.components,
    schemas: { ...v1.components?.schemas, ...v2.components?.schemas },
  },
};

const ast = await openapiTS(merged);
const source = `${COMMENT_HEADER}${astToString(ast)}\n`;

await mkdir(dirname(outputPath), { recursive: true });
await writeFile(outputPath, source, "utf8");

console.log(
  `Generated ${relative(frontendDir, outputPath)} from ${relative(frontendDir, v1Path)} + ${relative(frontendDir, v2Path)}`,
);
