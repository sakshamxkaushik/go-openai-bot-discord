package commands

import (
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/commands/dalle"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const imageCommandName = "image"


//The ImageCommand function takes an OpenAI client as an argument and returns a bot.Command struct that represents an image command for the Discord bot. 
//  Command is named image and is used to generate creative images from textual descriptions. 
//  DMPermission field is set to false, which means the command can only be used in guild channels. 
//  DefaultMemberPermissions field is set to discord.PermissionViewChannel, which means that all members can view the channel.

func ImageCommand(client *openai.Client) *bot.Command {
	return &bot.Command{
		Name:                     imageCommandName,
		Description:              "Generate creative images from textual descriptions",
		DMPermission:             false,
		DefaultMemberPermissions: discord.PermissionViewChannel,
		// The SubCommands field is set to a bot.Router struct that contains a single subcommand, which is defined by the dalle.Command function. 
		// The dalle.Command function takes the OpenAI client as an argument and returns a bot.Command struct that represents a DALL-E command for the Discord bot.
		SubCommands: bot.NewRouter([]*bot.Command{
			dalle.Command(client),
		}),
	}
}

// The ImageCommand function is used to define an image command for the bot that generates creative images from textual descriptions. 
//  DALL-E command is a subcommand of the image command and is used to generate images using the DALL-E image generation model. 
//  dalle.Command function likely defines the behavior of the DALL-E command, including how it interacts with the OpenAI client and how it generates images from textual descriptions.