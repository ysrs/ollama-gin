package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms/ollama"
	"net/http"
)

func main() {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/generate", generateResponse)
	}
	r.Run(":8080")
}

func generateResponse(c *gin.Context) {
	var requestData struct {
		Prompt string `json:"prompt"`
	}
	if err := c.BindJSON(&requestData); err != nil {
		c.JSONP(http.StatusBadRequest, gin.H{"error": "Invalid Json"})
		return
	}

	llm, err := ollama.New(ollama.WithModel("qwen"))
	if err != nil {
		c.JSONP(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response, err := llm.Call(context.Background(), requestData.Prompt)
	if err != nil {
		c.JSONP(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSONP(http.StatusOK, gin.H{"Response": response})
}
