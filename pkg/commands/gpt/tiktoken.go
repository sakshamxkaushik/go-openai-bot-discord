package gpt

import (
	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)


// The countMessageTokens function is used to count the number of tokens used by the OpenAI API to generate a single response message.

// Function takes a model string as input and returns a boolean indicating whether the model is implemented, as well as the number of tokens per message and per name.
func countMessageTokens(message openai.ChatCompletionMessage, model string) *int {
	ok, tokensPerMessage, tokensPerName := _tokensConfiguration(model)
	if !ok {
		return nil
	}

	enc, err := tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		enc, _ = tokenizer.Get(tokenizer.Cl100kBase)
	}

	tokens := _countMessageTokens(enc, tokensPerMessage, tokensPerName, message)
	return &tokens
}

// The countMessagesTokens function is used to count the number of tokens used by the OpenAI API to generate multiple response messages. 
func countMessagesTokens(messages []openai.ChatCompletionMessage, model string) *int {
	ok, tokensPerMessage, tokensPerName := _tokensConfiguration(model)
	if !ok {
		return nil
	}

	enc, err := tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		enc, _ = tokenizer.Get(tokenizer.Cl100kBase)
	}

	tokens := 0
	for _, message := range messages {
		tokens += _countMessageTokens(enc, tokensPerMessage, tokensPerName, message)
	}
	tokens += 3 // every reply is primed with <im_start>assistant

	return &tokens
}

// The countAllMessagesTokens function is used to count the number of tokens used by the OpenAI API to generate all response messages, including system messages.
func countAllMessagesTokens(systemMessage *openai.ChatCompletionMessage, messages []openai.ChatCompletionMessage, model string) *int {
	if systemMessage != nil {
		messages = append(messages, *systemMessage)
	}
	return countMessagesTokens(messages, model)
}

// The _tokensConfiguration function is used to return the number of tokens per message and per name for a given GPT model. 
// _tokensConfiguration returns the number of tokens per message and per name for a given GPT model.
// If the model is not implemented, ok is set to false.
func _tokensConfiguration(model string) (ok bool, tokensPerMessage int, tokensPerName int) {
	ok = true

	switch model {
	case openai.GPT3Dot5Turbo0301:
		tokensPerMessage = 4 // every message follows <im_start>{role/name}\n{content}<im_end>\n
		tokensPerName = -1   // if there's a name, the role is omitted
	case openai.GPT3Dot5Turbo,
		openai.GPT3Dot5Turbo0613,
		openai.GPT3Dot5Turbo16K,
		openai.GPT3Dot5Turbo16K0613,
		openai.GPT4,
		openai.GPT40314,
		openai.GPT40613,
		openai.GPT432K0314,
		openai.GPT432K0613:
		tokensPerMessage = 3
		tokensPerName = 1
	default:
		// Not implemented
		ok = false
		return
	}

	return
}

// The _countMessageTokens function is used to count the number of tokens used by the OpenAI API to generate a single response message. 
// The function takes a message of type openai.ChatCompletionMessage as input and returns an integer representing the number of tokens used.
func _countMessageTokens(enc tokenizer.Codec, tokensPerMessage int, tokensPerName int, message openai.ChatCompletionMessage) int {
	tokens := tokensPerMessage
	contentIds, _, _ := enc.Encode(message.Content)
	roleIds, _, _ := enc.Encode(message.Role)
	tokens += len(contentIds)
	tokens += len(roleIds)
	if message.Name != "" {
		tokens += tokensPerName
		nameIds, _, _ := enc.Encode(message.Name)
		tokens += len(nameIds)
	}
	return tokens
}
