package bot

import (
	discord "github.com/bwmarrin/discordgo"
)

// OptionsMap is an alias for a map that stores interaction options.
type OptionsMap = map[string]*discord.ApplicationCommandInteractionDataOption



// Context represents the context of a Discord bot command or interaction.
// The Context struct contains several fields, including a Session field, which is a pointer to a discord.Session struct, 
// a Caller field, which is a pointer to a Command struct, an Interaction field, which is a pointer to a discord.Interaction struct, 
// an Options field, which is an OptionsMap, and a handlers field, which is a slice of Handler interfaces.
type Context struct {
	*discord.Session
	Caller      *Command
	Interaction *discord.Interaction
	Options     OptionsMap

	handlers []Handler
}

// makeOptionMap function is defined to create an OptionsMap from a slice of discord.ApplicationCommandInteractionDataOption structs.
// The function iterates over the slice and adds each option to the map with its Name field as the key.

func makeOptionMap(options []*discord.ApplicationCommandInteractionDataOption) (m OptionsMap) {
	m = make(OptionsMap, len(options))

	for _, option := range options {
		m[option.Name] = option
	}

	return
}

// NewContext creates a new context for a command invocation.
// It takes in a discord session, the command caller, the interaction data, the parent option data,
// and a slice of handlers. It returns a pointer to a new context.
func NewContext(s *discord.Session, caller *Command, i *discord.Interaction, parent *discord.ApplicationCommandInteractionDataOption, handlers []Handler) *Context {
	options := i.ApplicationCommandData().Options
	if parent != nil {
		options = parent.Options
	}
	return &Context{
		Session:     s,
		Caller:      caller,
		Interaction: i,
		Options:     makeOptionMap(options),

		handlers: handlers,
	}
}


// Respond sends a response to the interaction.
func (ctx *Context) Respond(response *discord.InteractionResponse) error {
	return ctx.Session.InteractionRespond(ctx.Interaction, response)
}


// Edit updates the response content.
func (ctx *Context) Edit(content string) error {
	_, err := ctx.Session.InteractionResponseEdit(ctx.Interaction, &discord.WebhookEdit{
		Content: &content,
	})
	return err
}


// Response retrieves the original interaction response.
func (ctx *Context) Response() (*discord.Message, error) {
	return ctx.Session.InteractionResponse(ctx.Interaction)
}


// Next executes the next handler in the chain.
func (ctx *Context) Next() {
	if ctx.handlers == nil || len(ctx.handlers) == 0 {
		return
	}

	handler := ctx.handlers[0]
	ctx.handlers = ctx.handlers[1:]

	handler.HandleCommand(ctx)
}

// MessageContext represents the context in which a message-related command is executed.
type MessageContext struct {
	*discord.Session
	Caller  *Command
	Message *discord.Message

	handlers []MessageHandler
}


// NewMessageContext creates a new MessageContext instance.
func NewMessageContext(s *discord.Session, caller *Command, m *discord.Message, handlers []MessageHandler) *MessageContext {
	return &MessageContext{
		Session: s,
		Caller:  caller,
		Message: m,

		handlers: handlers,
	}
}

// Reply sends a reply message in the same channel as the original message.
func (ctx *MessageContext) Reply(content string) (m *discord.Message, err error) {
	m, err = ctx.Session.ChannelMessageSendReply(
		ctx.Message.ChannelID,
		content,
		ctx.Message.Reference(),
	)
	return
}



// EmbedReply sends a reply message with an embed in the same channel as the original message.
func (ctx *MessageContext) EmbedReply(embed *discord.MessageEmbed) (m *discord.Message, err error) {
	m, err = ctx.Session.ChannelMessageSendEmbedReply(
		ctx.Message.ChannelID,
		embed,
		ctx.Message.Reference(),
	)
	return
}

// AddReaction adds a reaction to the original message.
func (ctx *MessageContext) AddReaction(emojiID string) error {
	return ctx.Session.MessageReactionAdd(ctx.Message.ChannelID, ctx.Message.ID, emojiID)
}


// RemoveReaction removes a reaction from the original message.
func (ctx *MessageContext) RemoveReaction(emojiID string) error {
	return ctx.Session.MessageReactionsRemoveEmoji(ctx.Message.ChannelID, ctx.Message.ID, emojiID)
}


// ChannelTyping indicates that the bot is typing in the channel.
func (ctx *MessageContext) ChannelTyping() error {
	return ctx.Session.ChannelTyping(ctx.Message.ChannelID)
}


// Next executes the next handler in the chain.
func (ctx *MessageContext) Next() {
	if ctx.handlers == nil || len(ctx.handlers) == 0 {
		return
	}

	handler := ctx.handlers[0]
	ctx.handlers = ctx.handlers[1:]

	handler.HandleMessageCommand(ctx)
}
