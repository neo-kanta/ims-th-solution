// Package handler holds HTTP handlers for the reference_data module.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/application"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/transport/dto/request"
	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/transport/dto/response"
	"github.com/neo-kanta/ims-th-solution/backend/platform/httputil"
)

// SecuritiesHandler exposes canonical-security endpoints.
type SecuritiesHandler struct {
	service *application.Service
}

// NewSecuritiesHandler constructs the handler.
func NewSecuritiesHandler(service *application.Service) *SecuritiesHandler {
	return &SecuritiesHandler{service: service}
}

// SearchSecurities handles GET /reference-data/securities/search.
// @Summary Search canonical securities
// @Description Search canonical securities by query, asset type, provider, and status.
// @Tags ReferenceData
// @Security BearerAuth
// @Produce json
// @Param query query string false "Free-text query against ims_symbol/display_symbol/name/isin"
// @Param asset_type query string false "Filter by AssetType (EQUITY, BOND, FX, ...)"
// @Param provider query string false "Only securities with an ACTIVE mapping for this provider_code"
// @Param status query string false "Filter by SecurityStatus (ACTIVE, INACTIVE, SUSPENDED)"
// @Param limit query int false "Result limit (default 50, max 200)"
// @Success 200 {object} response.SearchResponse
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 500 {object} httputil.ErrorResponse
// @Router /reference-data/securities/search [get]
func (h *SecuritiesHandler) SearchSecurities(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 0
	if raw := q.Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	filter := domain.SecuritySearchFilter{
		Query:     q.Get("query"),
		AssetType: domain.AssetType(strings.ToUpper(q.Get("asset_type"))),
		Provider:  q.Get("provider"),
		Status:    domain.SecurityStatus(strings.ToUpper(q.Get("status"))),
		Limit:     limit,
	}
	items, err := h.service.SearchSecurities(r.Context(), filter)
	if err != nil {
		writeReferenceDataError(w, err)
		return
	}
	out := make([]response.SecurityDTO, 0, len(items))
	for _, sec := range items {
		out = append(out, response.ToSecurityDTO(sec))
	}
	httputil.OK(w, response.SearchResponse{Items: out})
}

// CreateSecurity handles POST /reference-data/securities.
// @Summary Create canonical security
// @Description Register a new canonical IMS security.
// @Tags ReferenceData
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body request.CreateSecurityRequest true "Security payload"
// @Success 201 {object} response.SecurityDTO
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /reference-data/securities [post]
func (h *SecuritiesHandler) CreateSecurity(w http.ResponseWriter, r *http.Request) {
	var req request.CreateSecurityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	in := application.CreateSecurityInput{
		IMSSymbol:          req.IMSSymbol,
		DisplaySymbol:      req.DisplaySymbol,
		PrimaryIdentifier:  req.PrimaryIdentifier,
		Name:               req.Name,
		AssetType:          domain.AssetType(strings.ToUpper(req.AssetType)),
		Currency:           req.Currency,
		CountryCode:        req.CountryCode,
		ExchangeMIC:        req.ExchangeMIC,
		ISIN:               req.ISIN,
		CUSIP:              req.CUSIP,
		FIGI:               req.FIGI,
		Status:             domain.SecurityStatus(strings.ToUpper(req.Status)),
		AutoBuildIMSSymbol: req.AutoBuildIMSSymbol,
	}
	sec, err := h.service.CreateSecurity(r.Context(), in)
	if err != nil {
		writeReferenceDataError(w, err)
		return
	}
	httputil.Created(w, response.ToSecurityDTO(*sec))
}

// GetSecurity handles GET /reference-data/securities/{security_id}.
// @Summary Get canonical security
// @Tags ReferenceData
// @Security BearerAuth
// @Produce json
// @Param security_id path string true "Security id"
// @Success 200 {object} response.SecurityDTO
// @Failure 404 {object} httputil.ErrorResponse
// @Router /reference-data/securities/{security_id} [get]
func (h *SecuritiesHandler) GetSecurity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "security_id")
	sec, err := h.service.GetSecurityByID(r.Context(), id)
	if err != nil {
		writeReferenceDataError(w, err)
		return
	}
	httputil.OK(w, response.ToSecurityDTO(*sec))
}

// UpdateSecurity handles PATCH /reference-data/securities/{security_id}.
// @Summary Patch canonical security
// @Tags ReferenceData
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param security_id path string true "Security id"
// @Param request body request.UpdateSecurityRequest true "Patch payload"
// @Success 200 {object} response.SecurityDTO
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /reference-data/securities/{security_id} [patch]
func (h *SecuritiesHandler) UpdateSecurity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "security_id")
	var req request.UpdateSecurityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	in := application.UpdateSecurityInput{
		ID:                id,
		DisplaySymbol:     req.DisplaySymbol,
		PrimaryIdentifier: req.PrimaryIdentifier,
		Name:              req.Name,
		ExchangeMIC:       req.ExchangeMIC,
		Currency:          req.Currency,
		CountryCode:       req.CountryCode,
		ISIN:              req.ISIN,
		CUSIP:             req.CUSIP,
		FIGI:              req.FIGI,
	}
	if req.AssetType != nil {
		at := domain.AssetType(strings.ToUpper(*req.AssetType))
		in.AssetType = &at
	}
	if req.Status != nil {
		st := domain.SecurityStatus(strings.ToUpper(*req.Status))
		in.Status = &st
	}
	sec, err := h.service.UpdateSecurity(r.Context(), in)
	if err != nil {
		writeReferenceDataError(w, err)
		return
	}
	httputil.OK(w, response.ToSecurityDTO(*sec))
}

// ListSecurityMappings handles GET /reference-data/securities/{security_id}/mappings.
// @Summary List provider mappings for a security
// @Tags ReferenceData
// @Security BearerAuth
// @Produce json
// @Param security_id path string true "Security id"
// @Success 200 {object} response.MappingsResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /reference-data/securities/{security_id}/mappings [get]
func (h *SecuritiesHandler) ListSecurityMappings(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "security_id")
	items, err := h.service.ListProviderMappings(r.Context(), id)
	if err != nil {
		writeReferenceDataError(w, err)
		return
	}
	out := make([]response.ProviderMappingDTO, 0, len(items))
	for _, m := range items {
		out = append(out, response.ToProviderMappingDTO(m))
	}
	httputil.OK(w, response.MappingsResponse{Items: out})
}

// AddSecurityMapping handles POST /reference-data/securities/{security_id}/mappings.
// @Summary Add provider mapping to a security
// @Tags ReferenceData
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param security_id path string true "Security id"
// @Param request body request.AddProviderMappingRequest true "Provider mapping payload"
// @Success 201 {object} response.ProviderMappingDTO
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Router /reference-data/securities/{security_id}/mappings [post]
func (h *SecuritiesHandler) AddSecurityMapping(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "security_id")
	var req request.AddProviderMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	in := application.AddProviderMappingInput{
		SecurityID:        id,
		ProviderCode:      req.ProviderCode,
		ProviderSymbol:    req.ProviderSymbol,
		ProviderExchange:  req.ProviderExchange,
		ProviderAssetType: req.ProviderAssetType,
		ProviderCurrency:  req.ProviderCurrency,
		Priority:          req.Priority,
		IsPrimary:         req.IsPrimary,
	}
	if req.ConfidenceScore != nil {
		dec, err := decimal.NewFromString(*req.ConfidenceScore)
		if err != nil {
			httputil.BadRequest(w, "confidence_score must be a decimal string")
			return
		}
		in.ConfidenceScore = &dec
	}
	m, err := h.service.AddProviderMapping(r.Context(), in)
	if err != nil {
		writeReferenceDataError(w, err)
		return
	}
	httputil.Created(w, response.ToProviderMappingDTO(*m))
}

// DeleteSecurityMapping handles DELETE /reference-data/securities/{security_id}/mappings/{mapping_id}.
// @Summary Soft-delete a provider mapping
// @Description Marks the mapping as INACTIVE rather than removing it physically.
// @Tags ReferenceData
// @Security BearerAuth
// @Param security_id path string true "Security id"
// @Param mapping_id path string true "Mapping id"
// @Success 204 "No Content"
// @Failure 404 {object} httputil.ErrorResponse
// @Router /reference-data/securities/{security_id}/mappings/{mapping_id} [delete]
func (h *SecuritiesHandler) DeleteSecurityMapping(w http.ResponseWriter, r *http.Request) {
	mappingID := chi.URLParam(r, "mapping_id")
	if err := h.service.RemoveProviderMapping(r.Context(), mappingID); err != nil {
		writeReferenceDataError(w, err)
		return
	}
	httputil.NoContent(w)
}

// ListUnmappedCandidates handles GET /reference-data/unmapped-candidates.
// @Summary List unmapped provider symbol candidates
// @Tags ReferenceData
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by candidate_status"
// @Param provider query string false "Filter by provider_code"
// @Param batch_id query string false "Filter by import batch id"
// @Param limit query int false "Result limit (default 100, max 500)"
// @Success 200 {object} response.CandidatesResponse
// @Router /reference-data/unmapped-candidates [get]
func (h *SecuritiesHandler) ListUnmappedCandidates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 0
	if raw := q.Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}
	items, err := h.service.ListUnmappedCandidates(r.Context(), domain.UnmappedCandidateFilter{
		Status:       strings.ToUpper(q.Get("status")),
		ProviderCode: q.Get("provider"),
		BatchID:      q.Get("batch_id"),
		Limit:        limit,
	})
	if err != nil {
		writeReferenceDataError(w, err)
		return
	}
	out := make([]response.UnmappedCandidateDTO, 0, len(items))
	for _, c := range items {
		out = append(out, response.ToCandidateDTO(c))
	}
	httputil.OK(w, response.CandidatesResponse{Items: out})
}

// MapUnmappedCandidate handles POST /reference-data/unmapped-candidates/{id}/map.
// @Summary Map a candidate to an existing security
// @Tags ReferenceData
// @Security BearerAuth
// @Accept json
// @Param candidate_id path string true "Candidate id"
// @Param request body request.MapCandidateRequest true "Security id to map to"
// @Success 204 "No Content"
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /reference-data/unmapped-candidates/{candidate_id}/map [post]
func (h *SecuritiesHandler) MapUnmappedCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "candidate_id")
	var req request.MapCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	if err := h.service.MapUnmappedCandidate(r.Context(), id, req.SecurityID); err != nil {
		writeReferenceDataError(w, err)
		return
	}
	httputil.NoContent(w)
}

// RejectUnmappedCandidate handles POST /reference-data/unmapped-candidates/{id}/reject.
// @Summary Reject an unmapped candidate
// @Tags ReferenceData
// @Security BearerAuth
// @Accept json
// @Param candidate_id path string true "Candidate id"
// @Param request body request.RejectCandidateRequest true "Rejection reason"
// @Success 204 "No Content"
// @Failure 400 {object} httputil.ErrorResponse
// @Failure 404 {object} httputil.ErrorResponse
// @Failure 409 {object} httputil.ErrorResponse
// @Router /reference-data/unmapped-candidates/{candidate_id}/reject [post]
func (h *SecuritiesHandler) RejectUnmappedCandidate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "candidate_id")
	var req request.RejectCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.BadRequest(w, "invalid JSON body")
		return
	}
	if err := h.service.RejectUnmappedCandidate(r.Context(), id, req.Reason); err != nil {
		writeReferenceDataError(w, err)
		return
	}
	httputil.NoContent(w)
}

func writeReferenceDataError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrSecurityNotFound),
		errors.Is(err, application.ErrCandidateNotFound),
		errors.Is(err, application.ErrMappingNotFound):
		httputil.NotFound(w, err.Error())
	case errors.Is(err, application.ErrSecurityConflict),
		errors.Is(err, application.ErrCandidateNotPending):
		httputil.Conflict(w, err.Error())
	case errors.Is(err, application.ErrInvalidRequest):
		httputil.BadRequest(w, err.Error())
	default:
		httputil.InternalError(w, err.Error())
	}
}
