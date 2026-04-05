package models

import "time"

type AdminUser struct {
	ID               string     `json:"id"`
	Provider         string     `json:"provider"`
	Role             UserRole   `json:"role"`
	Status           UserStatus `json:"status"`
	Email            string     `json:"email"`
	Name             string     `json:"name"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
	LastLoginAt      time.Time  `json:"lastLoginAt"`
	TransactionCount int        `json:"transactionCount"`
}

type AdminUserOption struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Email  string     `json:"email"`
	Role   UserRole   `json:"role"`
	Status UserStatus `json:"status"`
}

type AdminUserCreateRequest struct {
	Name     string     `json:"name"`
	Email    string     `json:"email"`
	Password string     `json:"password"`
	Role     UserRole   `json:"role"`
	Status   UserStatus `json:"status"`
}

type AdminUserCreateInput struct {
	Name     string
	Email    string
	Password string
	Role     UserRole
	Status   UserStatus
}

type AdminUserUpdateRequest struct {
	Name     string     `json:"name"`
	Email    string     `json:"email"`
	Password string     `json:"password"`
	Role     UserRole   `json:"role"`
	Status   UserStatus `json:"status"`
}

type AdminUserUpdateInput struct {
	Name     string
	Email    string
	Password string
	Role     UserRole
	Status   UserStatus
}
