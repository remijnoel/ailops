package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kaptinlin/jsonschema"
	openai "github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

type JSONSchema interface {
	json.Marshaler
	json.Unmarshaler
}

type OpenAIProvider struct {
	ApiKey       string
	client       *openai.Client
	systemPrompt string
	model        Model
}

func NewOpenAIProvider(apiKey string, systemPrompt string, model Model) *OpenAIProvider {
	return &OpenAIProvider{
		ApiKey:       apiKey,
		client:       openai.NewClient(apiKey),
		systemPrompt: systemPrompt,
		model:        model,
	}
}

func (p *OpenAIProvider) RequestCompletion(prompt string) (string, error) {
	log.Debugf("Requesting completion from OpenAI with prompt: %s", prompt)
	resp, err := p.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: p.model.Name,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: p.systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) RequestCompletionWithJSONSchema(prompt string, jsonSchema *jsonschema.Schema) (string, error) {

	resp, err := p.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: p.model.Name,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: p.systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:   "something",
					Strict: true,
					Schema: jsonSchema,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	completion := resp.Choices[0].Message.Content
	log.Debugf("OpenAI response: %s", completion)

	// Check if the response is valid in respect to the provided JSON schema
	validationResult := jsonSchema.ValidateJSON([]byte(completion))
	if validationResult.IsValid() {
		log.Debugf("Response is valid according to the declared schema")
	} else {
		errMsg := "Response is not valid according to the schema"
		if sch, err := jsonSchema.MarshalJSON(); err != nil {
			log.Errorf("Failed to marshal schema: %v", err)
		} else {
			log.Errorf("Schema: %s", sch)
		}

		for field, err := range validationResult.Errors {
			log.Errorf("Validation error in field '%s': %s", field, err)
			errMsg += "\n" + field + ": " + err.Error()
		}
		return "", fmt.Errorf("%s: %s", errMsg, completion)
	}

	return resp.Choices[0].Message.Content, nil
}
