package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/reynaldineo/redis-go-url-shortener/database"
	"github.com/reynaldineo/redis-go-url-shortener/helpers"
)

type (
	request struct {
		URL         string        `json:"url"`
		CustomShort string        `json:"custom_short"`
		Expiry      time.Duration `json:"expiry"`
	}

	response struct {
		URL            string        `json:"url"`
		CustomShort    string        `json:"custom_short"`
		Expiry         time.Duration `json:"expiry"`
		XRateRemaining int           `json:"x-rate-remaining"`
		XRateLimitRest time.Duration `json:"x-rate-limit-rest"`
	}
)

func ShortenURL(c *fiber.Ctx) error {

	body := new(request)

	if err := c.BodyParser(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// implement rate limiting

	r2 := database.CreateClient(1)
	defer r2.Close()

	valIp, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*time.Minute).Err()
	} else {
		valInt, _ := strconv.Atoi(valIp)
		if valInt < 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":           "Rate limit exceeded",
				"rate_limit_rest": limit / time.Nanosecond / time.Minute,
			})
		}
		_ = r2.Decr(database.Ctx, c.IP()).Err()
	}

	// check if the input is valid URL

	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	// check for domain error

	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Domain is not allowed"})
	}

	// enforce https, SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	return c.Status(fiber.StatusOK).JSON(response{})
}
