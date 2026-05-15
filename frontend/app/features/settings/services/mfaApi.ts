import {
  assertOpenApiResponse,
  unwrapOpenApiResponse,
  useOpenApiClient,
} from "~/api/openapi";
import type { paths } from "~/api/ims-api";

import type {
  PersonalMfaEnrollResult,
  PersonalMfaTotpInput,
} from "../account.types";

type MfaEnrollResponse =
  paths["/auth/mfa/enroll"]["post"]["responses"][200]["content"]["application/json"];
type MfaVerifyRequest =
  paths["/auth/mfa/verify"]["post"]["requestBody"]["content"]["application/json"];
type MfaDisableRequest =
  paths["/auth/mfa/disable"]["post"]["requestBody"]["content"]["application/json"];

function normalizeEnroll(response: MfaEnrollResponse): PersonalMfaEnrollResult {
  return {
    provisioning_uri: response.provisioning_uri ?? "",
    recovery_codes: response.recovery_codes ?? [],
  };
}

export const mfaApi = {
  async enroll(): Promise<PersonalMfaEnrollResult> {
    const client = useOpenApiClient();
    const response = await client.POST("/auth/mfa/enroll");
    return normalizeEnroll(unwrapOpenApiResponse<MfaEnrollResponse>(response));
  },

  async verify(payload: PersonalMfaTotpInput): Promise<void> {
    const client = useOpenApiClient();
    const body: MfaVerifyRequest = payload;
    const response = await client.POST("/auth/mfa/verify", { body });
    assertOpenApiResponse(response);
  },

  async disable(payload: PersonalMfaTotpInput): Promise<void> {
    const client = useOpenApiClient();
    const body: MfaDisableRequest = payload;
    const response = await client.POST("/auth/mfa/disable", { body });
    assertOpenApiResponse(response);
  },
};
