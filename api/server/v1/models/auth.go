package models

type AuthSignInResponse struct {
	CurrentUser *CurrentUserModel `json:"current_user"`
	AccessToken string            `json:"access_token"`
}