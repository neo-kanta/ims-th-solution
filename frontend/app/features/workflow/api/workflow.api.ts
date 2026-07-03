import { unwrapOpenApiResponse, useOpenApiClient } from "~/api/openapi";
import type { components } from "~/api/ims-api";

export const workflowApi = {
  /**
   * getDailyWorkflow - GET /workflow/daily?businessDate=...
   */
  async getDailyWorkflow(businessDate: string): Promise<components["schemas"]["DailyWorkflowResponse"]> {
    const client = useOpenApiClient();
    const response = await client.GET("/workflow/daily", {
      params: {
        query: { businessDate },
      },
    });
    return unwrapOpenApiResponse<components["schemas"]["DailyWorkflowResponse"]>(response);
  },

  /**
   * executeDailyTransition - POST /workflow/daily/execute
   */
  async executeDailyTransition(payload: components["schemas"]["DailyExecuteRequest"]): Promise<components["schemas"]["DailyWorkflowResponse"]> {
    const client = useOpenApiClient();
    const response = await client.POST("/workflow/daily/execute", {
      body: payload,
    });
    return unwrapOpenApiResponse<components["schemas"]["DailyWorkflowResponse"]>(response);
  },

  /**
   * getDailyTransitions - GET /workflow/daily/transitions?businessDate=...&page=...&pageSize=...
   */
  async getDailyTransitions(
    businessDate: string,
    page = 1,
    pageSize = 50,
  ): Promise<components["schemas"]["DailyTransitionsResponse"]> {
    const client = useOpenApiClient();
    const response = await client.GET("/workflow/daily/transitions", {
      params: {
        query: { businessDate, page, pageSize },
      },
    });
    return unwrapOpenApiResponse<components["schemas"]["DailyTransitionsResponse"]>(response);
  },

  /**
   * getTransitionRules - GET /workflow/transition-rules
   */
  async getTransitionRules(): Promise<components["schemas"]["TransitionRulesResponse"]> {
    const client = useOpenApiClient();
    const response = await client.GET("/workflow/transition-rules");
    return unwrapOpenApiResponse<components["schemas"]["TransitionRulesResponse"]>(response);
  },

  /**
   * getWorkflowSettings - GET /workflow/settings
   */
  async getWorkflowSettings(): Promise<components["schemas"]["WorkflowSettingsResponse"]> {
    const client = useOpenApiClient();
    const response = await client.GET("/workflow/settings");
    return unwrapOpenApiResponse<components["schemas"]["WorkflowSettingsResponse"]>(response);
  },

  /**
   * updateWorkflowSettings - PUT /workflow/settings
   */
  async updateWorkflowSettings(payload: components["schemas"]["DailySettingsUpdateRequest"]): Promise<components["schemas"]["WorkflowSettingsResponse"]> {
    const client = useOpenApiClient();
    const response = await client.PUT("/workflow/settings", {
      body: payload,
    });
    return unwrapOpenApiResponse<components["schemas"]["WorkflowSettingsResponse"]>(response);
  },
};

