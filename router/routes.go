package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"todoapp-gin/todo"
)

func SetupRouter(db todo.AstraDB) *gin.Engine {
	router := gin.Default()

	// [ CORS ]
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS
	// TRY: Gin's CORS Middleware - github.com/gin-contrib/cors
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Add("access-control-allow-origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Add("access-control-allow-methods", "GET,HEAD,POST,DELETE,OPTIONS,PUT,PATCH")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// [ /todos ]
	router.GET("/api/v1/:user_id/todos", func(c *gin.Context) {
		userID := c.Params.ByName("user_id")
		todoItems, err := db.All(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, todoItems)
	})

	router.POST("/api/v1/:user_id/todos", func(c *gin.Context) {
		var tdi todo.TodoItem
		if err := c.ShouldBindJSON(&tdi); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := c.Params.ByName("user_id")
		urlFunc := func(itemIDString string) string {
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-Host
			return fmt.Sprintf("https://%s/api/v1/%s/todos/%s", c.Request.Header.Get("X-Forwarded-Host"), userID, itemIDString)
		}
		todoItem, err := db.Create(userID, tdi, urlFunc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, todoItem)
	})

	router.DELETE("/api/v1/:user_id/todos", func(c *gin.Context) {
		userID := c.Params.ByName("user_id")
		if err := db.DeleteAll(userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "")
	})

	// [ /todos/:id ]
	router.GET("/api/v1/:user_id/todos/:item_id", func(c *gin.Context) {
		userID := c.Params.ByName("user_id")
		itemID := c.Params.ByName("item_id")
		todoItem, err := db.Retrieve(userID, itemID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, todoItem)
	})

	router.PATCH("/api/v1/:user_id/todos/:item_id", func(c *gin.Context) {
		tdiPatchMap := make(map[string]interface{})
		if err := c.ShouldBindJSON(&tdiPatchMap); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := c.Params.ByName("user_id")
		itemID := c.Params.ByName("item_id")
		todoItem, err := db.Update(userID, itemID, tdiPatchMap)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, todoItem)
	})

	router.DELETE("/api/v1/:user_id/todos/:item_id", func(c *gin.Context) {
		userID := c.Params.ByName("user_id")
		itemID := c.Params.ByName("item_id")
		if err := db.Delete(userID, itemID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "")
	})
	return router
}
