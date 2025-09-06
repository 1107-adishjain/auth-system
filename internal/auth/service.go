package auth

import (
    "context"
    "errors"
    "fmt"
    // "strconv"
    "time"

    "github.com/1107-adishjain/auth-system/internal/models"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
)

var (
    ErrUserNotFound        = errors.New("user not found")
    ErrInvalidPassword     = errors.New("invalid password")
    ErrEmailTaken          = errors.New("email is already taken")
    ErrTokenInvalid        = errors.New("token is invalid")
    ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")
)

type Service interface {
    Register(ctx context.Context, email, password string) (*models.User, error)
    Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
    Refresh(ctx context.Context, oldRefreshToken string) (newAccessToken string, newRefreshToken string, err error)
    Logout(ctx context.Context, refreshToken string) error
}

type service struct {
    repo            Repository
    jwtSecret       string
    accessTokenTTL  time.Duration
    refreshTokenTTL time.Duration
}

func NewService(repo Repository, jwtSecret string, accessTTL, refreshTTL int) Service {
    return &service{
        repo:            repo,
        jwtSecret:       jwtSecret,
        accessTokenTTL:  time.Duration(accessTTL) * time.Minute,
        refreshTokenTTL: time.Duration(refreshTTL) * time.Minute,
    }
}

func (s *service) Register(ctx context.Context, email, password string) (*models.User, error) {
    _, err := s.repo.GetUserByEmail(ctx, email)
    if err == nil {
        return nil, ErrEmailTaken
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("could not hash password: %w", err)
    }

    user := &models.User{
        Email:        email,
        PasswordHash: string(hashedPassword),
    }

    if err := s.repo.CreateUser(ctx, user); err != nil {
        return nil, fmt.Errorf("could not create user: %w", err)
    }
    return user, nil
}

func (s *service) Login(ctx context.Context, email, password string) (string, string, error) {
    user, err := s.repo.GetUserByEmail(ctx, email)
    if err != nil {
        return "", "", ErrUserNotFound
    }

    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
    if err != nil {
        return "", "", ErrInvalidPassword
    }

    accessToken, err := s.createAccessToken(user)
    if err != nil {
        return "", "", err
    }

    refreshToken, err := s.createRefreshToken(ctx, user)
    if err != nil {
        return "", "", err
    }

    return accessToken, refreshToken, nil
}

func (s *service) Refresh(ctx context.Context, oldRefreshToken string) (string, string, error) {
    claims, err := s.validateToken(oldRefreshToken)
    if err != nil {
        return "", "", err
    }

    userID, err := s.repo.GetRefreshTokenUserID(ctx, claims.ID)
    if err != nil || fmt.Sprint(userID) != claims.Subject {
        return "", "", ErrRefreshTokenRevoked
    }

    if err := s.repo.RevokeRefreshToken(ctx, claims.ID); err != nil {
        return "", "", err
    }

    user := &models.User{ID: userID}

    newAccessToken, err := s.createAccessToken(user)
    if err != nil {
        return "", "", err
    }

    newRefreshToken, err := s.createRefreshToken(ctx, user)
    if err != nil {
        return "", "", err
    }

    return newAccessToken, newRefreshToken, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
    claims, err := s.validateToken(refreshToken)
    if err != nil {
        return ErrTokenInvalid
    }
    return s.repo.RevokeRefreshToken(ctx, claims.ID)
}

func (s *service) createAccessToken(user *models.User) (string, error) {
    claims := &jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
        Subject:   fmt.Sprint(user.ID),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}

func (s *service) createRefreshToken(ctx context.Context, user *models.User) (string, error) {
    tokenID := uuid.New().String()
    claims := &jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenTTL)),
        Subject:   fmt.Sprint(user.ID),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        ID:        tokenID,
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", err
    }

    err = s.repo.StoreRefreshToken(ctx, user.ID, tokenID, s.refreshTokenTTL)
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

func (s *service) validateToken(tokenString string) (*jwt.RegisteredClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(s.jwtSecret), nil
    })

    if err != nil {
        return nil, ErrTokenInvalid
    }

    claims, ok := token.Claims.(*jwt.RegisteredClaims)
    if !ok || !token.Valid {
        return nil, ErrTokenInvalid
    }

    return claims, nil
}
