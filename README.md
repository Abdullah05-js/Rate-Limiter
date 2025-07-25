# 🔒 Fiber RateLimiter Middleware (per User Rank)

A flexible Fiber middleware for per-user rate limiting based on **user rank**, using Redis.  
Easily plug into your Fiber app and control API usage based on user tiers like `free`, `pro`, `admin`.

---

## 📦 Installation

```bash
go get github.com/kullaniciadi/rate-limiter/ratelimiter
```

---

## 🚀 Usage

```go
package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/kullaniciadi/rate-limiter/ratelimiter"
)

func main() {
	app := fiber.New()

	// Redis connection
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Quota per user rank
	quotaMap := map[string]int{
		"free":  50,
		"pro":   200,
		"admin": 1000,
	}

	// Apply the rate limiter middleware
	app.Use(ratelimiter.New(ratelimiter.Options{
		Redis:    rdb,
		Context:  context.Background(),
		QuotaMap: quotaMap,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome! ✅")
	})

	app.Listen(":3000")
}
```

---

## ⚙️ Middleware Behavior

- Rate limits are applied **per user**, using `userID` from `c.Locals("userID")`.
- Quotas are determined based on user `rank` from `c.Locals("rank")`.
- If the user's quota is exceeded:
  - Returns `429 Too Many Requests`
  - Sets `Retry-After` header (in seconds)
  - Response body:
    ```json
    {
      "message": "You have exceeded your quota. Try again later."
    }
    ```

---

## 🔧 Options

| Field        | Type               | Default          | Description                                  |
|--------------|--------------------|------------------|----------------------------------------------|
| `Redis`      | `*redis.Client`    | —                | Required Redis client                        |
| `Context`    | `context.Context`  | —                | Required context                             |
| `QuotaMap`   | `map[string]int`   | —                | Quotas by rank (e.g., `"free": 50`)          |
| `UserIDKey`  | `string`           | `"userID"`       | Locals key name for user ID                  |
| `RankKey`    | `string`           | `"rank"`         | Locals key name for rank                     |
| `Duration`   | `time.Duration`    | `24 * time.Hour` | How long the quota is valid (reset interval) |

---

## 📌 Example Response on Limit Exceed

```http
HTTP/1.1 429 Too Many Requests
Retry-After: 3600
Content-Type: application/json

{
  "message": "You have exceeded your quota. Try again later."
}
```

---

## 🧪 Testing

Manually test by setting `c.Locals("userID")` and `c.Locals("rank")` in a dummy middleware before this one.

```go
app.Use(func(c *fiber.Ctx) error {
	c.Locals("userID", "user123")
	c.Locals("rank", "free")
	return c.Next()
})
```

---

## 💬 Contributing

PRs and issues welcome!

---

## 📄 License

MIT
