package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
	"net/http"
)

func main() {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/generate", generateResponse)
		v1.POST("/translate", translate)
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

func translate(c *gin.Context) {
	var requestData struct {
		OutputLang string `json:"outputLang"`
		Text       string `json:"text"`
	}
	if err := c.BindJSON(&requestData); err != nil {
		c.JSONP(http.StatusBadRequest, gin.H{"error": "Invalid Json"})
		return
	}

	prompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		prompts.NewSystemMessagePromptTemplate("你是一个只能翻译文本的翻译引擎，不需要进行解释。", nil),
		prompts.NewHumanMessagePromptTemplate("翻译这段文字到 {{.outputLang}}: {{.text}}", []string{"outputLang", "text"}),
	})

	vals := map[string]any{
		"outputLang": requestData.OutputLang,
		"text":       requestData.Text,
	}

	messages, err := prompt.FormatMessages(vals)
	if err != nil {
		c.JSONP(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	llm, err := ollama.New(ollama.WithModel("qwen"))
	if err != nil {
		c.JSONP(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	content := []llms.MessageContent{
		llms.TextParts(messages[0].GetType(), messages[0].GetContent()),
		llms.TextParts(messages[1].GetType(), messages[1].GetContent()),
	}
	response, err := llm.GenerateContent(context.Background(), content)

	if err != nil {
		c.JSONP(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSONP(http.StatusOK, gin.H{"Response": response.Choices[0].Content})
}
