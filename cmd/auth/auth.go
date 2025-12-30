// Package auth provides authentication functionality for YouTube API access.
// It handles OAuth2 authentication flow, token management, and client setup.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

// ClientConfig represents OAuth2 client configuration structure as provided by Google.
// It contains credentials and endpoints for OAuth2 authentication.
type ClientConfig struct {
	Installed struct {
		ClientID                string `json:"client_id"`                   // OAuth client ID
		ProjectID               string `json:"project_id"`                  // Google Cloud project ID
		AuthURI                 string `json:"auth_uri"`                    // Authorization endpoint
		TokenURI                string `json:"token_uri"`                   // Token endpoint
		AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"` // Certificate URL for the auth provider
		ClientSecret            string `json:"client_secret"`               // OAuth client secret
	} `json:"installed"`
}

// getTokenPath returns the file path where OAuth tokens are stored.
func getTokenPath() string {
	return "youtube-token.json"
}

// Login initiates the OAuth2 authentication flow for YouTube API access.
// It prompts the user to authorize access in a browser and captures the authorization code.
func Login() error {
	config, err := loadClientConfig()
	if err != nil {
		return err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.Installed.ClientID,
		ClientSecret: config.Installed.ClientSecret,
		RedirectURL:  "http://localhost",
		Scopes: []string{
			youtube.YoutubeUploadScope,
			youtube.YoutubeReadonlyScope,
		},
		Endpoint: google.Endpoint,
	}

	// Start a lightweight local server to capture the ?code= from http://localhost redirect
	codeCh := make(chan string, 1)
	srv := &http.Server{Addr: ":80"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			// Show a helpful message if no code present
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintln(w, "Missing 'code' in query. Please copy the full URL and paste the code in the terminal.")
			return
		}
		_, _ = fmt.Fprintln(w, "Auth complete. You can close this tab.")
		// Send code to channel for exchange
		select {
		case codeCh <- code:
		default:
		}
		go func() {
			// Shutdown server shortly after capturing code
			time.Sleep(500 * time.Millisecond)
			_ = srv.Shutdown(context.Background())
		}()
	})

	go func() {
		_ = srv.ListenAndServe()
	}()

	// Open the consent URL in the default browser (Windows-friendly)
	authURL := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", authURL).Start()
	fmt.Printf("\nA browser window opened to complete Google consent.\nIf it doesn't open, visit:\n\n%v\n\n", authURL)
	fmt.Println("Waiting for redirect to http://localhost with ?code=... (or you can paste manually)")

	var code string
	select {
	case code = <-codeCh:
		// captured automatically
		fmt.Println("Received authorization code via redirect.")
	case <-time.After(60 * time.Second):
		// Fallback to manual paste
		fmt.Print("Paste the authorization code from the browser URL: ")
		if _, err := fmt.Scan(&code); err != nil {
			return fmt.Errorf("unable to read authorization code: %v", err)
		}
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("unable to exchange code for token: %v", err)
	}

	return saveToken(token)
}

// loadClientConfig reads and parses the OAuth client configuration file.
// Returns the parsed client configuration or an error if the file cannot be read or parsed.
func loadClientConfig() (*ClientConfig, error) {
	config := &ClientConfig{}
	configFile := "client_secrets.json"

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %v", err)
	}

	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("error parsing configuration: %v", err)
	}

	return config, nil
}

// saveToken persists an OAuth token to the filesystem for future use.
// The token is stored in the file specified by getTokenPath().
func saveToken(token *oauth2.Token) error {
	tokenPath := getTokenPath()
	f, err := os.OpenFile(tokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to create token file: %v", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

// GetClient retrieves the stored OAuth token.
// Returns an error if the token doesn't exist or can't be parsed.
func GetClient() (*oauth2.Token, error) {
	tokenPath := getTokenPath()
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("token not found. Run 'gopherupload login' first: %v", err)
	}

	token := &oauth2.Token{}
	err = json.Unmarshal(data, token)
	if err != nil {
		return nil, fmt.Errorf("error reading token: %v", err)
	}

	return token, nil
}

// GetAuthenticatedClient returns an HTTP client with automatic token refresh capability.
// The token will be automatically refreshed when it expires and saved for future use.
func GetAuthenticatedClient() (*http.Client, error) {
	config, err := loadClientConfig()
	if err != nil {
		return nil, err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.Installed.ClientID,
		ClientSecret: config.Installed.ClientSecret,
		RedirectURL:  "http://localhost",
		Scopes: []string{
			youtube.YoutubeUploadScope,
			youtube.YoutubeReadonlyScope,
		},
		Endpoint: google.Endpoint,
	}

	token, err := GetClient()
	if err != nil {
		return nil, err
	}

	// Create a token source that automatically refreshes the token
	ctx := context.Background()
	tokenSource := oauthConfig.TokenSource(ctx, token)

	// Wrap with ReuseTokenSource to cache the token and only refresh when needed
	// This also allows us to save the new token when it's refreshed
	reusableTokenSource := oauth2.ReuseTokenSource(token, &savingTokenSource{
		tokenSource: tokenSource,
	})

	return oauth2.NewClient(ctx, reusableTokenSource), nil
}

// savingTokenSource wraps a token source and saves new tokens to disk
type savingTokenSource struct {
	tokenSource oauth2.TokenSource
}

// Token returns a token, saving it to disk if it's a new one
func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.tokenSource.Token()
	if err != nil {
		return nil, err
	}

	// Save the potentially refreshed token
	if err := saveToken(token); err != nil {
		// Log the error but don't fail - we still have a valid token
		fmt.Printf("Warning: could not save refreshed token: %v\n", err)
	}

	return token, nil
}
