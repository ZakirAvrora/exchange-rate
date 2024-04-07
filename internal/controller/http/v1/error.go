package v1

import (
	"errors"
	"net/http"

	"github.com/ZakirAvrora/exchange-rate/internal/exchangerates"
	"github.com/gin-gonic/gin"
)

type response struct {
	Error string `json:"error" example:"error message"`
}

func processError(c *gin.Context, err error) {
	if errors.Is(err, exchangerates.ErrNoRecord) {
		errorResponse(c, http.StatusNotFound, "exchange rate for pair was not found")
	} else if errors.Is(err, exchangerates.ErrNotSupportedBaseCurrency) {
		errorResponse(c, http.StatusBadRequest, "not supported base currency")
	} else if errors.Is(err, exchangerates.ErrNotSupportedSecondaryCurrency) {
		errorResponse(c, http.StatusBadRequest, "not supported secondary currency")
	} else {
		errorResponse(c, http.StatusInternalServerError, "we are experiecning internal server errors, retry later")
	}
}

func errorResponse(c *gin.Context, code int, msg string) {
	c.AbortWithStatusJSON(code, response{msg})
}
