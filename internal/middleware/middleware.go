package middleware

import (
	"ai-learn-english/pkg/logger"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
)

// ConnectionLimiter limits the number of concurrent connections
type ConnectionLimiter struct {
	limit    int
	waitlist chan struct{}
}

func NewConnectionLimiter(limit int) *ConnectionLimiter {
	return &ConnectionLimiter{
		limit:    limit,
		waitlist: make(chan struct{}, limit),
	}
}

func (cl *ConnectionLimiter) Acquire() bool {
	select {
	case cl.waitlist <- struct{}{}:
		return true
	default:
		return false
	}
}

func (cl *ConnectionLimiter) Release() {
	select {
	case <-cl.waitlist:
	default:
	}
}

// connectionLimiterMiddleware creates a middleware for connection limiting
func connectionLimiterMiddleware(limiter *ConnectionLimiter) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !limiter.Acquire() {
			return c.Status(fiber.StatusServiceUnavailable).SendString("Server is at maximum capacity")
		}
		defer limiter.Release()
		return c.Next()
	}
}

// panicRecoveryMiddleware creates a middleware for panic recovery
func panicRecoveryMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				stack := debug.Stack()
				logger.WithFields(map[string]interface{}{
					"panic":      r,
					"method":     c.Method(),
					"path":       c.Path(),
					"ip":         c.IP(),
					"user_agent": c.Get("User-Agent"),
					"stack":      string(stack),
				}).Errorf("Panic recovered")

				// Return 500 Internal Server Error
				err := c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Internal Server Error",
					"message": "An unexpected error occurred",
				})
				if err != nil {
					logger.WithField("error", err).Errorf("Failed to send error response")
				}
			}
		}()
		return c.Next()
	}
}
