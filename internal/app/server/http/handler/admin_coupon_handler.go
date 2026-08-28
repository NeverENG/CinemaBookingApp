package handler

import (
	"net/http"
	"strconv"

	"github.com/NeverENG/CinemaBookingApp/internal/app/server/http/resp"
	"github.com/NeverENG/CinemaBookingApp/internal/app/server/service"
	"github.com/gin-gonic/gin"
)

// AdminCouponHandler 优惠券管理。
type AdminCouponHandler struct {
	coupons *service.AdminCouponSvc
}

func NewAdminCouponHandler(coupons *service.AdminCouponSvc) *AdminCouponHandler {
	return &AdminCouponHandler{coupons: coupons}
}

type couponTemplateRequest struct {
	Name             string `json:"name" binding:"required"`
	Type             string `json:"type" binding:"required"`
	ValueCents       int64  `json:"value_cents"`
	PercentBp        int    `json:"percent_bp"`
	MinSpendCents    int64  `json:"min_spend_cents"`
	MaxDiscountCents int64  `json:"max_discount_cents"`
	Redeemable       bool   `json:"redeemable"`
	RedeemPoints     int    `json:"redeem_points"`
	ValidDays        int    `json:"valid_days"`
	TotalQty         int    `json:"total_qty"`
	PerUserLimit     int    `json:"per_user_limit"`
}

func (h *AdminCouponHandler) CreateTemplate(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	var req couponTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	tpl, err := h.coupons.CreateTemplate(c.Request.Context(), scope.AdminID, service.CouponTemplateInput{
		Name: req.Name, Type: req.Type, ValueCents: req.ValueCents, PercentBp: req.PercentBp,
		MinSpendCents: req.MinSpendCents, MaxDiscountCents: req.MaxDiscountCents, Redeemable: req.Redeemable,
		RedeemPoints: req.RedeemPoints, ValidDays: req.ValidDays, TotalQty: req.TotalQty, PerUserLimit: req.PerUserLimit,
	})
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, tpl)
}

func (h *AdminCouponHandler) ListTemplates(c *gin.Context) {
	templates, err := h.coupons.ListTemplates(c.Request.Context())
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, templates)
}

func (h *AdminCouponHandler) SetTemplateStatus(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	templateID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, http.StatusBadRequest, "invalid template id")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.coupons.SetTemplateStatus(c.Request.Context(), scope.AdminID, templateID, req.Status); err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, gin.H{"id": templateID, "status": req.Status})
}

type issueCouponRequest struct {
	UserID     int64 `json:"user_id" binding:"required"`
	TemplateID int64 `json:"template_id" binding:"required"`
}

func (h *AdminCouponHandler) IssueToUser(c *gin.Context) {
	scope, ok := adminScopeFrom(c)
	if !ok {
		resp.Fail(c, http.StatusUnauthorized, "missing admin")
		return
	}
	var req issueCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	coupon, err := h.coupons.IssueToUser(c.Request.Context(), scope.AdminID, req.UserID, req.TemplateID)
	if err != nil {
		resp.Error(c, err)
		return
	}
	resp.OK(c, coupon)
}
