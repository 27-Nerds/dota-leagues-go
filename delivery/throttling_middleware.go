package delivery

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
)

var (
	store       throttled.GCRAStore
	quota       throttled.RateQuota
	rateLimiter *throttled.GCRARateLimiter
)

// if throttled is initialized in init, it will be global for all the requests that are  using it
// if we need individual throttling for each route, we can move this code into IPRateLimit func
// also this logic can be moved from Group Level middleware to Root Level
// https://echo.labstack.com/middleware
// the good idea is to move from memstore to redisstore
// https://github.com/throttled/throttled/blob/master/store/redigostore/redigostore.go
func init() {
	log.Println("init")
	store, err := memstore.New(65536)
	if err != nil {
		log.Panicf("IPRateLimit - ipRateLimiter.Get - err: %v, ", err)
		//return nil
	}

	quota = throttled.RateQuota{
		MaxRate:  throttled.PerMin(100),
		MaxBurst: 20,
	}
	rateLimiter, err = throttled.NewGCRARateLimiter(store, quota)
	if err != nil {
		log.Panicf("IPRateLimit - ipRateLimiter.Get - err: %v, ", err)
		//	return nil
	}

}

// IPRateLimit rate limiting middleware
func IPRateLimit() echo.MiddlewareFunc {

	// Return middleware handler
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ip := c.RealIP()

			isLimited, RateLimitResult, err := rateLimiter.RateLimit(ip, 1)
			if err != nil {
				log.Printf("IPRateLimit - ipRateLimiter.Get - err: %v, %s on %s", err, ip, c.Request().URL)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"success": false,
					"message": err,
				})
			}

			h := c.Response().Header()
			h.Set("X-RateLimit-Limit", strconv.Itoa(RateLimitResult.Limit))
			h.Set("X-RateLimit-Remaining", strconv.Itoa(RateLimitResult.Remaining))
			h.Set("X-RateLimit-Reset", strconv.Itoa(int(RateLimitResult.ResetAfter.Milliseconds()/1000)))

			if isLimited {
				log.Printf("Too Many Requests from %s on %s", ip, c.Request().URL)
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"success": false,
					"message": "Too Many Requests on " + c.Request().URL.String(),
				})
			}

			return next(c)
		}
	}
}
