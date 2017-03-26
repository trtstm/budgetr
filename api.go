package main

import (
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/trtstm/budgetr/config"
	"github.com/trtstm/budgetr/controllers"
)

func startAPI() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	e.Use(middleware.GzipWithConfig(middleware.DefaultGzipConfig))

	e.Static("/", "web/dist")

	// Restricted group
	r := e.Group("/api")

	r.GET("/categories", controllers.CategoryController.Index)

	r.GET("/expenditures", controllers.ExpenditureController.Index)
	r.GET("/expenditures/:id", controllers.ExpenditureController.Show)
	r.POST("/expenditures/:id", controllers.ExpenditureController.Update)
	r.DELETE("/expenditures/:id", controllers.ExpenditureController.Delete)
	r.POST("/expenditures", controllers.ExpenditureController.Create)

	r.GET("/stats/categories", controllers.CategoryStatsController.Index)

	r.POST("/exports/excel", controllers.ExportController.ExportExcel)

	e.Logger.Fatal(e.Start(config.Config.Hostname + ":" + strconv.Itoa(int(config.Config.Port))))
}
