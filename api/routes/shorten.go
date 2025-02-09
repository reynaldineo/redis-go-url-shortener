package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
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
