package gpt

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/constants"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/utils"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)
// The file imports several packages, including http, io, log, and discordgo.

// See https://openai.com/pricing
const (
	gptPricePerPromptTokenGPT3Dot5Turbo0613     = 0.0000015
	gptPricePerCompletionTokenGPT3Dot5Turbo0613 = 0.000002

	gptPricePerPromptTokenGPT3Dot5Turbo16K0613     = 0.000003
	gptPricePerCompletionTokenGPT3Dot5Turbo16K0613 = 0.000004

	gptPricePerPromptTokenGPT40613     = 0.00003
	gptPricePerCompletionTokenGPT40613 = 0.00006

	gptPricePerPromptTokenGPT432K0613     = 0.00006
	gptPricePerCompletionTokenGPT432K0613 = 0.00012
)

const (
	gptTruncateLimitGPT3Dot5Turbo0301 = 3500
	gptTruncateLimitGPT40314          = 6500
	gptTruncateLimitGPT432K0314       = 30500
)


// The shouldHandleMessageType function is used to determine whether a given Discord message should be processed by the bot. 
// It returns true if the message type is discord.MessageTypeDefault or discord.MessageTypeReply.
func shouldHandleMessageType(t discord.MessageType) bool {
	return t == discord.MessageTypeDefault || t == discord.MessageTypeReply
}

type chatGPTResponse struct {
	content string
	usage   openai.Usage
}

// The sendChatGPTRequest function sends a request to the OpenAI API to generate a response to a given prompt using the GPT model. 
// The function takes a client object, which is used to make the API request, and a cacheItem object, which contains the messages that make up the conversation. 
// The function returns a chatGPTResponse object, which contains the generated response and usage information.
func sendChatGPTRequest(client *openai.Client, cacheItem *MessagesCacheData) (*chatGPTResponse, error) {
	// Create message with ChatGPT
	messages := cacheItem.Messages
	if cacheItem.SystemMessage != nil {
		messages = append([]openai.ChatCompletionMessage{*cacheItem.SystemMessage}, messages...)
	}

	req := openai.ChatCompletionRequest{
		Model:    cacheItem.Model,
		Messages: messages,
	}

	if cacheItem.Temperature != nil {
		req.Temperature = *cacheItem.Temperature
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		req,
	)
	if err != nil {
		return nil, err
	}

	// Save response to context cache
	responseContent := resp.Choices[0].Message.Content
	cacheItem.Messages = append(cacheItem.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: responseContent,
	})
	cacheItem.TokenCount = resp.Usage.TotalTokens
	return &chatGPTResponse{
		content: responseContent,
		usage:   resp.Usage,
	}, nil
}

// The getUrlData function sends an HTTP GET request to a given URL and returns the response body as a string.
func getUrlData(client *http.Client, url string) (string, error) {
	res, err := client.Get(url)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// The getContentOrURLData function takes a string and returns either the string itself or the response body of an HTTP GET request to the URL specified by the string.
func getContentOrURLData(client *http.Client, s string) (content string, err error) {
	if utils.IsURL(s) {
		content, err = getUrlData(client, s)
	}
	return content, err
}

// The parseInteractionReply function takes a Discord message and extracts the prompt, context, model, and temperature from the message's embeds.
func parseInteractionReply(discordMessage *discord.Message) (prompt string, context string, model string, temperature *float32) {
	if discordMessage.Embeds == nil || len(discordMessage.Embeds) == 0 {
		return
	}

	for _, embed := range discordMessage.Embeds {
		if embed.Description != "" {
			prompt = embed.Description
		}
		for _, field := range embed.Fields {
			switch field.Name {
			case gptCommandOptionPrompt.humanReadableString():
				prompt = field.Value
			case gptCommandOptionContext.humanReadableString():
				if context == "" {
					// file context always gets precedence
					context = field.Value
				}
			case gptCommandOptionContextFile.humanReadableString():
				context = field.Value
			case gptCommandOptionModel.humanReadableString():
				model = field.Value
			case gptCommandOptionTemperature.humanReadableString():
				parsedValue, err := strconv.ParseFloat(field.Value, 32)
				if err != nil {
					log.Printf("[GID: %s, CHID: %s, MID: %s] Failed to parse temperature value from the message with the error: %v\n", discordMessage.GuildID, discordMessage.ChannelID, discordMessage.ID, err)
					continue
				}
				temp := float32(parsedValue)
				temperature = &temp
			}
		}
	}

	return
}

// The modelTruncateLimit function takes a model name and returns the maximum number of tokens that can be used in a message for that model.
func modelTruncateLimit(model string) *int {
	var truncateLimit int
	switch model {
	case openai.GPT3Dot5Turbo, openai.GPT3Dot5Turbo0301:
		// gpt-3.5-turbo may change over time. Assigning truncate limit assuming gpt-3.5-turbo-0301
		truncateLimit = gptTruncateLimitGPT3Dot5Turbo0301
	case openai.GPT4, openai.GPT40314:
		// gpt-4 may change over time. Assigning truncate limit assuming gpt-4-0314
		truncateLimit = gptTruncateLimitGPT40314
	case openai.GPT432K, openai.GPT432K0314:
		// gpt-4-32k may change over time. Assigning truncate limit assuming gpt-4-32k-0314
		truncateLimit = gptTruncateLimitGPT432K0314
	default:
		// Not implemented
		return nil
	}
	return &truncateLimit
}


// The adjustMessageTokens function removes messages from a conversation until the total number of tokens in the conversation is below the maximum allowed for the model.
func adjustMessageTokens(cacheItem *MessagesCacheData) {
	truncateLimit := modelTruncateLimit(cacheItem.Model)
	if truncateLimit == nil {
		return
	}

	for cacheItem.TokenCount > *truncateLimit {
		message := cacheItem.Messages[0]
		cacheItem.Messages = cacheItem.Messages[1:]
		removedTokens := countMessageTokens(message, cacheItem.Model)
		if removedTokens == nil {
			return
		}
		cacheItem.TokenCount -= *removedTokens
	}
}


// The isCacheItemWithinTruncateLimit function checks whether a given conversation is within the maximum token limit for the model.
func isCacheItemWithinTruncateLimit(cacheItem *MessagesCacheData) (ok bool, count int) {
	truncateLimit := modelTruncateLimit(cacheItem.Model)
	if truncateLimit == nil {
		return true, 0
	}

	tokens := countAllMessagesTokens(cacheItem.SystemMessage, cacheItem.Messages, cacheItem.Model)
	if tokens == nil {
		return true, 0
	}
	cacheItem.TokenCount = *tokens

	return *tokens <= *truncateLimit, *tokens
}


// The generateThreadTitleBasedOnInitialPrompt function generates a thread title based on the initial prompt of a conversation.
func generateThreadTitleBasedOnInitialPrompt(ctx *bot.Context, client *openai.Client, threadID string, messages []openai.ChatCompletionMessage) {
	conversation := make([]map[string]string, len(messages))
	for i, msg := range messages {
		conversation[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	// Combine the conversation messages into a single string
	var conversationTextBuilder strings.Builder
	for _, msg := range conversation {
		conversationTextBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg["role"], msg["content"]))
	}
	conversationText := conversationTextBuilder.String()

	// Create a prompt that asks the model to generate a title
	prompt := fmt.Sprintf("%s\nGenerate a short and concise title summarizing the conversation in the same language. The title must not contain any quotes. The title should be no longer than 60 characters:", conversationText)

	resp, err := client.CreateCompletion(context.Background(), openai.CompletionRequest{
		Model:       openai.GPT3TextDavinci003,
		Prompt:      prompt,
		Temperature: 0.5,
		MaxTokens:   75,
	})
	if err != nil {
		log.Printf("[GID: %s, threadID: %s] Failed to generate thread title with the error: %v\n", ctx.Interaction.GuildID, threadID, err)
		return
	}

	_, err = ctx.Session.ChannelEditComplex(threadID, &discord.ChannelEdit{
		Name: resp.Choices[0].Text,
	})
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Failed to update thread title with the error: %v\n", ctx.Interaction.GuildID, threadID, err)
	}
}


// The attachUsageInfo function adds usage information to a Discord message.
func attachUsageInfo(s *discord.Session, m *discord.Message, usage openai.Usage, model string) {
	extraInfo := fmt.Sprintf("Completion Tokens: %d, Total: %d%s", usage.CompletionTokens, usage.TotalTokens, generateCost(usage, model))

	utils.DiscordChannelMessageEdit(s, m.ID, m.ChannelID, nil, []*discord.MessageEmbed{
		{
			Footer: &discord.MessageEmbedFooter{
				Text:    extraInfo,
				IconURL: constants.OpenAIBlackIconURL,
			},
		},
	})
}

// The generateCost function calculates the cost of using the GPT model based on the number of prompt and completion tokens used. 
// The cost is returned as a string.
func generateCost(usage openai.Usage, model string) string {
	var cost float64

	switch model {
	case openai.GPT3Dot5Turbo, openai.GPT3Dot5Turbo0301, openai.GPT3Dot5Turbo0613:
		// gpt-3.5-turbo may change over time. Calculating usage assuming gpt-3.5-turbo-0301
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT3Dot5Turbo0613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT3Dot5Turbo0613
	case openai.GPT3Dot5Turbo16K, openai.GPT3Dot5Turbo16K0613:
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT3Dot5Turbo16K0613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT3Dot5Turbo16K0613
	case openai.GPT4, openai.GPT40314, openai.GPT40613:
		// gpt-4 may change over time. Calculating usage assuming gpt-4-0613
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT40613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT40613
	case openai.GPT432K, openai.GPT432K0314, openai.GPT432K0613:
		// gpt-4-32k may change over time. Calculating usage assuming gpt-4-32k-0613
		cost = float64(usage.PromptTokens)*gptPricePerPromptTokenGPT432K0613 + float64(usage.CompletionTokens)*gptPricePerCompletionTokenGPT432K0613
	default:
		// Not implemented
		return ""
	}

	return fmt.Sprintf("\nLLM Cost: $%f", cost)
}
