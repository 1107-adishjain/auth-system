package auth

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    service         Service
    accessTokenTTL  int
    refreshTokenTTL int
}

func NewAuthHandler(s Service, accessTTL, refreshTTL int) *AuthHandler {
    return &AuthHandler{
        service:         s,
        accessTokenTTL:  accessTTL,
        refreshTokenTTL: refreshTTL,
    }
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
        return
    }

    if req.Email == "" || req.Password == "" {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Email and password are required"})
        return
    }

    user, err := h.service.Register(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        if err == ErrEmailTaken {
            c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
        return
    }
    // The response struct has been updated to use uint for ID
    c.JSON(http.StatusCreated, UserResponse{
        ID:        user.ID,
        Email:     user.Email,
        CreatedAt: user.CreatedAt,
    })
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
        return
    }

    accessToken, refreshToken, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
        return
    }

    c.SetCookie("access_token", accessToken, h.accessTokenTTL*60, "/", "", true, true)
    c.SetCookie("refresh_token", refreshToken, h.refreshTokenTTL*60, "/auth", "", true, true)

    c.JSON(http.StatusOK, LoginResponse{Message: "Logged in successfully"})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
    refreshToken, err := c.Cookie("refresh_token")
    if err != nil {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Refresh token not found"})
        return
    }

    newAccessToken, newRefreshToken, err := h.service.Refresh(c.Request.Context(), refreshToken)
    if err != nil {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired refresh token"})
        return
    }

    c.SetCookie("access_token", newAccessToken, h.accessTokenTTL*60, "/", "", true, true)
    c.SetCookie("refresh_token", newRefreshToken, h.refreshTokenTTL*60, "/auth", "", true, true)

    c.JSON(http.StatusOK, LoginResponse{Message: "Tokens refreshed successfully"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
    refreshToken, err := c.Cookie("refresh_token")
    if err != nil {
        c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "No active session to log out"})
        return
    }

    if err := h.service.Logout(c.Request.Context(), refreshToken); err != nil {
        c.SetCookie("access_token", "", -1, "/", "", true, true)
        c.SetCookie("refresh_token", "", -1, "/auth", "", true, true)
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to log out"})
        return
    }

    c.SetCookie("access_token", "", -1, "/", "", true, true)
    c.SetCookie("refresh_token", "", -1, "/auth", "", true, true)

    c.JSON(http.StatusOK, LoginResponse{Message: "Logged out successfully"})
}
