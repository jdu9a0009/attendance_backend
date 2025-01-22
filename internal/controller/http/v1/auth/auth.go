package auth

import (
	"attendance/backend/foundation/web"
	"attendance/backend/internal/commands"
	"attendance/backend/internal/entity"
	"attendance/backend/internal/repository/postgres/user"
	"fmt"
	"net/http"
	"regexp"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	errIncorrectPassword   = errors.New("社員番号またはメールアドレス が間違っています")
	errIncorrectEmployeeId = errors.New("パスワードが間違っています")
)

type Controller struct {
	user User
}

func NewController(user User) *Controller {
	return &Controller{user: user}
}

// Helper function to check if a string is a valid email
func isValidEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

// @Description SignIn User
// @Summary SignIn User
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body user.SignInRequest true "Sign In"
// @Success 200 {object} web.ErrorResponse
// @Failure 400,404,500,401 {object} web.ErrorResponse
// @Router /api/v1/sign-in [post]
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

	var detail *entity.User
	if isValidEmail(data.EmployeeID) {
		// Fetch user by Email
		detail, err = uc.user.GetByEmployeeEmail(c.Ctx, data.EmployeeID)
	} else {
		// Fetch user by EmployeeID
		detail, err = uc.user.GetByEmployeeID(c.Ctx, data.EmployeeID)
	}

	if err != nil || detail == nil {
		fmt.Println("User not found or invalid credentials for Identifier:", data.EmployeeID)
		return c.RespondError(&web.Error{
			Err:    errIncorrectPassword,
			Status: http.StatusUnauthorized,
		})
	}

	if detail.Password == nil {
		fmt.Println("Password not found for Identifier:", data.EmployeeID)
		return c.RespondError(&web.Error{
			Err:    errIncorrectEmployeeId,
			Status: http.StatusNotFound,
		})
	}

	// Verify password
	if err = bcrypt.CompareHashAndPassword([]byte(*detail.Password), []byte(data.Password)); err != nil {
		fmt.Println("Incorrect password for Identifier:", data.EmployeeID)
		return c.RespondError(&web.Error{
			Err:    errIncorrectEmployeeId,
			Status: http.StatusUnauthorized,
		})
	}

	// Generate tokens
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

// @Description Refresh Token
// @Summary Refresh Token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh body user.RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} web.ErrorResponse
// @Failure 400,404,500,401 {object} web.ErrorResponse
// @Router /api/v1/refresh-token [post]
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
