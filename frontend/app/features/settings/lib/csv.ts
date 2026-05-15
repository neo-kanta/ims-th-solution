/**
 * Spreadsheet formula injection prevention. If a cell starts with `=`, `+`,
 * `-`, or `@`, prefix with a single quote so Excel/Google Sheets does not
 * evaluate the cell as a formula. Mirrors the backend audit-export sanitizer
 * (see backend/internal/audit/transport/handler/audit_handler.go::sanitizeCSVCell).
 */
function sanitizeCsvCell(value: string): string {
  if (value === "") return value;
  const trimmed = value.replace(/^[\s\t]+/, "");
  if (trimmed === "") return value;
  const first = trimmed[0];
  if (first === "=" || first === "+" || first === "-" || first === "@") {
    return `'${value}`;
  }
  return value;
}

/**
 * Quotes a CSV field per RFC 4180 if it contains comma, quote, CR, or LF.
 * Inner double-quotes are doubled.
 */
function quoteCsvField(value: string): string {
  if (/[",\r\n]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`;
  }
  return value;
}

export function buildCsv(headers: string[], rows: string[][]): Blob {
  const encoder = "﻿"; // UTF-8 BOM so Excel detects encoding
  const lines: string[] = [];
  lines.push(headers.map((h) => quoteCsvField(sanitizeCsvCell(h))).join(","));
  for (const row of rows) {
    lines.push(row.map((cell) => quoteCsvField(sanitizeCsvCell(cell))).join(","));
  }
  return new Blob([encoder + lines.join("\r\n")], {
    type: "text/csv;charset=utf-8",
  });
}

export function downloadBlob(blob: Blob, filename: string): void {
  if (typeof window === "undefined") return;
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}
