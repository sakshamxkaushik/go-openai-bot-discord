package utils

import (
	"log"

	discord "github.com/bwmarrin/discordgo"
)

// ToggleThreadLock locks or unlocks a Discord thread, based on the 'locked' parameter.
// If the thread is already locked/unlocked, the function does nothing.
// The function takes a discord.Session pointer, a string representing a channel ID, and a bool representing whether the thread should be locked or unlocked as arguments.
// The function calls the ChannelEditComplex function on the session to edit the channel and set the Locked field to the locked parameter.
// If an error occurs during the editing process, the function logs an error message with the channel ID and the error.
func ToggleDiscordThreadLock(s *discord.Session, channelID string, locked bool) {
	_, err := s.ChannelEditComplex(channelID, &discord.ChannelEdit{
		Locked: &locked,
	})
	if err != nil {
		log.Printf("[CHID: %s] Failed to lock/unlock Thread with the error: %v\n", channelID, err)
	}
}

// Sends a message to a specified Discord channel, either as a reply to another message if a message reference is provided or as a standalone message if the message reference is nil
// The function takes a discord.Session pointer, a string representing a channel ID, a string representing the message content, and a *discord.MessageReference pointer representing
// the message reference as arguments. If the message reference is not nil, the function calls the ChannelMessageSendReply function on the session to send the message as a reply to
// the referenced message. If the message reference is nil, the function calls the ChannelMessageSend function on the session to send the message as a standalone message.
// The function returns a *discord.Message pointer and an error.
func DiscordChannelMessageSend(s *discord.Session, channelID string, content string, messageReference *discord.MessageReference) (m *discord.Message, err error) {
	if messageReference != nil {
		m, err = s.ChannelMessageSendReply(channelID, content, messageReference)
	} else {
		m, err = s.ChannelMessageSend(channelID, content)
	}
	return
}

//The DiscordChannelMessageEdit function is used to edit a message in a specified Discord channel. The function takes a discord.Session pointer, a string representing a message ID,
// a string representing a channel ID, a *string representing the message content, and a slice of *discord.MessageEmbed pointers representing the message embeds as arguments.

func DiscordChannelMessageEdit(s *discord.Session, messageID string, channelID string, content *string, embeds []*discord.MessageEmbed) error {
	_, err := s.ChannelMessageEditComplex( // The function calls the ChannelMessageEditComplex
		&discord.MessageEdit{ // function on the session to edit the message
			Content: content,   // and set the Content
			Embeds:  embeds,    //  and Embeds fields to the provided content and embeds.
			ID:      messageID, // The ID field is set to the messageID parameter
			Channel: channelID, //, and the Channel field is set to the channelID parameter.
		},
	)
	return err // If an error occurs during the editing process, the function returns the error.
}
