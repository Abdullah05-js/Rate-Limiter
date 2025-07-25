package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type Options struct {
	Redis      *redis.Client
	Context    context.Context
	QuotaMap   map[string]int
	UserIDKey  string // default: "userID"
	RankKey    string // default: "rank"
	Duration   time.Duration // default: 24 * time.Hour
}

func New(options Options) fiber.Handler {
	// defaults
	if options.UserIDKey == "" {
		options.UserIDKey = "userID"
	}
	if options.RankKey == "" {
		options.RankKey = "rank"
	}
	if options.Duration == 0 {
		options.Duration = 24 * time.Hour
	}

	return func(c *fiber.Ctx) error {
		userID, ok1 := c.Locals(options.UserIDKey).(string)
		rank, ok2 := c.Locals(options.RankKey).(string)

		if !ok1 || !ok2 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Missing userID or rank in Locals",
			})
		}

		key := "rate:" + userID

		exist, err := options.Redis.Exists(options.Context, key).Result()
		if err != nil {
			return fiber.NewError(500, err.Error())
		}

		if exist != 1 {
			limit, ok := options.QuotaMap[rank]
			if !ok {
				return fiber.NewError(fiber.StatusForbidden, "Invalid rank")
			}

			pipe := options.Redis.TxPipeline()
			pipe.Set(options.Context, key, limit, options.Duration)
			pipe.Decr(options.Context, key)
			if _, err := pipe.Exec(options.Context); err != nil {
				return fiber.NewError(500, err.Error())
			}
			return c.Next()
		}

		quotaStr, err := options.Redis.Get(options.Context, key).Result()
		if err != nil {
			return fiber.NewError(500, err.Error())
		}

		quota, err := strconv.Atoi(quotaStr)
		if err != nil {
			return fiber.NewError(500, err.Error())
		}

		if quota <= 0 {
			retryAfter := int(options.Duration.Seconds())
			c.Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"message": "You have exceeded your quota. Try again later.",
			})
		}

		if _, err := options.Redis.Decr(options.Context, key).Result(); err != nil {
			return fiber.NewError(500, err.Error())
		}

		return c.Next()
	}
}
