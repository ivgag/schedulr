package rest

import (
	"net/http"

	"github.com/ivgag/schedulr/model"
	"github.com/ivgag/schedulr/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(userService *service.UserService) *gin.Engine {
	router := gin.Default()

	router.GET("/oauth2callback/google", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		err := userService.LinkAccount(state, model.ProviderGoogle, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "https://t.me/schedulr_bot")
	})

	return router
}
