package gpt

import (
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

var gptDefaultModel = openai.GPT3Dot5Turbo

const commandName = "gpt"

// The Command function is used to define a command for the Discord bot. The function takes several arguments, including a *openai.Client pointer, 

// he function takes several arguments, including a *openai.Client pointer, a slice of strings representing completion models, a *MessagesCache pointer,
// and an *IgnoredChannelsCache pointer. The function creates a new bot.Command struct and sets its Name and Description fields to "gpt" and "Start conversation with ChatGPT", respectively.
func Command(client *openai.Client, completionModels []string, messagesCache *MessagesCache, ignoredChannelsCache *IgnoredChannelsCache) *bot.Command {
	temperatureOptionMinValue := 0.0
	opts := []*discord.ApplicationCommandOption{		// The function then creates a slice of *discord.ApplicationCommandOptions representing the different options 
		{												// that can be used with the command. The options include a prompt, context, context file, model, and temperature. 
			Type:        discord.ApplicationCommandOptionString,	// // The function sets the Required field to true for the prompt option and false for the other options.
			Name:        gptCommandOptionPrompt.string(),
			Description: "ChatGPT prompt",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        gptCommandOptionContext.string(),
			Description: "Sets context that guides the AI assistant's behavior during the conversation",
			Required:    false,
		},
		{
			Type:        discord.ApplicationCommandOptionAttachment,
			Name:        gptCommandOptionContextFile.string(),
			Description: "File that sets context that guides the AI assistant's behavior during the conversation",
			Required:    false,
		},
	}
	numberOfModels := len(completionModels)
	if numberOfModels > 0 {
		gptDefaultModel = completionModels[0] // set first model as default one
	}
	if numberOfModels > 1 {
		var modelChoices []*discord.ApplicationCommandOptionChoice		// If there is more than one completion model, the function creates a slice of *discord.ApplicationCommandOptionChoices
		for i, model := range completionModels {						// representing the different models and adds it to the options slice. The function sets the first model as the default model.
			name := model
			if i == 0 {
				name += " (Default)"
			}
			modelChoices = append(modelChoices, &discord.ApplicationCommandOptionChoice{
				Name:  name,
				Value: model,
			})
		}
		opts = append(opts, &discord.ApplicationCommandOption{
			Type:        discord.ApplicationCommandOptionString,
			Name:        gptCommandOptionModel.string(),
			Description: "GPT model",
			Required:    false,
			Choices:     modelChoices,
		})
	}
	opts = append(opts, &discord.ApplicationCommandOption{
		Type:        discord.ApplicationCommandOptionNumber,
		Name:        gptCommandOptionTemperature.string(),
		Description: "What sampling temperature to use, between 0.0 and 2.0. Lower - more focused and deterministic",
		MinValue:    &temperatureOptionMinValue,
		MaxValue:    2.0,
		Required:    false,
	})

	// The function then creates a new bot.HandlerFunc and bot.MessageHandlerFunc to handle the command and message events, respectively. 
	// The chatGPTHandler and chatGPTMessageHandler functions are called with the appropriate arguments to handle the events.
	return &bot.Command{
		Name:        commandName,
		Description: "Start conversation with ChatGPT", /// Command struct and sets its Name and Description fields to "gpt" and "Start conversation with ChatGPT", respectively.
		Options:     opts,
		Handler: bot.HandlerFunc(func(ctx *bot.Context) {
			chatGPTHandler(ctx, client, messagesCache)  
		}),
		MessageHandler: bot.MessageHandlerFunc(func(ctx *bot.MessageContext) {
			chatGPTMessageHandler(ctx, client, messagesCache, ignoredChannelsCache) 
			// The chatGPTHandler function is used to handle the gpt command for the Discord bot.
			// The function takes a bot.Context pointer, a *openai.Client pointer, and a *MessagesCache pointer as arguments.
		}),
	}
}
