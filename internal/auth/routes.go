package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/register", handler.Register)
		authRoutes.POST("/login", handler.Login)
		authRoutes.POST("/refresh", handler.Refresh)
		authRoutes.POST("/logout", handler.Logout)
		// Add other routes like /verify, /password/forgot, /password/reset here as we move forward sir.
	}
}
