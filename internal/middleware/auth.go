package middleware

import (
	"attendance/backend/foundation/web"
	"attendance/backend/internal/auth"
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

func Authenticate(a *auth.Auth, role ...string) web.Middleware {
	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(c *web.Context) error {

			// Expecting: Bearer <token>
			authStr := c.Request.Header.Get("authorization")

			// Parse the authorization header.
			parts := strings.Split(authStr, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: Bearer <token>")
				return c.RespondError(web.NewRequestError(err, http.StatusUnauthorized))
			}

			// Validate the token is signed by us.
			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return c.RespondError(web.NewRequestError(err, http.StatusUnauthorized))
			}

			//check role inside token data
			if ok := claims.Authorized(role...); !ok && (len(role) > 0) {
				return c.RespondError(web.NewRequestError(errors.New("attempted action is not allowed"), http.StatusUnauthorized))
			}

			// check if claims from database
			//if err = a.CheckClaimsDataFromDatabase(c.Ctx, claims); err != nil {
			//	return c.RespondError(err)
			//}

			// Add claims to the context so that they can be retrieved later.
			c.Ctx = context.WithValue(c.Ctx, auth.Key, claims)

			// Call the next handler.
			return handler(c)
		}

		return h
	}

	return m
}
func ValidateEmailAndPhoneInput() web.Middleware {
	// Regex definitions for email and phone number validation.
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex := regexp.MustCompile(`^\+?\d*$`)

	return func(handler web.Handler) web.Handler {
		return func(c *web.Context) error {
			// Get email and phone from request form.
			email := c.Request.FormValue("email")
			phone := c.Request.FormValue("phone")

			// Validate email if provided.
			if email == "" {
				return c.RespondError(web.NewRequestError(errors.New("メールアドレスは必須です"), http.StatusBadRequest))
			}
			if !emailRegex.MatchString(email) {
				return c.RespondError(web.NewRequestError(errors.New("無効なメールアドレス形式"), http.StatusBadRequest))
			}

			if !phoneRegex.MatchString(phone) {
				return c.RespondError(web.NewRequestError(errors.New("無効な電話番号形式"), http.StatusBadRequest))
			}

			// Proceed to the next handler if validation passes.
			return handler(c)
		}
	}
}

// isHalfWidth checks if a string contains only half-width characters.
func isHalfWidth(s string) bool {
	// Normalize the string to NFC form.
	normalized := norm.NFC.String(s)
	for _, r := range normalized {
		// Full-width character detection
		if r >= '\uFF01' && r <= '\uFF60' || r >= '\uFFE0' && r <= '\uFFEF' {
			return false
		}
	}
	return true
}

func ValidateHalfWidthInput() web.Middleware {
	return func(handler web.Handler) web.Handler {
		return func(c *web.Context) error {
			// Iterate over form values and validate each one.
			for _, values := range c.Request.Form {
				for _, value := range values {
					if !isHalfWidth(value) {
						return c.RespondError(web.NewRequestError(
							errors.New("入力は半角文字のみ使用可能"), http.StatusBadRequest))
					}
				}
			}

			// Proceed to the next handler if validation passes.
			return handler(c)
		}
	}
}
