package handler

import (
	"net/http"

	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

func writeComplianceStatusError(w http.ResponseWriter, code, message, checkGroupID string) {
	details := map[string]string{}
	if checkGroupID != "" {
		details["check_group_id"] = checkGroupID
	}
	httputil.JSON(w, http.StatusUnprocessableEntity, httputil.ErrorResponse{
		Error:   message,
		Code:    code,
		Details: details,
	})
}
