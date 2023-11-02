package dalle

import (
	"context"
	"log"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)


// The code block contains two middleware functions: imageInteractionResponseMiddleware and imageModerationMiddleware. 
// Middleware functions are functions that are executed before or after the main function of a program and are used to perform additional processing or validation.


// The imageInteractionResponseMiddleware function is used to handle the response to the user's interaction with the bot. 
// The function logs a message indicating that the interaction has been invoked and sends a response to the user indicating that the interaction is being processed.
// If an error occurs during the sending of the response, the function sends an error message to the user indicating that the response failed to send.
func imageInteractionResponseMiddleware(ctx *bot.Context) {
	log.Printf("[GID: %s, i.ID: %s] Image interaction invoked by UserID: %s\n", ctx.Interaction.GuildID, ctx.Interaction.ID, ctx.Interaction.Member.User.ID)

	err := ctx.Respond(&discord.InteractionResponse{
		Type: discord.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("[GID: %s, i.ID: %s] Failed to respond to interactrion with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		return
	}

	ctx.Next()
}

// The imageModerationMiddleware function is used to moderate the generated images to ensure they are appropriate. 
// The function logs a message indicating that the moderation middleware is being performed and extracts the prompt from the bot.Context object. 
// The function then sends a request to the DALL-E API to moderate the prompt. If the prompt violates the usage policies of the DALL-E API, 
// the function sends an error message to the user indicating that the prompt is not allowed. If the prompt is allowed, the function proceeds 
// to the next middleware or the main function.
func imageModerationMiddleware(ctx *bot.Context, client *openai.Client) {
	log.Printf("[GID: %s, i.ID: %s] Performing interaction moderation middleware\n", ctx.Interaction.GuildID, ctx.Interaction.ID)

	var prompt string
	if option, ok := ctx.Options[imageCommandOptionPrompt.String()]; ok {
		prompt = option.StringValue()
	} else {
		// We can't have empty prompt, unfortunately
		// this should not happen, discord prevents empty required options
		log.Printf("[GID: %s, i.ID: %s] Failed to parse prompt option\n", ctx.Interaction.GuildID, ctx.Interaction.ID)
		ctx.Respond(&discord.InteractionResponse{
			Type: discord.InteractionResponseChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: "ERROR: Failed to parse prompt option",
			},
		})
		return
	}

	resp, err := client.Moderations(
		context.Background(),
		openai.ModerationRequest{
			Input: prompt,
		},
	)
	if err != nil {
		// do not block request if moderation api failed
		log.Printf("[GID: %s, i.ID: %s] OpenAI Moderation API request failed with the error: %v\n", ctx.Interaction.GuildID, ctx.Interaction.ID, err)
		ctx.Next()
		return
	}

	if resp.Results[0].Flagged {
		// response was flagged, send error
		log.Printf("[GID: %s, i.ID: %s] Interaction was flagged by Moderation API, prompt: \"%s\"\n", ctx.Interaction.GuildID, ctx.Interaction.ID, prompt)
		ctx.FollowupMessageCreate(ctx.Interaction, true, &discord.WebhookParams{
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "❌ Error",
					Description: "The provided prompt contains text that violates OpenAI's usage policies and is not allowed by their safety system",
					Color:       0xff0000,
				},
			},
		})
		return
	}

	ctx.Next()
}
