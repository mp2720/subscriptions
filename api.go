package main

import (
	"database/sql"
	"errors"
	"fmt"
	"mp2720/subscriptions/sqlgen"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type API struct {
	Queries *sqlgen.Queries
	Log     *Logger
}

func (api *API) RegisterHandlers(r *gin.RouterGroup) {
	r.POST("/subscriptions", api.createSubscription)
	r.GET("/subscriptions/:id", api.getSubscriptionByID)
	r.GET("/subscriptions", api.getAllSubscriptions)
	r.GET("/subscriptions/stats", api.getSubscriptionsStats)
	r.POST("/subscriptions/:id/cancel", api.cancelSubscriptionByID)
}

type CreateSubscriptionReq struct {
	ServiceName string     `json:"service_name"`
	Price       uint       `json:"price"`
	UserUUID    uuid.UUID  `json:"user_uuid" example:"cd2a0341-669e-48ca-9210-839b943e75ae"`
	StartDate   *MonthYear `json:"start_date" swaggertype:"string" example:"02-2026"`
	EndDate     *MonthYear `json:"end_date" swaggertype:"string" example:"04-2026"`
}

type SubscriptionResp struct {
	Id              int64      `json:"id"`
	ServiceName     string     `json:"service_name"`
	Price           uint       `json:"price"`
	UserUUID        uuid.UUID  `json:"user_uuid" example:"cd2a0341-669e-48ca-9210-839b943e75ae"`
	StartDate       MonthYear  `json:"start_date" swaggertype:"string" example:"02-2026"`
	EndDate         *MonthYear `json:"end_date" swaggertype:"string" example:"04-2026"`
	CancelationDate *MonthYear `json:"cancelation_date" swaggertype:"string" example:"03-2026"`
}

func subscriptionRespFromSQL(sub *sqlgen.Subscription) SubscriptionResp {
	var endDate *MonthYear
	if sub.EndDate.Valid {
		endDate = &MonthYear{sub.EndDate.Time}
	}
	var cancelationDate *MonthYear
	if sub.CancelationDate.Valid {
		cancelationDate = &MonthYear{sub.CancelationDate.Time}
	}
	return SubscriptionResp{
		Id:              sub.ID.Int64,
		ServiceName:     sub.ServiceName,
		Price:           uint(sub.Price),
		UserUUID:        uuid.UUID(sub.UserUuid),
		StartDate:       TruncateTimeToMonth(sub.StartDate),
		EndDate:         endDate,
		CancelationDate: cancelationDate,
	}
}

// createSubscription godoc
//
//	@Summary	Create subscription
//	@Schemes
//	@Produce	json
//	@Param		subscription	body		CreateSubscriptionReq	true	"subscription"
//	@Success	201				{object}	SubscriptionResp		"created"
//	@Failure	400				{object}	HTTPErrorResp			"Bad request"
//	@Router		/subscriptions [post]
func (api *API) createSubscription(c *gin.Context) {
	var req CreateSubscriptionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		respError(c,
			http.StatusBadRequest,
			fmt.Sprintf("Invalid body: %s", err),
		)
		return
	}

	currentMonthYear := TruncateTimeToMonth(time.Now())

	if req.StartDate == nil {
		req.StartDate = &currentMonthYear
	}
	var endDateSQL sql.NullTime
	if req.EndDate != nil {
		endDateSQL.Time = req.EndDate.Time
		endDateSQL.Valid = true
	}

	if req.StartDate.Before(currentMonthYear.Time) {
		respError(c,
			http.StatusBadRequest,
			"Invalid subscription period: start date is from past",
		)
		return
	}
	if req.EndDate != nil && req.EndDate.Before(req.StartDate.Time) {
		respError(c,
			http.StatusBadRequest,
			"Invalid subscription period: end date is before the start date",
		)
		return
	}

	createdSub, err := api.Queries.CreateSubscription(c.Request.Context(), sqlgen.CreateSubscriptionParams{
		ServiceName: req.ServiceName,
		Price:       int32(req.Price),
		UserUuid:    req.UserUUID[:],
		StartDate:   req.StartDate.Time,
		EndDate:     endDateSQL,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	api.Log.Verbosef("subscription created with id=%d", createdSub.ID.Int64)

	c.JSON(http.StatusCreated, subscriptionRespFromSQL(&createdSub))
}

// getSubscriptionByID godoc
//
//	@Summary	Get subscription by ID
//	@Schemes
//	@Produce	json
//	@Param		id	path		int					true	"subscription id"
//	@Success	200	{object}	SubscriptionResp	"ok"
//	@Failure	400	{object}	HTTPErrorResp		"Bad request"
//	@Failure	404	{object}	HTTPErrorResp		"Not found"
//	@Router		/subscriptions/{id} [get]
func (api *API) getSubscriptionByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respError(c,
			http.StatusBadRequest,
			fmt.Sprintf("Invalid ID: %s", err),
		)
		return
	}

	sub, err := api.Queries.GetSubscriptionById(
		c.Request.Context(),
		sql.NullInt64{Int64: id, Valid: true},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respError(c,
				http.StatusNotFound,
				"Subscription not found",
			)
		} else {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		return
	}

	c.JSON(http.StatusOK, subscriptionRespFromSQL(&sub))
}

// getAllSubscriptions godoc
//
//	@Summary	Get all subscriptions
//	@Schemes
//	@Produce	json
//	@Param		user-uuid		query		string				false	"user UUID"
//	@Param		service-name	query		string				false	"service name"
//	@Success	200				{object}	[]SubscriptionResp	"ok"
//	@Failure	400				{object}	HTTPErrorResp		"Bad request"
//	@Router		/subscriptions [get]
func (api *API) getAllSubscriptions(c *gin.Context) {
	var serviceName sql.NullString
	if serviceNameStr, present := c.GetQuery("service-name"); present {
		serviceName = sql.NullString{String: serviceNameStr, Valid: true}
	}
	var userUUIDSlice []byte
	if _, present := c.GetQuery("user-uuid"); present {
		userUUID, err := uuid.Parse(c.Query("user-uuid"))
		if err != nil {
			respError(c,
				http.StatusBadRequest,
				"Invalid user UUID",
			)
			return
		}
		userUUIDSlice = userUUID[:]
	}

	subs, err := api.Queries.GetAllSubscriptions(
		c.Request.Context(),
		sqlgen.GetAllSubscriptionsParams{
			ServiceName: serviceName,
			UserUuid:    userUUIDSlice,
		},
	)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	subsResp := make([]SubscriptionResp, 0, len(subs))
	for _, sub := range subs {
		subsResp = append(subsResp, subscriptionRespFromSQL(&sub))
	}

	c.JSON(http.StatusOK, subsResp)
}

type SubscriptionsStatsResp struct {
	Revenue int `json:"revenue"`
}

// getSubscriptionsStats godoc
//
//	@Summary	Get subscriptions stats
//	@Schemes
//	@Description	Get stats for period (which is a closed interval), user and service filters
//	@Produce		json
//	@Param			user-uuid		query		string					false	"user UUID"
//	@Param			service-name	query		string					false	"service name"
//	@Param			period-start	query		string					true	"period start"
//	@Param			period-end		query		string					true	"period end"
//	@Success		200				{object}	SubscriptionsStatsResp	"ok"
//	@Failure		400				{object}	HTTPErrorResp			"Bad request"
//	@Router			/subscriptions/stats [get]
func (api *API) getSubscriptionsStats(c *gin.Context) {
	var serviceName sql.NullString
	if serviceNameStr, present := c.GetQuery("service-name"); present {
		serviceName = sql.NullString{String: serviceNameStr, Valid: true}
	}
	var userUUIDSlice []byte
	if _, present := c.GetQuery("user-uuid"); present {
		userUUID, err := uuid.Parse(c.Query("user-uuid"))
		if err != nil {
			respError(c,
				http.StatusBadRequest,
				"Invalid user UUID",
			)
			return
		}
		userUUIDSlice = userUUID[:]
	}
	periodStart := MonthYear{}
	if err := periodStart.Parse(c.Query("period-start")); err != nil {
		respError(c,
			http.StatusBadRequest,
			"Invalid period start",
		)
		return
	}
	periodEnd := MonthYear{}
	if err := periodEnd.Parse(c.Query("period-end")); err != nil {
		respError(c,
			http.StatusBadRequest,
			"Invalid period end",
		)
		return
	}

	if periodEnd.Before(periodStart.Time) {
		respError(c,
			http.StatusBadRequest,
			"Invalid period",
		)
		return
	}

	revenue, err := api.Queries.CalculateSubscriptionsRevenue(
		c.Request.Context(),
		sqlgen.CalculateSubscriptionsRevenueParams{
			PeriodStart: sql.NullTime{Time: periodStart.Time, Valid: true},
			PeriodEnd:   sql.NullTime{Time: periodEnd.Time, Valid: true},
			ServiceName: serviceName,
			UserUuid:    userUUIDSlice,
		},
	)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, SubscriptionsStatsResp{int(revenue)})
}

// cancelSubscriptionByID godoc
//
//	@Summary	Cancel subscription
//	@Schemes
//	@Produce	json
//	@Param		id	path	int	true	"subscription id"
//	@Success	204	"ok"
//	@Failure	404	{object}	HTTPErrorResp	"No subscription that is not ended yet with such ID"
//	@Router		/subscriptions/{id}/cancel [post]
func (api *API) cancelSubscriptionByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		respError(c,
			http.StatusBadRequest,
			fmt.Sprintf("Invalid ID: %s", err),
		)
		return
	}

	currentMonth := TruncateTimeToMonth(time.Now())

	rows, err := api.Queries.CancelSubscriptionByID(
		c.Request.Context(),
		sqlgen.CancelSubscriptionByIDParams{
			ID:              sql.NullInt64{Int64: id, Valid: true},
			CancelationDate: sql.NullTime{Time: currentMonth.Time, Valid: true},
		},
	)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if rows == 0 {
		respError(c,
			http.StatusNotFound,
			"Subscription not found",
		)
	}

	api.Log.Verbosef("subscription with id=%d canceled", id)

	c.Status(http.StatusNoContent)
}
