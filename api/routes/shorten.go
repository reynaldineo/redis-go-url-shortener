package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

	var id string

	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	val, _ := r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Custom short URL already exists",
		})
	}

	if body.Expiry == 0 {
		body.Expiry = 24 * time.Hour
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save URL",
		})
	}

	valIp, _ = r.Get(database.Ctx, c.IP()).Result()
	rateRemainInt, _ := strconv.Atoi(valIp)

	valTTL, _ := r.TTL(database.Ctx, c.IP()).Result()
	RateLimitReset := valTTL / time.Nanosecond / time.Minute

	customShort := os.Getenv("DOMAIN") + "/" + id

	resp := response{
		URL:            body.URL,
		CustomShort:    customShort,
		Expiry:         body.Expiry,
		XRateRemaining: rateRemainInt | 10,
		XRateLimitRest: RateLimitReset | 30,
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
