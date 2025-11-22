package dto

import (
	"github.com/google/uuid"
)

type UserResponse struct {
	UUID     uuid.UUID `json:"uuid"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName *string `json:"full_name"`
}

type UserPatch struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	FullName *ErasableString `json:"full_name"`
}
