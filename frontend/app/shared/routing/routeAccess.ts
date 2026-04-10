function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function resolveNestedMeta(meta: Record<string, unknown>) {
  const nested = meta.meta;
  return isRecord(nested) ? nested : undefined;
}

export function resolveRequiredPermission(meta: unknown): string | undefined {
  if (!isRecord(meta)) {
    return undefined;
  }

  if (typeof meta.permission === "string") {
    return meta.permission;
  }

  const nestedMeta = resolveNestedMeta(meta);
  return typeof nestedMeta?.permission === "string"
    ? nestedMeta.permission
    : undefined;
}

export function isPublicRouteMeta(meta: unknown): boolean {
  if (!isRecord(meta)) {
    return false;
  }

  if (meta.auth === false) {
    return true;
  }

  const nestedMeta = resolveNestedMeta(meta);
  return nestedMeta?.auth === false;
}
