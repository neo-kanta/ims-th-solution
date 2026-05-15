import { useRuntimeConfig } from "#imports";
import createClient from "openapi-fetch";
import type { Client } from "openapi-fetch";

import { useAuthStore } from "~/stores/useAuthStore";

import type { paths } from "./ims-api";

export type ImsOpenApiClient = Client<paths>;

interface ApiEnvelope<T> {
  data: T;
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

export function unwrapOpenApiResponse<T>(result: OpenApiResult<T>): T {
  assertOpenApiResponse(result);

  if (!("data" in result) || result.data === undefined) {
    throw new OpenApiRequestError(result.response, "API response did not include data");
  }

  return isApiEnvelope<T>(result.data) ? result.data.data : result.data;
}

export function useOpenApiClient(): ImsOpenApiClient {
  const config = useRuntimeConfig();
  const authStore = useAuthStore();
  const baseUrl = (import.meta.server
    ? config.apiBaseUrl
    : config.public.apiBaseUrl) as string;

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
