package commands

import (
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/commands/gpt"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const chatCommandName = "chat"

// The ChatCommandParams struct defines parameters for the ChatCommand function. 
// These parameters include an OpenAI client, a slice of OpenAI completion models, a cache for GPT messages, and a cache for ignored channels.
type ChatCommandParams struct {
	OpenAIClient           *openai.Client
	OpenAICompletionModels []string
	GPTMessagesCache       *gpt.MessagesCache
	IgnoredChannelsCache   *gpt.IgnoredChannelsCache
}


// The ChatCommand function returns a bot.Command struct that represents a chat command for the Discord bot. 
// The command is named chat and is used to start a conversation with an AI language model. 
func ChatCommand(params *ChatCommandParams) *bot.Command {      // The ChatCommand function is used to define a chat command for the bot that starts a conversation with an AI language model.
	return &bot.Command{				     					
		Name:                     chatCommandName,
		Description:              "Start conversation with LLM",
		DMPermission:             false,						  // The DMPermission field is set to false, which means the command can only be used in guild channels. 
		DefaultMemberPermissions: discord.PermissionViewChannel,  // The DefaultMemberPermissions field is set to discord.PermissionViewChannel, which means that all members can view the channel. 
		Type:                     discord.ChatApplicationCommand, // The Type field is set to discord.ChatApplicationCommand, which means that the command is a chat command.


		// The SubCommands field is set to a bot.Router struct that contains a single subcommand, which is defined by the gpt.Command function. 
		// The gpt.Command function takes the OpenAI client, the OpenAI completion models, the GPT messages cache, and the ignored channels cache as arguments, and returns a bot.
		SubCommands: bot.NewRouter([]*bot.Command{
			gpt.Command(params.OpenAIClient, params.OpenAICompletionModels, params.GPTMessagesCache, params.IgnoredChannelsCache), // Command struct that represents a GPT command for the Discord bot.

		}),				//  The gpt.Command function is used to define a subcommand for the chat command that uses the GPT language model.
	}
}
