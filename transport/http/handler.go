package http

import (
	"net/http"

	"github.com/alexliesenfeld/health"
	"github.com/gin-gonic/gin"
	"github.com/mgjules/minion/docs"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// Success defines the structure for a successful response.
type Success struct {
	Message string `json:"message"`
}

// handleHealthCheck godoc
// @Summary      Health Check
// @Description  checks if server is running
// @Tags         core
// @Produce      json
// @Success      200  {object}	health.CheckerResult
// @Success      503  {object}  health.CheckerResult
// @Router       / [get]
func (s *Server) handleHealthCheck() gin.HandlerFunc {
	opts := s.health.CompileHealthCheckerOption()
	checker := health.NewChecker(opts...)

	return gin.WrapF(
		health.NewHandler(
			checker,
		),
	)
}

// handleVersion godoc
// @Summary      Health Check
// @Description  checks the server's version
// @Tags         core
// @Produce      json
// @Success      200  {object}  build.Info
// @Router       /version [get]
func (s *Server) handleVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, s.build)
	}
}

// handleMinion  godoc
// @Summary      Minion introduction
// @Description  returns the minion's introduction
// @Tags         core
// @Produce      json
// @Success      200  {string}	string 	"My name is '{name}' and I have a secret key '{key}'."
// @Router       /api/minion [get]
func (s *Server) handleMinion() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, s.minion.String())
	}
}

func (Server) handleSwagger() gin.HandlerFunc {
	docs.SwaggerInfo.BasePath = "/"

	url := ginSwagger.URL("/swagger/doc.json")

	return ginSwagger.WrapHandler(swaggerFiles.Handler, url)
}
