package jwt

import "github.com/golang-jwt/jwt/v5"

// CustomClaims 自定义 JWT Claims
type CustomClaims struct {
	UserID  int   `json:"user_id"`
	RoleIds []int `json:"role_ids"`
	jwt.RegisteredClaims
}

// GetUserID 获取用户 ID
func (c *CustomClaims) GetUserID() int {
	return c.UserID
}

// GetRoleIds 获取角色 ID 列表
func (c *CustomClaims) GetRoleIds() []int {
	return c.RoleIds
}
