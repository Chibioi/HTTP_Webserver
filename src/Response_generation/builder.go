package response

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"strings"
)

type Request struct {
	Path   string
	Header map[string]string
}

// Success responses
func OK(body []byte) *Response {
	return BaseResponse(http.StatusOK).WithBody("text/plain", body)
}

func JSON(data interface{}) *Response {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return InternalServerError(err) // Handle marshal error.
	}
	return BaseResponse(http.StatusOK).
		WithBody("application/json", jsonData)
}

func HTML(content string) *Response {
	return BaseResponse(http.StatusOK).
		WithBody("text/html", []byte(content))
}

// Error responses
func BadRequest(message string) *Response {
	return BaseResponse(http.StatusBadRequest).
		WithBody("text/plain", []byte(message))
}

func NotFound() *Response {
	return BaseResponse(http.StatusNotFound).
		WithBody("text/plain", []byte("404 Not Found"))
}

func InternalServerError(err error) *Response {
	return BaseResponse(http.StatusInternalServerError).
		WithBody("text/plain", []byte("500 Internal Server Error\n"+err.Error()))
}

// Redirect
func Redirect(url string, permanent bool) *Response {
	status := http.StatusFound // 302
	if permanent {
		status = http.StatusMovedPermanently // 301
	}
	return BaseResponse(status).
		WithHeader("Location", url)
}

func HandleRequest(conn net.Conn, req *Request) {
	// Example 1: JSON response
	if req.Path == "/api" {
		data := map[string]interface{}{"status": "ok"}
		JSON(data).Write(conn)
		return
	}

	// Example 2: Error handling
	if !isAuthorized(req) {
		BaseResponse(http.StatusUnauthorized).
			WithBody("text/plain", []byte("Unauthorized")).
			WithHeader("WWW-Authenticate", `Basic realm="Restricted"`).
			Write(conn)
		return
	}

	// Example 3: Set cookie
	cookie := &http.Cookie{
		Name:  "session",
		Value: "abc123",
		Path:  "/",
	}

	OK([]byte("Welcome!")).WithCookie(cookie).Write(conn)
}

func isAuthorized(req *Request) bool {
	authHeader := req.Header["Authorization"]
	if authHeader == "" {
		return false // No authorization header provided
	}

	authParts := strings.Split(authHeader, " ")
	if len(authParts) != 2 || authParts[0] != "Basic" {
		return false // Invalid authorization header format
	}

	decoded, err := base64.StdEncoding.DecodeString(authParts[1])
	if err != nil {
		return false // Invalid base64 encoding
	}

	credentials := strings.Split(string(decoded), ":")
	if len(credentials) != 2 {
		return false // Invalid credentials format
	}

	username := credentials[0]
	password := credentials[1]

	// Replace with your actual authentication logic (e.g., database lookup)
	return authenticateUser(username, password)
}

func authenticateUser(username, password string) bool {
	// Example: Hardcoded credentials (replace with your actual logic)
	if username == "myuser" && password == "mypassword" {
		return true
	}
	//Example: check against a database of users.
	// if checkDatabase(username, password) {
	//      return true
	// }

	return false
}
