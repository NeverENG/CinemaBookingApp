package resp

import (
	"errors"
	"net/http"

	"github.com/NeverENG/CinemaBookingApp/internal/core/domain"
	"github.com/gin-gonic/gin"
)

// Body 统一响应体。
type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: 0, Msg: "ok", Data: data})
}

func Fail(c *gin.Context, status int, msg string) {
	c.JSON(status, Body{Code: status, Msg: msg})
}

// Error 把领域错误翻译成 HTTP 状态码与文案；未知错误统一 500。
// 业务规则在 domain，HTTP 翻译只在这一处。
func Error(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrSessionNotFound),
		errors.Is(err, domain.ErrOrderNotFound),
		errors.Is(err, domain.ErrTicketNotFound),
		errors.Is(err, domain.ErrPaymentNotFound),
		errors.Is(err, domain.ErrMovieNotFound),
		errors.Is(err, domain.ErrHallNotFound),
		errors.Is(err, domain.ErrCinemaNotFound):
		Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrSeatNotAvailable),
		errors.Is(err, domain.ErrSeatLockConflict),
		errors.Is(err, domain.ErrSessionNotBookable),
		errors.Is(err, domain.ErrCouponNotAvailable),
		errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrTicketNotUsable),
		errors.Is(err, domain.ErrOrderExpired),
		errors.Is(err, domain.ErrPaymentAmountMismatch),
		errors.Is(err, domain.ErrSessionTimeConflict),
		errors.Is(err, domain.ErrSessionLockedForChange),
		errors.Is(err, domain.ErrOrderNotRefundable),
		errors.Is(err, domain.ErrRefundExists),
		errors.Is(err, domain.ErrChangeMovieMismatch):
		Fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrChangeSeatCount):
		Fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrMovieInvalid),
		errors.Is(err, domain.ErrHallInvalid),
		errors.Is(err, domain.ErrSeatLayoutInvalid),
		errors.Is(err, domain.ErrSessionInvalid),
		errors.Is(err, domain.ErrInvalidInput):
		Fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrUsernameTaken):
		Fail(c, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrEmailNotRegistered):
		Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrResetCodeInvalid):
		Fail(c, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		Fail(c, http.StatusUnauthorized, err.Error())
	case errors.Is(err, domain.ErrAccountLocked):
		Fail(c, http.StatusLocked, err.Error())
	case errors.Is(err, domain.ErrRefundNotFound):
		Fail(c, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrForbidden):
		Fail(c, http.StatusForbidden, err.Error())
	default:
		Fail(c, http.StatusInternalServerError, "internal error")
	}
}
