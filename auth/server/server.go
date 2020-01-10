package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/andrewmthomas87/litterbox/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/mediocregopher/radix/v3"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const googleUserInfoEndpointURL = "https://www.googleapis.com/userinfo/v2/me"

type claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type googleUserInfoResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func handleSignIn(state string, config *oauth2.Config) gin.HandlerFunc {
	redirectUrl := config.AuthCodeURL(state)

	return func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
	}
}

func handleSignOut() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetCookie(viper.GetString("auth.cookieName"), "", 0, "/", "", false, true)
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}

func handleAuth(state string, store *auth.Store, config *oauth2.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Redirect(http.StatusTemporaryRedirect, "/")

		errString := c.Query("error")
		if len(errString) > 0 {
			return
		}

		if state != c.Query("state") {
			return
		}

		code := c.Query("code")
		if len(code) == 0 {
			return
		}

		token, err := config.Exchange(c, code)
		if err != nil {
			return
		}

		client := config.Client(c, token)

		resp, err := client.Get(googleUserInfoEndpointURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		var user googleUserInfoResponse
		if err := json.Unmarshal(body, &user); err != nil {
			return
		}

		if err := store.SetUser(c, user.ID, user.Email, user.Name, token.AccessToken, token.RefreshToken); err != nil {
			return
		}

		claims := claims{
			UserID:         user.ID,
			StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Unix()},
		}
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := jwtToken.SignedString([]byte(viper.GetString("auth.signingKey")))
		if err != nil {
			return
		}

		c.SetCookie(viper.GetString("auth.cookieName"), tokenString, int(30*24*time.Hour.Seconds()), "/", "", false, true)
	}
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	stateBytes := make([]byte, 64)
	if _, err := rand.Read(stateBytes); err != nil {
		log.Fatal(err)
	}
	state := url.QueryEscape(string(stateBytes))

	db, err := sqlx.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			viper.GetString("database.user"),
			viper.GetString("database.password"),
			viper.GetString("database.host"),
			viper.GetInt("database.port"),
			viper.GetString("database.database")))
	if err != nil {
		log.Fatal(err)
	}

	pool, err := radix.NewPool(viper.GetString("redis.network"),
		viper.GetString("redis.address"),
		viper.GetInt("redis.poolSize"))

	store := auth.NewStore(db, pool)

	config := &oauth2.Config{
		ClientID:     viper.GetString("google.clientID"),
		ClientSecret: viper.GetString("google.clientSecret"),
		RedirectURL:  viper.GetString("google.redirectURL"),
		Scopes: []string{
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}

	r := gin.Default()
	g := r.Group("/auth")

	g.GET("/sign-in", handleSignIn(state, config))
	g.GET("/sign-out", handleSignOut())
	g.GET("/", handleAuth(state, store, config))

	_ = r.Run(viper.GetString("auth.serverAddress"))
}
