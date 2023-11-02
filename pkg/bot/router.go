package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	discord "github.com/bwmarrin/discordgo"
)

// Router manages application commands and their handlers.
type Router struct {
	commands           map[string]*Command
	registeredCommands []*discord.ApplicationCommand
}

// The NewRouter function creates a new Router struct with an initial set of commands.
func NewRouter(initial []*Command) (r *Router) {
	r = &Router{commands: make(map[string]*Command, len(initial))}
	for _, cmd := range initial {
		r.Register(cmd)
	}

	return
}

// The Register function adds a command to the commands map.
func (r *Router) Register(cmd *Command) {
	if _, ok := r.commands[cmd.Name]; !ok {
		r.commands[cmd.Name] = cmd
	}
}

// The Get function retrieves a command from the commands map by name.
func (r *Router) Get(name string) *Command {
	if r == nil {
		return nil
	}
	return r.commands[name]
}

// The List function returns a slice of all commands in the commands map.
func (r *Router) List() (list []*Command) {
	if r == nil {
		return nil
	}

	for _, c := range r.commands {
		list = append(list, c)
	}
	return
}

// The Count function returns the number of commands in the commands map.
func (r *Router) Count() (c int) {
	if r == nil {
		return 0
	}
	return len(r.commands)
}

// The getSubcommand function is used to retrieve a subcommand from a command based on an ApplicationCommandInteractionDataOption.
func (r *Router) getSubcommand(cmd *Command, opt *discord.ApplicationCommandInteractionDataOption, parent []Handler) (*Command, *discord.ApplicationCommandInteractionDataOption, []Handler) {
	if cmd == nil {
		return nil, nil, nil
	}

	subcommand := cmd.SubCommands.Get(opt.Name)
	switch opt.Type {
	case discordgo.ApplicationCommandOptionSubCommand:
		return subcommand, opt, append(parent, append(subcommand.Middlewares, subcommand.Handler)...)
	case discordgo.ApplicationCommandOptionSubCommandGroup:
		return r.getSubcommand(subcommand, opt.Options[0], append(parent, subcommand.Middlewares...))
	}

	return cmd, nil, append(parent, cmd.Handler)
}

// The getMessageHandlers function is used to retrieve all message handlers for a command and its subcommands.
func (r *Router) getMessageHandlers(cmd *Command) []MessageHandler {
	var handlers []MessageHandler

	if cmd.MessageHandler != nil {
		handlers = append(handlers, cmd.MessageHandler)
	}

	if cmd.SubCommands != nil {
		for _, cmd := range cmd.SubCommands.List() {
			handlers = append(handlers, r.getMessageHandlers(cmd)...)
		}
	}

	return handlers
}

// The HandleInteraction function is used to handle interaction events in the Discord bot.
// It retrieves the command from the commands map based on the interaction data, and then retrieves the appropriate subcommand based on the interaction options.
func (r *Router) HandleInteraction(s *discord.Session, i *discord.InteractionCreate) {
	if i.Type != discord.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	cmd := r.Get(data.Name)
	if cmd == nil {
		return
	}

	var parent *discord.ApplicationCommandInteractionDataOption
	handlers := append(cmd.Middlewares, cmd.Handler)
	if len(data.Options) != 0 {
		cmd, parent, handlers = r.getSubcommand(cmd, data.Options[0], cmd.Middlewares)
	}

	// It then creates a new Context struct and calls the Next method to execute the command's handlers.
	if cmd != nil {
		ctx := NewContext(s, cmd, i.Interaction, parent, handlers)
		ctx.Next()
	}
}

// The HandleMessage function is used to handle message events in the Discord bot.
// It retrieves all message handlers for each command in the commands map and creates a new MessageContext struct for each handler.
// It then calls the Next method to execute the message handlers.
func (r *Router) HandleMessage(s *discord.Session, m *discord.MessageCreate) {
	for _, cmd := range r.commands {
		handlers := r.getMessageHandlers(cmd)
		if len(handlers) > 0 {
			ctx := NewMessageContext(s, cmd, m.Message, handlers)
			ctx.Next()
		}
	}
}

// this is called in the bot.go file, takes in the particular discord session and the guildID
// syncs the commands to the bot

// Sync registers all the application commands in the router to the given guild using the provided session.
// It returns an error if the session's user is nil or if there is an error in registering the commands.
func (r *Router) Sync(s *discord.Session, guild string) (err error) {
	//first we will check the user state in the session
	//if it is nil, we will print out an error
	if s.State.User == nil {
		return fmt.Errorf("cannot determine application id")
	}

	//then we create a variable called commands which is of type ApplicationCommand struct
	//this struct is mentioned in the command.go file
	var commands []*discord.ApplicationCommand
	//we will first range over all the commands one by one and each command is captured in c
	//we will append all these specific application commands to the commands variable
	for _, c := range r.commands {
		commands = append(commands, c.ApplicationCommand())
	}
	//the applicationcommandbulkoverwrite function helps us to register the commands that are
	//now present in the commands variable, we just need to pass the user's id, guild and the list of commands

	r.registeredCommands, err = s.ApplicationCommandBulkOverwrite(s.State.User.ID, guild, commands)
	return
}

// The ClearCommands function takes a discord.Session pointer and a string representing a guild ID as arguments.
// The function is used to delete all registered commands for the bot from the Discord API.
func (r *Router) ClearCommands(s *discord.Session, guild string) (errors []error) { // The function first checks if the user associated with the session is not nil.
	if s.State.User == nil { // If it is nil, it returns an error.
		return []error{fmt.Errorf("cannot determine application id")}
	}

	// The function then iterates over all registered commands in the registeredCommands slice of the Router struct.
	for _, v := range r.registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, guild, v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
