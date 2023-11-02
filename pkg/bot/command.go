package bot

import (
	discord "github.com/bwmarrin/discordgo"
)

//in this file, we will define a se of GO types and methods related to handling and
//processing application commands specifically for our discord bot

// we first create a handler interface that is intended for handling commands
type Handler interface {
	HandleCommand(ctx *Context)
}

// we define a type here, handlerfunc which has a function that takes in context
type HandlerFunc func(ctx *Context)

func (f HandlerFunc) HandleCommand(ctx *Context) { f(ctx) }


// messageHandler interface is used for handling message-related commands.

type MessageHandler interface {
	HandleMessageCommand(ctx *MessageContext)
}
// messageHandlerFunc is an adapter type that allows functions to implement the MessageHandler interface.
type MessageHandlerFunc func(ctx *MessageContext)

// this function handles message related commands
func (f MessageHandlerFunc) HandleMessageCommand(ctx *MessageContext) { f(ctx) }

// the struct that defines how an application command looks like
type Command struct {
	Name                     string      
	Description              string
	DMPermission             bool		// Indicates if the command is allowed in direct messages.
	DefaultMemberPermissions int64
	Options                  []*discord.ApplicationCommandOption
	Type                     discord.ApplicationCommandType

	Handler        Handler			// Command handler
	Middlewares    []Handler		// Middleware handlers for the command
	MessageHandler MessageHandler   // Message command handler (for message-based interactions).
	
	//the subcommands is of type router, which can be used to handle subcommands
	SubCommands *Router
}


// ApplicationCommand converts the Command struct into a discord.ApplicationCommand which is used to register the command with discord.
// The method then iterates over the SubCommands field of the Command struct, which is a List of Command structs.
func (cmd Command) ApplicationCommand() *discord.ApplicationCommand {
	applicationCommand := &discord.ApplicationCommand{
		Name:                     cmd.Name,
		Description:              cmd.Description,
		DMPermission:             &cmd.DMPermission,
		DefaultMemberPermissions: &cmd.DefaultMemberPermissions,
		Options:                  cmd.Options,
		Type:                     cmd.Type,
	}
	for _, subcommand := range cmd.SubCommands.List() {
		applicationCommand.Options = append(applicationCommand.Options, subcommand.ApplicationCommandOption())
	}
	return applicationCommand
}



// ApplicationCommandOption method is called on the Command to get an ApplicationCommandOption struct. 
// ApplicationCommandOption converts the Command struct into a discord.ApplicationCommandOption which is used to register the command with discord.
func (cmd Command) ApplicationCommandOption() *discord.ApplicationCommandOption {
	applicationCommand := cmd.ApplicationCommand()	// The ApplicationCommandOption method calls the ApplicationCommand method on the Command struct to get an ApplicationCommand struct. 
	typ := discord.ApplicationCommandOptionSubCommand


	// ApplicationCommand struct. It then sets the typ variable to discord.ApplicationCommandOptionSubCommand if the SubCommands 
	// field of the Command struct is empty, and to discord.ApplicationCommandOptionSubCommandGroup otherwise.

	if cmd.SubCommands != nil && cmd.SubCommands.Count() != 0 {
		typ = discord.ApplicationCommandOptionSubCommandGroup
	}


	// Finally, the ApplicationCommandOption struct is returned. 
	// This method is used to convert a Command struct to an ApplicationCommandOption struct.
	return &discord.ApplicationCommandOption{
		Name:        applicationCommand.Name,
		Description: applicationCommand.Description,
		Options:     applicationCommand.Options,
		Type:        typ,
	}
}
