package auth

import (
	"fmt"
	"net/http"
	"attendance/backend/foundation/web"
	"attendance/backend/internal/commands"
	"attendance/backend/internal/repository/postgres/user"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// build is the git version of this hard_skill. It is set using build flags in the makefile.
var build = "develop"

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
		fmt.Println("Error binding request data:", err)
		return c.RespondError(&web.Error{
			Err:    errors.New("invalid request data"),
			Status: http.StatusBadRequest,
		})
	}

	detail, err := uc.user.GetByEmployeeID(c.Ctx, data.EmployeeID)
	if err != nil {
		fmt.Println("Error retrieving user by EmployeeID:", err)
		return c.RespondError(&web.Error{
			Err:    errors.New("user retrieval failed"),
			Status: http.StatusInternalServerError,
		})
	}

	if detail.Password == nil {
		fmt.Println("User not found for EmployeeID:", data.EmployeeID)
		return c.RespondError(&web.Error{
			Err:    errors.New("user not found"),
			Status: http.StatusNotFound,
		})
	}

	if err = bcrypt.CompareHashAndPassword([]byte(*detail.Password), []byte(data.Password)); err != nil {
		fmt.Println("Incorrect password for EmployeeID:", data.EmployeeID)
		return c.RespondError(&web.Error{
			Err:    errors.New("incorrect password"),
			Status: http.StatusForbidden, // Changed to 403 to reflect forbidden access
		})
	}

	accessToken, refreshToken, err := commands.GenToken(user.AuthClaims{
		ID:   detail.ID,
		Role: *detail.Role,
	}, "./private.pem")

	if err != nil {
		fmt.Println("Error generating tokens:", err)
		return c.RespondError(&web.Error{
			Err:    errors.New("token generation failed"),
			Status: http.StatusInternalServerError,
		})
	}

	return c.Respond(map[string]interface{}{
		"status": true,
		"data": map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"role":          *detail.Role,
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
