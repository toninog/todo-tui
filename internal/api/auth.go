package api

// AuthRequest represents login/register request
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents login/register response
type AuthResponse struct {
	OK    bool   `json:"ok"`
	Token string `json:"token"`
}

// UserInfo represents the current user info
type UserInfo struct {
	ID      int    `json:"id"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

// Login authenticates with email/password and stores the token
func (c *Client) Login(email, password string) error {
	req := AuthRequest{Email: email, Password: password}
	var resp AuthResponse

	if err := c.Post("/auth/login", req, &resp); err != nil {
		return err
	}

	// Save credentials for auto-login
	c.SaveCredentials(email, password)

	if resp.Token != "" {
		return c.saveToken(resp.Token)
	}

	return nil
}

// Register creates a new account and stores the token
func (c *Client) Register(email, password string) error {
	req := AuthRequest{Email: email, Password: password}
	var resp AuthResponse

	if err := c.Post("/auth/register", req, &resp); err != nil {
		return err
	}

	// Save credentials for auto-login
	c.SaveCredentials(email, password)

	if resp.Token != "" {
		return c.saveToken(resp.Token)
	}

	return nil
}

// Logout clears the local token and credentials
func (c *Client) Logout() error {
	// Try to call logout endpoint (optional, as token is stateless)
	c.Post("/auth/logout", nil, nil)
	c.ClearCredentials()
	return c.clearToken()
}

// GetCurrentUser fetches the current user's info
func (c *Client) GetCurrentUser() (*UserInfo, error) {
	var user UserInfo
	if err := c.Get("/auth/me", &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidateToken checks if the current token is valid
func (c *Client) ValidateToken() bool {
	if !c.HasToken() {
		return false
	}
	_, err := c.GetCurrentUser()
	return err == nil
}
