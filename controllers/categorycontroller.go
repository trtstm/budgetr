package controllers

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/trtstm/budgetr/db"
	"github.com/trtstm/budgetr/log"
	"github.com/trtstm/budgetr/models"
)

type categoryController struct {
}

func (c *categoryController) Index(ctx echo.Context) error {
	categories := []*models.Category{}

	if q := db.DB.Find(&categories); q.Error != nil {
		log.Errorf("CategoryController::Index Could not execute find query: %v", q.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	log.Infof("CategoryController::Index Returning %d categories.", len(categories))
	return ctx.JSON(http.StatusOK, echo.Map{
		"data": TransformCategory(categories...),
	})
}

// CategoryController for /categories endpoint.
var CategoryController categoryController
