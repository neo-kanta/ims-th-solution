package request

type CreateUserRequest struct {
	Username    string `json:"username" validate:"required,min=3"`
	DisplayName string `json:"display_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
}

type AdminResetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type AssignGroupsRequest struct {
	GroupIDs []string `json:"group_ids" validate:"required"`
}
