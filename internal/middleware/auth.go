package middleware

import (
	"attendance/backend/foundation/web"
	"attendance/backend/internal/auth"
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
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
func ValidateUserInput() web.Middleware {
	phoneRegex := regexp.MustCompile(`^0\d{1,4}-?\d{1,4}-?\d{3,4}$`)

	return func(handler web.Handler) web.Handler {
		return func(c *web.Context) error {
			phone := c.Request.FormValue("phone")
			email := c.Request.FormValue("email")

			// Normalize the phone number (add dashes if needed).
			normalizedPhone := addDashesToPhone(phone)

			// Check if input is a valid phone number.
			if !isValidPhoneNumber(normalizedPhone, phoneRegex) {
				return c.RespondError(web.NewRequestError(errors.New("無効な電話番号形式"), http.StatusBadRequest))
			}

			// Check if input is a valid email.
			if !isValidEmail(email) {
				return c.RespondError(web.NewRequestError(errors.New("無効なメールアドレス形式"), http.StatusBadRequest))
			}

			// Call the next handler if validation passes.
			return handler(c)
		}
	}
}

// addDashesToPhone formats a Japanese phone number by adding dashes.
// It converts numbers like "0358462131" to "03-5846-2131" or "0451234567" to "045-123-4567".
func addDashesToPhone(phone string) string {
	// Remove any existing dashes just in case.
	phone = strings.ReplaceAll(phone, "-", "")

	// Ensure the phone number starts with a 0 and has 10 or 11 digits.
	if len(phone) < 10 || len(phone) > 11 || phone[0] != '0' {
		return phone // Return as-is if it's not of expected length.
	}

	// For numbers starting with "03" (e.g., Tokyo area codes), format as "03-xxxx-xxxx".
	if strings.HasPrefix(phone, "03") {
		return phone[:2] + "-" + phone[2:6] + "-" + phone[6:]
	}

	// For numbers starting with a "0X" (e.g., 0xx-yyyy-zzzz like 045 for Yokohama), format as "0xx-xxx-xxxx".
	if len(phone) == 10 {
		return phone[:3] + "-" + phone[3:6] + "-" + phone[6:]
	}

	// For numbers with 11 digits (e.g., mobile numbers like 09012345678), format as "0xx-xxxx-xxxx".
	return phone[:3] + "-" + phone[3:7] + "-" + phone[7:]
}

// isValidPhoneNumber validates the phone number against the regex.
func isValidPhoneNumber(input string, regex *regexp.Regexp) bool {
	return regex.MatchString(input)
}

// isValidEmail performs a simple email format validation.
func isValidEmail(input string) bool {
	return strings.Contains(input, "@") && strings.Contains(input, ".")
}
