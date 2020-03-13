package main

import (
	"fmt"
	"github.com/99designs/gqlgen/handler"
	"github.com/andrewmthomas87/litterbox/auth"
	"github.com/andrewmthomas87/litterbox/graphql"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/mediocregopher/radix/v3"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go"
	"log"
	"net/http"
	"time"
)

type claims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func handleAuth(store *auth.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(viper.GetString("auth.cookieName"))
		if err != nil || len(tokenString) == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(viper.GetString("auth.signingKey")), nil
		})
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*claims); ok && token.Valid {
			exists, err := store.UserExists(claims.UserID)
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
			} else if !exists {
				c.AbortWithStatus(http.StatusUnauthorized)
			} else {
				c.Set("user_id", claims.UserID)
			}
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func handleGraphQLQuery(db *sqlx.DB) gin.HandlerFunc {
	h := handler.GraphQL(graphql.NewExecutableSchema(graphql.Config{Resolvers: &graphql.Resolver{Db: db}}))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request.WithContext(c))
	}
}

func handlePlayground() gin.HandlerFunc {
	h := handler.Playground("GraphQL", "http://localhost:8080/graphql/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(fmt.Errorf("Fatal error config file: %s \n", err))
	}

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

	stripe.Key = viper.GetString("stripe.key")

	r := gin.Default()
	g := r.Group("/graphql")
	if viper.GetBool("dev") {
		g.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	authorized := g.Group("/")
	authorized.Use(handleAuth(store))

	authorized.POST("/query", handleGraphQLQuery(db))
	if viper.GetBool("dev") {
		authorized.GET("/playground", handlePlayground())
	}

	_ = r.Run(viper.GetString("api.serverAddress"))
}
