package controllers

import (
	"github.com/erdemkosk/golang-twitter-timeline-service/internal/services"
	"github.com/gofiber/fiber/v2"
)

type TimelineController struct {
	timelineService services.TimelineService
}

func CreateTimelineController() *TimelineController {
	timelineService := services.CreateTimelineService()
	return &TimelineController{timelineService: *timelineService}
}

func (t TimelineController) GetTimelineByUserId(c *fiber.Ctx) error {
	id := c.Params("id")

	tweet, err := t.timelineService.GetTimelineByUserId(c.Context(), id)

	if err != nil {
		return c.JSON(fiber.Map{
			"code": fiber.ErrBadRequest.Code,
			"data": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code": 200,
		"data": tweet,
	})
}
