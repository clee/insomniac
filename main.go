package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/google/go-github/github"

	gin "gopkg.in/gin-gonic/gin.v1"
	githubhook "gopkg.in/rjz/githubhook.v0"
)

func main() {
	r := gin.Default()
	gin.SetMode(gin.DebugMode)

	secretenv := os.Getenv("GITHUB_SECRET")
	if secretenv == "" {
		log.Fatal("$GITHUB_SECRET must be set!")
	}
	secret := []byte(secretenv)

	r.POST("/event_handler", func(c *gin.Context) {
		hook, err := githubhook.Parse(secret, c.Request)
		if err != nil {
			log.Printf("error parsing hook: %+v", err.Error())
			return
		}

		event := github.PullRequestEvent{}
		err = json.Unmarshal(hook.Payload, &event)
		if err != nil {
			log.Printf("error parsing event: %+v", err.Error())
			return
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set!")
	}
	r.Run(":" + port)
}
