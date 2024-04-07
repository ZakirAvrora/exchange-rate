package v1

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/ZakirAvrora/exchange-rate/internal/exchangerates"
	"github.com/gin-gonic/gin"
)

type translationRoutes struct {
	t exchangerates.RecodsService
}

func newExchangeRatesRoutes(handler *gin.RouterGroup, t exchangerates.RecodsService) {
	r := &translationRoutes{t}

	h := handler.Group("/exchangerates")
	{
		h.GET("/:id", r.fetchByID)
		h.POST("/refresh", r.refresh)
		h.GET("/latest", r.latest)
	}
}

type exchangeResponse struct {
	Rate       float64   `json:"rate"`
	UpdateTime time.Time `json:"update_time"`
}

// @Summary     Getting rate by identifier
// @Description	Display exchange rate value and update time for corresponding identifier request
// @ID          get-rate-by-identifier
// @Tags  	    exchangerates
// @Accept      json
// @Param 		id path string true "unique identifier" Format(uuid)
// @Success     200 {object} exchangeResponse
// @Failure     404 {object} response
// @Failure     500 {object} response
// @Router      /exchangerates/{id} [get]
func (r *translationRoutes) fetchByID(c *gin.Context) {
	id := c.Param("id")
	record, err := r.t.FetchByIdentifier(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, exchangerates.ErrNoRecord) {
			errorResponse(c, http.StatusNotFound, "exchange rate for identifier was not found")
		} else {

			errorResponse(c, http.StatusInternalServerError, "we are experiecning internal server errors, retry later")
		}
		return
	}

	c.JSON(http.StatusOK, exchangeResponse{Rate: record.Rate, UpdateTime: record.Updated_At})
}

type doRefreshRequest struct {
	Base      string `json:"base"       binding:"required"  example:"EUR"`
	Secondary string `json:"secondary"  binding:"required"  example:"MXN"`
}

type identiferResponse struct {
	Identifier string `json:"identifier"`
}

// @Summary     Update exchange rate
// @Description The service assigns an identifier to the update request.
// @Description The service updates quotes in the background, i.e. the request handler do not perform the update.
// @ID          update-exchange-rate
// @Tags  	    exchangerates
// @Accept      json
// @Produce     json
// @Param       request body doRefreshRequest true "Set up currency pair"
// @Success     200 {object} identiferResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /exchangerates/refresh [post]
func (r *translationRoutes) refresh(c *gin.Context) {
	var request doRefreshRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Println("http - v1 - latest", err)
		errorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	identifier, err := r.t.Refresh(
		c.Request.Context(),
		request.Base,
		request.Secondary,
	)
	if err != nil {
		processError(c, err)
		return
	}

	c.JSON(http.StatusCreated, identiferResponse{identifier})
}

// @Summary     Getting latest exchange rate for currency pair
// @Description The request specifies the currency pair code.
// @Description In the response, the service provides the price value and update time.
// @ID          get-latest-rate
// @Tags  	    exchangerates
// @Accept      json
// @Param       base query string true "first currency code of pair" Enums(EUR)
// @Param       secondary query string true "second currency code of pair" Enums(BTC, MXN, USD, BYR, AED, KZT, RUB, XAU, XAG, LYD)
// @Success     200 {object} exchangeResponse
// @Failure     400 {object} response
// @Failure     404 {object} response
// @Failure     500 {object} response
// @Router      /exchangerates/latest [get]
func (r *translationRoutes) latest(c *gin.Context) {
	base := c.Query("base")
	secondary := c.Query("secondary")

	record, err := r.t.FetchLatest(
		c.Request.Context(),
		base,
		secondary,
	)

	if err != nil {
		processError(c, err)
		return
	}

	c.JSON(http.StatusOK, exchangeResponse{
		Rate:       record.Rate,
		UpdateTime: record.Updated_At,
	})
}
