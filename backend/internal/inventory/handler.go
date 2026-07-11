package inventory

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Oliveszn/OneDesk/internal/billing"
	"github.com/Oliveszn/OneDesk/internal/httputil"
	"github.com/Oliveszn/OneDesk/internal/reqctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(s *Service, l *slog.Logger) *Handler {
	return &Handler{
		service: s,
		logger:  l,
	}
}

// CreateWarehouse godoc
//
//	@Summary		Create a new warehouse
//	@Description	Deploys a distinct inventory tracking location workspace for the tenant.
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateWarehouseRequest	true	"Warehouse creation parameters"
//	@Success		201		{object}	WarehouseResponse
//	@Failure		400		{object}	map[string]string	"Invalid payload or request processing error"
//	@Failure		401		{object}	map[string]string	"Missing valid tenant validation credentials"
//	@Router			/v1/warehouses [post]
func (h *Handler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("create_warehouse: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req CreateWarehouseRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("create_warehouse: invalid structural payload decode match", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.CreateWarehouse(r.Context(), tenantID, req.Name)
	if err != nil {
		h.logger.Error("create_warehouse: processing pipeline error", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, resp)
}

// ListWarehouses godoc
//
//	@Summary		List all warehouses
//	@Description	Returns a full list of storage locations associated with the active tenant.
//	@Tags			inventory
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		WarehouseResponse
//	@Failure		401	{object}	map[string]string	"Missing valid tenant validation credentials"
//	@Failure		500	{object}	map[string]string	"Internal service storage lookup engine failure"
//	@Router			/v1/warehouses [get]
func (h *Handler) ListWarehouses(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("list_warehouses: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	resp, err := h.service.ListWarehouses(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("list_warehouses: database query failure", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// CreateProduct godoc
//
//	@Summary		Create a product item
//	@Description	Adds a unique tracking product variant profile under the current client scope. Enforces item ceiling controls.
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateProductRequest	true	"Product registration data template"
//	@Success		201		{object}	ProductResponse
//	@Failure		400		{object}	map[string]string	"Invalid structural input payload values"
//	@Failure		401		{object}	map[string]string	"Missing valid tenant validation credentials"
//	@Failure		403		{object}	map[string]string	"Product catalog resource entitlement quota exhausted"
//	@Failure		409		{object}	map[string]string	"SKU code entry collision matching active entity records"
//	@Router			/v1/products [post]
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("create_product: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req CreateProductRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("create_product: JSON syntax mapping failure", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.CreateProduct(r.Context(), tenantID, req)
	if err != nil {
		switch {
		case errors.Is(err, billing.ErrPlanLimitReached):
			h.logger.Warn("create_product: plan ceiling enforcement bounce block", "tenant_id", tenantID, "sku", req.SKU)
			httputil.WriteError(w, http.StatusForbidden, "product limit reached for your plan — upgrade to add more")
		case errors.Is(err, ErrSKUTaken):
			h.logger.Warn("create_product: duplicate identifier registration reject", "tenant_id", tenantID, "sku", req.SKU)
			httputil.WriteError(w, http.StatusConflict, "sku already exists")
		default:
			h.logger.Error("create_product: catalog save block failure", "tenant_id", tenantID, "error", err.Error())
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, resp)
}

// ListProducts godoc
//
//	@Summary		List catalog tracking records
//	@Description	Provides full product catalog metrics visibility owned by the matching business group profile.
//	@Tags			inventory
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		ProductResponse
//	@Failure		401	{object}	map[string]string	"Missing valid tenant validation credentials"
//	@Failure		500	{object}	map[string]string	"Core relational lookup tracking query fault context"
//	@Router			/v1/products [get]
func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("list_products: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	resp, err := h.service.ListProducts(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("list_products: transactional tracking collection dump fault", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// AdjustStock godoc
//
//	@Summary		Adjust product stock level
//	@Description	Mutates available balance counts up or down across distinct warehouses. Rejects operations that result in a negative total.
//	@Tags			inventory
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			productId	path		string				true	"Target catalog item system resource ID (UUID)"
//	@Param			request		body		AdjustStockRequest	true	"Storage destination map and balancing metrics delta values"
//	@Success		200			{object}	StockLevelResponse
//	@Failure		400			{object}	map[string]string	"Malformed route parameters or tracking payload errors"
//	@Failure		401			{object}	map[string]string	"Missing valid tenant validation credentials"
//	@Failure		404			{object}	map[string]string	"Target unique identification index resource missing"
//	@Failure		409			{object}	map[string]string	"Balancing step rejected because final allocation would fall below zero"
//	@Router			/v1/products/{productId}/stock/adjust [post]
func (h *Handler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("adjust_stock: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		h.logger.Warn("adjust_stock: malformed path token string value", "tenant_id", tenantID, "input_param", chi.URLParam(r, "productId"))
		httputil.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req AdjustStockRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("adjust_stock: target adjustments format parser match crash", "tenant_id", tenantID, "product_id", productID, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.AdjustStock(r.Context(), tenantID, productID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidAdjustment):
			h.logger.Warn("adjust_stock: operational run rejected due to safety depletion constraints", "tenant_id", tenantID, "product_id", productID, "warehouse_id", req.WarehouseID, "delta_attempt", req.Delta)
			httputil.WriteError(w, http.StatusConflict, "adjustment would result in negative stock")
		case errors.Is(err, ErrProductNotFound):
			h.logger.Warn("adjust_stock: target context record query not found matching tracking fields", "tenant_id", tenantID, "product_id", productID)
			httputil.WriteError(w, http.StatusNotFound, "product not found")
		default:
			h.logger.Error("adjust_stock: transaction commit worker breakdown step", "tenant_id", tenantID, "product_id", productID, "error", err.Error())
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// GetStock godoc
//
//	@Summary		Get product stock levels
//	@Description	Returns current stock level breakdowns and reorder triggers across all operational warehouse points.
//	@Tags			inventory
//	@Produce		json
//	@Security		BearerAuth
//	@Param			productId	path		string	true	"Target inventory resource look up track key ID (UUID)"
//	@Success		200			{array}		StockLevelResponse
//	@Failure		400			{object}	map[string]string	"Malformed route identification parameters"
//	@Failure		401			{object}	map[string]string	"Missing valid tenant validation credentials"
//	@Failure		404			{object}	map[string]string	"No item registration matches the requested ID"
//	@Failure		500			{object}	map[string]string	"Internal query execution layer exception structural crash"
//	@Router			/v1/products/{productId}/stock [get]
func (h *Handler) GetStock(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("get_stock: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		h.logger.Warn("get_stock: malformed target path validation index sequence", "tenant_id", tenantID, "input_param", chi.URLParam(r, "productId"))
		httputil.WriteError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	resp, err := h.service.GetStockLevels(r.Context(), tenantID, productID)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			h.logger.Warn("get_stock: request reference item does not exist inside system catalog", "tenant_id", tenantID, "product_id", productID)
			httputil.WriteError(w, http.StatusNotFound, "product not found")
			return
		}
		h.logger.Error("get_stock: analytics persistence tracking step fault", "tenant_id", tenantID, "product_id", productID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
