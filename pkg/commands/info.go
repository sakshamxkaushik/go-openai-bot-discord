package commands

import (
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/constants"
	discord "github.com/bwmarrin/discordgo"
)

const (
	infoCommandName = "info"
)


// This File is used to define an info command for the bot that displays information about the current version of the bot.

// The infoHandler function is used to handle the info command for the Discord bot. The function takes a bot.Context pointer as an argument and sends a response to the Discord API. 
// The response includes an actions row with a single link button that points to the project's GitHub repository, and an embed that displays the current version of the bot.
func infoHandler(ctx *bot.Context) {
	ctx.Respond(&discord.InteractionResponse{
		Type: discord.InteractionResponseChannelMessageWithSource,
		Data: &discord.InteractionResponseData{
			// Note: only visible to the user who invoked the command
			Flags: discord.MessageFlagsEphemeral,
			// Content: "Surprise!",
			Components: []discord.MessageComponent{
				discord.ActionsRow{
					Components: []discord.MessageComponent{
						&discord.Button{
							Label: "Source code",
							Style: discord.LinkButton,
							URL:   "https://github.com/akhilsharma90/go-openai-bot-discord",
						},
					},
				},
			},
			Embeds: []*discord.MessageEmbed{
				{
					Title:       "Bot Version",
					Description: "Version: " + constants.Version,
					Color:       0x00bfff,
				},
			},
		},
	})
}

// The InfoCommand function returns a bot.Command struct that represents the info command for the Discord bot.
//  command is named info and is used to show information about the current version of the Open AI bot. 
//  DMPermission field is set to true, which means the command can be used in direct messages. 
//  DefaultMemberPermissions field is set to discord.PermissionViewChannel, which means that all members can view the channel.
func InfoCommand() *bot.Command {
	return &bot.Command{
		Name:                     infoCommandName,
		Description:              "Show information about current version of Open AI bot",
		DMPermission:             true,
		DefaultMemberPermissions: discord.PermissionViewChannel,
		Handler:                  bot.HandlerFunc(infoHandler), // The Handler field is set to the infoHandler function, which is used to handle the command.
	}
}
