const DEFAULT_PAGE_SIZE = 200;
const DEFAULT_MAX_PAGES = 500;

export interface CollectedPages<T> {
  items: T[];
  total: number;
}

export interface OffsetPage<T> {
  items: T[];
  total?: number;
  offset?: number;
}

export interface NumberedPage<T> {
  items: T[];
  total?: number;
  page?: number;
}

function pageTotal(value: number | undefined): number {
  if (!Number.isSafeInteger(value) || (value ?? -1) < 0) {
    throw new Error("The API returned invalid pagination metadata.");
  }
  return value as number;
}

/**
 * Collect an offset-paginated endpoint without treating its first page as the
 * complete directory. A changing total or non-advancing page fails closed.
 */
export async function collectOffsetPages<T>(
  loadPage: (offset: number, limit: number) => Promise<OffsetPage<T>>,
  pageSize = DEFAULT_PAGE_SIZE,
  maxPages = DEFAULT_MAX_PAGES,
): Promise<CollectedPages<T>> {
  const items: T[] = [];
  let expectedTotal: number | null = null;
  let offset = 0;

  for (let pageIndex = 0; pageIndex < maxPages; pageIndex += 1) {
    const payload = await loadPage(offset, pageSize);
    const total = pageTotal(payload.total);

    if (expectedTotal === null) expectedTotal = total;
    if (total !== expectedTotal) {
      throw new Error("The API result changed while pagination was in progress.");
    }
    if (payload.offset !== undefined && payload.offset !== offset) {
      throw new Error("The API returned a non-advancing pagination offset.");
    }

    items.push(...payload.items);
    if (items.length === expectedTotal) return { items, total: expectedTotal };
    if (items.length > expectedTotal) {
      throw new Error("The API returned more records than its reported total.");
    }
    if (payload.items.length === 0) {
      throw new Error("The API pagination ended before all records were loaded.");
    }

    offset += payload.items.length;
  }

  throw new Error("The API pagination exceeded the safe page limit.");
}

/** Page-number variant used by the portfolio directory endpoint. */
export async function collectNumberedPages<T>(
  loadPage: (page: number, limit: number) => Promise<NumberedPage<T>>,
  pageSize = DEFAULT_PAGE_SIZE,
  maxPages = DEFAULT_MAX_PAGES,
): Promise<CollectedPages<T>> {
  const items: T[] = [];
  let expectedTotal: number | null = null;

  for (let page = 1; page <= maxPages; page += 1) {
    const payload = await loadPage(page, pageSize);
    const total = pageTotal(payload.total);

    if (expectedTotal === null) expectedTotal = total;
    if (total !== expectedTotal) {
      throw new Error("The API result changed while pagination was in progress.");
    }
    if (payload.page !== undefined && payload.page !== page) {
      throw new Error("The API returned a non-advancing page number.");
    }

    items.push(...payload.items);
    if (items.length === expectedTotal) return { items, total: expectedTotal };
    if (items.length > expectedTotal) {
      throw new Error("The API returned more records than its reported total.");
    }
    if (payload.items.length === 0) {
      throw new Error("The API pagination ended before all records were loaded.");
    }
  }

  throw new Error("The API pagination exceeded the safe page limit.");
}
