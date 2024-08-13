package auth

import (
	"fmt"
	"net/http"
	"university-backend/foundation/web"
	"university-backend/internal/commands"
	"university-backend/internal/repository/postgres/user"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Controller struct {
	user User
}

func NewController(user User) *Controller {
	return &Controller{user: user}
}

func (uc Controller) SignIn(c *web.Context) error {
	var data user.SignInRequest

	err := c.BindFunc(&data, "EmployeeID", "Password")
	if err != nil {
		return c.RespondError(err)
	}

	detail, err := uc.user.GetByEmployeeID(c.Ctx, data.EmployeeID)
	if err != nil {
		return c.RespondError(err)
	}

	if detail.Password == nil {
		return c.RespondError(&web.Error{
			Err:    errors.New("area not found"),
			Status: http.StatusNotFound,
		})
	}

	if err = bcrypt.CompareHashAndPassword([]byte(*detail.Password), []byte(data.Password)); err != nil {
		return c.RespondError(web.NewRequestError(errors.New(fmt.Sprintf("incorrect password. error: %v", err)), http.StatusBadRequest))
	}

	accessToken, refreshToken, err := commands.GenToken(user.AuthClaims{
		ID:   detail.ID,
		Role: *detail.Role,
	}, "./private.pem")

	if err != nil {
		return c.RespondError(err)
	}

	return c.Respond(map[string]interface{}{
		"status": true,
		"data": map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
		"error": nil,
	}, http.StatusOK)
}

func (uc Controller) RefreshToken(c *web.Context) error {
	var data user.RefreshTokenRequest

	err := c.BindFunc(&data, "AccessToken", "RefreshToken")
	if err != nil {
		return c.RespondError(err)
	}

	_, refreshTokenClaims, err := commands.VerifyTokens(data.AccessToken, data.RefreshToken, "./private.pem")
	if err != nil {
		return c.RespondError(web.NewRequestError(err, http.StatusUnauthorized))
	}

	// Generate new tokens
	userClaims := user.AuthClaims{
		ID:   refreshTokenClaims.UserId,
		Role: refreshTokenClaims.Role,
	}

	accessToken, refreshToken, err := commands.GenToken(userClaims, "./private.pem")
	if err != nil {
		return c.RespondError(web.NewRequestError(errors.Wrap(err, "generating new tokens"), http.StatusInternalServerError))
	}

	return c.Respond(map[string]interface{}{
		"status": true,
		"data": map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
		"error": nil,
	}, http.StatusOK)
}
