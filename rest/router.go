package rest

import (
	"net/http"
	"strconv"

	"github.com/ivgag/schedulr/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(userService *service.UserService) *gin.Engine {
	router := gin.Default()

	router.GET("/users/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		user, err := userService.GetUserByID(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	router.GET("/outh2callback", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		err := userService.ConnectGoogleAccount(state, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Account connected successfully"})
	})

	return router
}
