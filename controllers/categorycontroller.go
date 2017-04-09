package controllers

import (
	"net/http"
	"strconv"
	"strings"

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

func (c *categoryController) Update(ctx echo.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		log.Infof("categoryController::Update Could not parse id `%s`: '%v'.", ctx.Param("id"), err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	params := &struct {
		Name string `json:"name" form:"name"`
	}{}

	if err := ctx.Bind(params); err != nil {
		log.Infof("categoryController::Update Could not bind params: '%v'.", err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	category := &models.Category{}

	if q := db.DB.Where("id = ?", id).First(category); q.Error != nil {
		log.Errorf("CategoryController::Update Could not execute query: %v", q.Error)
		return ctx.NoContent(http.StatusNotFound)
	}

	oldName := category.Name
	category.Name = strings.TrimSpace(params.Name)

	if len(category.Name) == 0 {
		log.Infof("categoryController::Update Name cant be empty.")
		return ctx.NoContent(http.StatusBadRequest)
	}

	if q := db.DB.Save(category); q.Error != nil {
		log.Errorf("CategoryController::Update Could not save category: %v", q.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	log.Infof("CategoryController::Update Changed name from '%s' to '%s'.", oldName, category.Name)
	return ctx.JSON(http.StatusOK, TransformCategory(category)[0])
}

// CategoryController for /categories endpoint.
var CategoryController categoryController
