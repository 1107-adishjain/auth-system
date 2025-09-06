package auth

import (
    "context"
    "fmt"
    "strconv"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/1107-adishjain/auth-system/internal/models"
    "gorm.io/gorm"
)

type Repository interface {
    CreateUser(ctx context.Context, user *models.User) error
    GetUserByEmail(ctx context.Context, email string) (*models.User, error)
    StoreRefreshToken(ctx context.Context, userID uint, tokenID string, expiresIn time.Duration) error
    GetRefreshTokenUserID(ctx context.Context, tokenID string) (uint, error)
    RevokeRefreshToken(ctx context.Context, tokenID string) error
}

type repository struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewRepository(db *gorm.DB, redis *redis.Client) Repository {
    return &repository{db: db, redis: redis}
}

func (r *repository) CreateUser(ctx context.Context, user *models.User) error {
    result := r.db.WithContext(ctx).Create(user)
    return result.Error
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
    if result.Error != nil {
        return nil, result.Error
    }
    return &user, nil
}

func (r *repository) StoreRefreshToken(ctx context.Context, userID uint, tokenID string, expiresIn time.Duration) error {
    return r.redis.Set(ctx, tokenID, fmt.Sprint(userID), expiresIn).Err()
}

func (r *repository) GetRefreshTokenUserID(ctx context.Context, tokenID string) (uint, error) {
    val, err := r.redis.Get(ctx, tokenID).Result()
    if err != nil {
        return 0, err
    }
    userID, err := strconv.ParseUint(val, 10, 64)
    if err != nil {
        return 0, err
    }
    return uint(userID), nil
}

func (r *repository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
    return r.redis.Del(ctx, tokenID).Err()
}
