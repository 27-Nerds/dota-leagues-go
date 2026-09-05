package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
)

// IPRateLimit rate limiting with default config
func IPRateLimit() echo.MiddlewareFunc {
	return IPRateLimitWithConfig(100, 20)
}

// IPRateLimitWithConfig rate limiting middleware with config
func IPRateLimitWithConfig(perMin int, burst int) echo.MiddlewareFunc {

	store, err := memstore.NewCtx(65536)
	if err != nil {
		panic(fmt.Errorf("create rate limit store: %w", err))
	}

	quota := throttled.RateQuota{
		MaxRate:  throttled.PerMin(perMin),
		MaxBurst: burst,
	}
	rateLimiter, err := throttled.NewGCRARateLimiterCtx(store, quota)
	if err != nil {
		panic(fmt.Errorf("create rate limiter: %w", err))
	}

	// Return middleware handler
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			// Echo's root static-file route serves logos and frontend assets.
			// These requests must not consume the API's per-IP quota. Match the
			// registered route, so API parameters ending in .png remain limited.
			if c.Path() == "/*" && c.Request().Method == http.MethodGet {
				return next(c)
			}

			ip := c.RealIP()

			isLimited, RateLimitResult, err := rateLimiter.RateLimitCtx(c.Request().Context(), ip, 1)
			if err != nil {
				slog.ErrorContext(c.Request().Context(), "check rate limit", "ip", ip, "path", c.Path(), "error", err)
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
				slog.WarnContext(c.Request().Context(), "rate limit exceeded", "ip", ip, "path", c.Path())
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"success": false,
					"message": "Too Many Requests on " + c.Request().URL.String(),
				})
			}

			return next(c)
		}
	}
}
