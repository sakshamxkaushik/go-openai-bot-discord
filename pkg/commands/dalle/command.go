package dalle

import (
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const commandName = "dalle"


// The Command function is defined in this code block, which returns a bot.Command object. 
// The bot.Command object represents a command that can be executed by a Discord bot. 
// The Command function takes an openai.Client object as input, which is used to interact with the DALL-E API.
func Command(client *openai.Client) *bot.Command {
	numberOptionMinValue := 1.0
	return &bot.Command{

		// The Name and Description fields of the bot.Command object are set to the constant commandName and a description of the command, respectively.
		// The Options field of the bot.Command object is set to an array of discord.ApplicationCommandOption objects, 
		// which represent the command options that can be passed to the command.
		Name:        commandName,	
		Description: "Generate creative images from textual descriptions using OpenAI Dalle 2",
		Options: []*discord.ApplicationCommandOption{
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionPrompt.String(),
				Description: "A text description of the desired image",
				Required:    true,
			},
			{
				Type:        discord.ApplicationCommandOptionString,
				Name:        imageCommandOptionSize.String(),
				Description: "The size of the generated images",
				Required:    false,
				Choices: []*discord.ApplicationCommandOptionChoice{
					{
						Name:  openai.CreateImageSize256x256 + " (Default)",
						Value: openai.CreateImageSize256x256,
					},
					{
						Name:  openai.CreateImageSize512x512,
						Value: openai.CreateImageSize512x512,
					},
					{
						Name:  openai.CreateImageSize1024x1024,
						Value: openai.CreateImageSize1024x1024,
					},
				},
			},
			{
				Type:        discord.ApplicationCommandOptionInteger,
				Name:        imageCommandOptionNumber.String(),
				Description: "The number of images to generate (default 1, max 4)",
				MinValue:    &numberOptionMinValue,
				MaxValue:    4,
				Required:    false,
			},
		},
		// The Handler field of the bot.Command object is set to a bot.HandlerFunc object, which is a function that handles the execution of the command.
		// The Handler function calls the imageHandler function, passing in the bot.Context object and the openai.Client object as arguments.
		Handler: bot.HandlerFunc(func(ctx *bot.Context) {
			imageHandler(ctx, client)
		}),

		// The Middlewares field of the bot.Command object is set to an array of bot.Handler objects,
		// which represent middleware functions that are executed before the command is executed. 
		// The first middleware function is imageInteractionResponseMiddleware, which handles the response to the user's interaction with the command. 
		// The second middleware function is imageModerationMiddleware, which moderates the generated images to ensure they are appropriate.
		Middlewares: []bot.Handler{
			bot.HandlerFunc(imageInteractionResponseMiddleware),   	// When a user interacts with the command, the imageInteractionResponseMiddleware 
			bot.HandlerFunc(func(ctx *bot.Context) {				// function is executed before the command is executed. The function takes a 
				imageModerationMiddleware(ctx, client)				// bot.HandlerFunc object as input, which represents the function that handles 
			}),														// the execution of the command.
		},
	}
}
