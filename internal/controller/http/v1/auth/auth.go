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
	fmt.Println("Auth:", data)
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
			Err:    errors.New("user not found"),
			Status: http.StatusNotFound,
		})
	}

	//if *detail.Password != data.Password {
	//	return c.RespondError(&web.Error{
	//		Err:    errors.New("incorrect password"),
	//		Status: http.StatusBadRequest,
	//	})
	//}

	if err = bcrypt.CompareHashAndPassword([]byte(*detail.Password), []byte(data.Password)); err != nil {
		return c.RespondError(web.NewRequestError(errors.New(fmt.Sprintf("incorrect password!")), http.StatusBadRequest))
	}

	//var cfg struct {
	//	conf.Version
	//	Args conf.Args
	//	DB   struct {
	//		User       string `conf:"default:postgres"`
	//		Password   string `conf:"default:1"`
	//		Host       string `conf:"default:0.0.0.0"`
	//		Name       string `conf:"default:onda_b2b"`
	//		DisableTLS bool   `conf:"default:true"`
	//	}
	//}
	//cfg.Version.SVN = build

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
