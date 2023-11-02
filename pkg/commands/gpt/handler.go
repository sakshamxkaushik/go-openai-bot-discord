package gpt

import (
	"fmt"
	"log"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/constants"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/utils"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const (
	// Discord expects the auto_archive_duration to be one of the following values: 60, 1440, 4320, or 10080,
	// which represent the number of minutes before a thread is automatically archived
	// (1 hour, 1 day, 3 days, or 7 days, respectively).
	gptDiscordThreadAutoArchivewDurationMinutes = 60

	gptInteractionEmbedColor  = 0x000000
	gptPendingMessage         = "⌛ Wait a moment, please..."
	gptContextOptionMaxLength = 1024 // due to discord embed field value limitation
)

// chatGPTHandler handles the chatGPT interaction by parsing the options provided by the user, preparing the cache item, and responding to the interaction.
// If the interaction was invoked in a thread, it is ignored.
// If the prompt option is not provided or is empty, an error message is sent to the user.
// If the model option is provided, it is used to determine the model to use for the conversation.
// If the context-file option is provided, the content of the file is used as the system message for the conversation.
// If the context option is provided, it is used as the system message for the conversation.
// If the provided context exceeds the maximum length, an error message is sent to the user.
// The cache item is then used to generate a response to the user's prompt.
//func chatGPTHandler(ctx *bot.Context, client *openai.Client, messagesCache *MessagesCache) {
	// code block from the selection goes here


func chatGPTHandler(ctx *bot.Context, client *openai.Client, messagesCache *MessagesCache) {
	ch, err := ctx.Session.State.Channel(ctx.Interaction.ChannelID)
	if err == nil && ch.IsThread() {
		// ignore interactions invoked in threads
		log.Printf("[GID: %s, i.ID: %s] Interaction was invoked in the existing thread, ignoring\n", ctx.Interaction.GuildID, ctx.Interaction.ID)
		return
	}

	log.Printf("[GID: %s, i.ID: %s] ChatGPT interaction invoked by UserID: %s\n", ctx.Interaction.GuildID, ctx.Interaction.ID, ctx.Interaction.Member.User.ID)

	err = ctx.Respond(&discord.InteractionResponse{
		Type: discord.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Failed to respond to interactrion with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		return
	}

	var prompt string
	if option, ok := ctx.Options[gptCommandOptionPrompt.string()]; ok {
		prompt = option.StringValue()
	} else {
		// We can't have empty prompt, unfortunately
		// this should not happen, discord prevents empty required options
		log.Printf("[GID: %s, i.ID: %s] Failed to parse prompt option\n", ctx.Interaction.GuildID, ctx.Interaction.ID)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "❌ Error",
					Description: "Failed to parse prompt option",
					Color:       0xff0000,
				},
			},
		})
		return
	}

	fields := make([]*discord.MessageEmbedField, 0, 4)
	fields = append(fields, &discord.MessageEmbedField{
		Value: "\u200B",
	})

	// Determine model
	model := gptDefaultModel
	if option, ok := ctx.Options[gptCommandOptionModel.string()]; ok {
		model = option.StringValue()
		log.Printf("[GID: %s, i.ID: %s] Model provided: %s\n", ctx.Interaction.GuildID, ctx.Interaction.ID, model)
	}

	// Prepare cache item
	cacheItem := &MessagesCacheData{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Model: model,
	}

	// Set context of the conversation as a system message. File option takes precedence
	if option, ok := ctx.Options[gptCommandOptionContextFile.string()]; ok {
		attachmentID := option.Value.(string)
		attachmentURL := ctx.Interaction.ApplicationCommandData().Resolved.Attachments[attachmentID].URL


		// The function then calls the getContentOrURLData function to retrieve the content of the attachment file or the data from the attachment URL. 
		// If an error occurs during the retrieval of the attachment data, the function logs an error message and sends a follow-up message to the Discord API 
		// indicating that the attachment data could not be retrieved.
		
		context, err := getContentOrURLData(ctx.Client, attachmentURL)
		if err != nil {
			log.Printf("[GID: %s, i.ID: %s] Failed to get context file data with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
			ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
				Embeds: []*discord.MessageEmbed{
					{
						Title:       "Failed to get attachment data",
						Description: err.Error(),
						Color:       0xff0000,
					},
				},
			})
			return
		}


		// If the attachment data is successfully retrieved, the function sets the SystemMessage field of the cacheItem struct to a new openai.ChatCompletionMessage struct.
		// The ChatCompletionMessage struct represents a message generated by the OpenAI API and includes information about the message content and role.
		cacheItem.SystemMessage = &openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: context,
		}

		if ok, count := isCacheItemWithinTruncateLimit(cacheItem); !ok {
			// Message exceeds allowed token input from the user
			truncateLimit := count
			if limit := modelTruncateLimit(model); limit != nil {
				truncateLimit = *limit
			}
			
			// The function then appends a new discord.MessageEmbedField to the fields slice. The MessageEmbedField represents a field in a Discord message embed and includes a name and value.
			ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
				Embeds: []*discord.MessageEmbed{
					{
						Title:       "Failed to process context file",
						Description: fmt.Sprintf("Context file is `%d` tokens, which exceeds allowed token limit of `%d` for model `%s`.\nPlease provide a shorter file or use `context` option instead", count, truncateLimit, model),
						Color:       0xff0000,
					},
				},
			})
			log.Printf("[GID: %s, i.ID: %s] User provided context file has %d tokens, which exceeds allowed token limit of `%d` for model `%s`.\n", ctx.Interaction.GuildID, ctx.Interaction.ID, count, truncateLimit, model)
			return
		}
		
		// The name of the field is set to the human-readable string of the gptCommandOptionContextFile option, and the value is set to the attachment URL.
		fields = append(fields, &discord.MessageEmbedField{
			Name:  gptCommandOptionContextFile.humanReadableString(),
			Value: attachmentURL,
		})

		log.Printf("[GID: %s, i.ID: %s] Context file provided: [AID: %s]\n", ctx.Interaction.GuildID, ctx.Interaction.ID, attachmentID)
	} else if option, ok := ctx.Options[gptCommandOptionContext.string()]; ok {
		context := option.StringValue()
		if len(context) >= gptContextOptionMaxLength {
			log.Printf("[GID: %s, i.ID: %s] User-provided context is above limit of %d characters\n", ctx.Interaction.GuildID, ctx.Interaction.ID, gptContextOptionMaxLength)
			ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
				Embeds: []*discord.MessageEmbed{
					{
						Title:       "Failed to process command",
						Description: fmt.Sprintf("Provided context is above the limit of %d characters. Please use `context-file` option instead", gptContextOptionMaxLength),
						Color:       0xff0000,
					},
				},
			})
			return
		}
		cacheItem.SystemMessage = &openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: context,
		}
		fields = append(fields, &discord.MessageEmbedField{
			Name:  gptCommandOptionContext.humanReadableString(),
			Value: context,
		})
		log.Printf("[GID: %s, i.ID: %s] Context provided: %s\n", ctx.Interaction.GuildID, ctx.Interaction.ID, context)
	}

	// Add model info field after context
	fields = append(fields, &discord.MessageEmbedField{
		Name:  gptCommandOptionModel.humanReadableString(),
		Value: model,
	})


	// The function then checks if the gptCommandOptionTemperature option is present in the command options. 
	// If the option is present, the function retrieves the temperature value from the option and sets it as the Temperature field of the cacheItem struct. 
	// The function also appends a new discord.MessageEmbedField to the fields slice with the name set to the human-readable string of the gptCommandOptionTemperature
	// option and the value set to the temperature value.
	if option, ok := ctx.Options[gptCommandOptionTemperature.string()]; ok {
		temp := float32(option.FloatValue())
		cacheItem.Temperature = &temp
		fields = append(fields, &discord.MessageEmbedField{
			Name:  gptCommandOptionTemperature.humanReadableString(),
			Value: fmt.Sprintf("%g", temp),
		})
		log.Printf("[GID: %s, i.ID: %s] Temperature provided: %g\n", ctx.Interaction.GuildID, ctx.Interaction.ID, temp)
	}
	// The function then responds to the interaction with a reference and user ping. The response includes a message embed with a description of the prompt,
	// an author field indicating the user who made the request, and a list of fields that includes the selected temperature value.
	// Respond to interaction with a reference and user ping
	_, err = ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
		Embeds: []*discord.MessageEmbed{
			{
				Description: prompt,
				Color:       gptInteractionEmbedColor,
				Author: &discord.MessageEmbedAuthor{
					Name:         "OpenAI chat request by " + ctx.Interaction.Member.User.Username,
					IconURL:      ctx.Interaction.Member.User.AvatarURL("32"),
					ProxyIconURL: constants.OpenAIBlackIconURL,
				},
				Fields: fields,
			},
		},
	}) 
	// The function then logs a message indicating that the temperature was provided and includes the guild ID and interaction ID.
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Failed to respond to interactrion with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "Failed to process command",
					Description: err.Error(),
					Color:       0xff0000,
				},
			},
		})
		return
	}

	// Get interaction ID so we can create a thread on top of it
	m, err := ctx.Response()
	if err != nil {
		// Without interaction reference we cannot create a thread with the response of ChatGPT
		// Maybe in the future just try to post a new message instead, but for now just cancel
		log.Printf("[GID: %s, i.ID: %s] Failed to get interaction reference with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		ctx.Edit(fmt.Sprintf("Failed to get interaction reference with error: %v", err))
		return
	}

	ch, err = ctx.Session.State.Channel(m.ChannelID)
	if err != nil || ch.IsThread() {
		log.Printf("[GID: %s, i.ID: %s] Interaction reply was in a thread, or there was an error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		return
	}

	thread, err := ctx.Session.MessageThreadStartComplex(m.ChannelID, m.ID, &discord.ThreadStart{
		Name:                "New chat",
		AutoArchiveDuration: gptDiscordThreadAutoArchivewDurationMinutes,
		Invitable:           false,
	})

	if err != nil {
		// Without thread we cannot reply our answer
		log.Printf("[GID: %s, i.ID: %s] Failed to create a thread with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		return
	}

	// Lock the thread while we are generating ChatGPT answser
	utils.ToggleDiscordThreadLock(ctx.Session, thread.ID, true)

	// add user to the thread
	ctx.ThreadMemberAdd(thread.ID, ctx.Interaction.Member.User.ID)

	channelMessage, err := utils.DiscordChannelMessageSend(ctx.Session, thread.ID, gptPendingMessage, nil)
	if err != nil {
		// Without reply  we cannot edit message with the response of ChatGPT
		// Maybe in the future just try to post a new message instead, but for now just cancel
		log.Printf("[GID: %s, i.ID: %s] Failed to reply in the thread with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		return
	}

	messagesCache.Add(thread.ID, cacheItem)

	log.Printf("[GID: %s, i.ID: %s] ChatGPT Request invoked with [Model: %s]. Current cache size: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, cacheItem.Model, len(cacheItem.Messages))
	resp, err := sendChatGPTRequest(client, cacheItem)
	if err != nil {
		// ChatGPT failed for whatever reason, tell users about it
		log.Printf("[GID: %s, i.ID: %s] OpenAI request ChatCompletion failed with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		emptyString := ""
		utils.DiscordChannelMessageEdit(ctx.Session, channelMessage.ID, channelMessage.ChannelID, &emptyString, []*discord.MessageEmbed{
			{
				Title:       "❌ OpenAI API failed",
				Description: err.Error(),
				Color:       0xff0000,
			},
		})
		return
	}



	//The code block is used to edit a message in a Discord channel with the response generated by the OpenAI API. 
	// The function first defers a call to the ToggleDiscordThreadLock function to release the thread lock. 
	// The function then calls the generateThreadTitleBasedOnInitialPrompt function to generate a thread title based on the initial prompt.
	// Unlock the thread at the end
	defer utils.ToggleDiscordThreadLock(ctx.Session, thread.ID, false)

	go generateThreadTitleBasedOnInitialPrompt(ctx, client, thread.ID, cacheItem.Messages)

	log.Printf("[GID: %s, i.ID: %s] ChatGPT Request [Model: %s] responded with a usage: [PromptTokens: %d, CompletionTokens: %d, TotalTokens: %d]\n", ctx.Interaction.GuildID, ctx.Interaction.ID, cacheItem.Model, resp.usage.PromptTokens, resp.usage.CompletionTokens, resp.usage.TotalTokens)

	// The function logs a message indicating the details of the ChatGPT request, including the guild ID, interaction ID, model, and usage statistics. 
	// The function then splits the response content into multiple messages using the splitMessage function.

	// The function then calls the DiscordChannelMessageEdit function to edit the original message in the Discord channel with the first message in the messages slice. 
	// If an error occurs during the editing of the message, the function logs an error message and sends a follow-up message to the Discord API indicating that an error occurred.

	// Split message into multiple messages
	messages := splitMessage(resp.content)
	err = utils.DiscordChannelMessageEdit(ctx.Session, channelMessage.ID, channelMessage.ChannelID, &messages[0], nil)
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Discord API failed with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		emptyString := ""
		utils.DiscordChannelMessageEdit(ctx.Session, channelMessage.ID, channelMessage.ChannelID, &emptyString, []*discord.MessageEmbed{
			{
				Title:       "❌ Discord API Error",
				Description: err.Error(),
				Color:       0xff0000,
			},
		})
		return
	}

	if len(messages) > 1 {
		// if there are more messages, send them as a thread reply
		for _, message := range messages[1:] {
			channelMessage, err = utils.DiscordChannelMessageSend(ctx.Session, thread.ID, message, nil)
			if err != nil {
				log.Printf("[GID: %s, i.ID: %s] Discord API failed with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
			}
		}
	}

	attachUsageInfo(ctx.Session, channelMessage, resp.usage, cacheItem.Model)
}
