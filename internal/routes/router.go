package router

import (
	"github.com/erdemkosk/golang-twitter-timeline-service/internal/controllers"
	"github.com/gofiber/fiber/v2"
)

func Initalize(router *fiber.App) {

	timelineController := controllers.CreateTimelineController()

	router.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).SendString("Hello, World!")
	})

	timelines := router.Group("/timelines")
	timelines.Get("user/:id", timelineController.GetTimelineByUserId)

	router.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "404: Not Found",
		})
	})

}
