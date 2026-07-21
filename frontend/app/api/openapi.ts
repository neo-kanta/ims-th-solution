import { useRuntimeConfig } from "#imports";
import createClient from "openapi-fetch";
import type { Client } from "openapi-fetch";

import { useAuthStore } from "~/stores/useAuthStore";

import { v2BaseUrlFrom } from "./urls";
import type { paths } from "./ims-api";

export type ImsOpenApiClient = Client<paths>;

interface ApiEnvelope<T> {
  // Swagger models httputil.SuccessResponse.data as optional because the
  // generic Go field is overridden by each endpoint annotation. Runtime
  // success responses still require it, which the unwrap helpers enforce.
  data?: T;
  message?: string;
}

type OpenApiResult<T> =
  | {
      data: T | ApiEnvelope<T>;
      error?: never;
      response: Response;
    }
  | {
      data?: never;
      error: unknown;
      response: Response;
    };

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isApiEnvelope<T>(value: unknown): value is ApiEnvelope<T> {
  if (!isRecord(value) || !("data" in value)) {
    return false;
  }

  return Object.keys(value).every((key) => key === "data" || key === "message");
}

function getErrorMessage(error: unknown, response: Response): string {
  if (isRecord(error)) {
    const apiError = error.error;
    if (typeof apiError === "string" && apiError.trim()) {
      return apiError;
    }

    const message = error.message;
    if (typeof message === "string" && message.trim()) {
      return message;
    }
  }

  return response.statusText || "API request failed";
}

export class OpenApiRequestError extends Error {
  status: number;
  details: unknown;

  constructor(response: Response, details: unknown) {
    super(getErrorMessage(details, response));
    this.name = "OpenApiRequestError";
    this.status = response.status;
    this.details = details;
  }
}

export function assertOpenApiResponse(result: OpenApiResult<unknown>): void {
  if ("error" in result && result.error !== undefined) {
    throw new OpenApiRequestError(result.response, result.error);
  }
}

/**
 * Defensive envelope unwrap for both accurately documented
 * `httputil.SuccessResponse{data=...}` endpoints and older endpoints whose
 * annotations still describe only the inner shape.
 *
 * Shared with the workflow store and the dashboard tasks composable so the
 * "swallow the envelope" rule lives in exactly one place.
 */
export function unwrapEnvelope<T>(payload: unknown): T | undefined {
  if (payload !== null && isApiEnvelope<T>(payload)) {
    return payload.data;
  }
  return payload as T | undefined;
}

export function unwrapOpenApiResponse<T>(result: OpenApiResult<T>): T {
  assertOpenApiResponse(result);

  if (!("data" in result) || result.data === undefined) {
    throw new OpenApiRequestError(result.response, "API response did not include data");
  }

  if (isApiEnvelope<T>(result.data)) {
    if (result.data.data === undefined) {
      throw new OpenApiRequestError(
        result.response,
        "API response envelope did not include data",
      );
    }
    return result.data.data;
  }

  return result.data;
}

function createAuthedClient(baseUrl: string): ImsOpenApiClient {
  const authStore = useAuthStore();

  const client = createClient<paths>({
    baseUrl,
  });

  client.use({
    onRequest({ request }) {
      if (authStore.token) {
        request.headers.set("Authorization", `Bearer ${authStore.token}`);
      }

      return request;
    },
    onResponse({ response }) {
      if (response.status === 401) {
        authStore.clearAuth();
      }

      return response;
    },
  });

  return client;
}

export function useOpenApiClient(): ImsOpenApiClient {
  const config = useRuntimeConfig();
  const baseUrl = (import.meta.server
    ? config.apiBaseUrl
    : config.public.apiBaseUrl) as string;

  return createAuthedClient(baseUrl);
}

/**
 * Client for Portfolio V2 routes (docs/api/portfolio-v2-api-ddd.md), mounted
 * under `/api/v2` alongside the V1 API under `/api/v1`. See
 * {@link v2BaseUrlFrom} (app/api/urls.ts) for the base URL derivation.
 */
export function useOpenApiClientV2(): ImsOpenApiClient {
  const config = useRuntimeConfig();
  const v1BaseUrl = (import.meta.server
    ? config.apiBaseUrl
    : config.public.apiBaseUrl) as string;

  return createAuthedClient(v2BaseUrlFrom(v1BaseUrl));
}
