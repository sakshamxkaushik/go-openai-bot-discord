package gpt

import (
	"strings"

	"github.com/sashabaranov/go-openai"
)

const discordMaxMessageLength = 2000

// splitMessage, is used to split a message into multiple messages that are short enough to be sent as Discord messages. 
// The function takes a message string as input and returns a slice of strings representing the split messages.

func splitMessage(message string) []string {
	if len(message) <= discordMaxMessageLength {
		// the message is short enough to be sent as is
		return []string{message}
	}



	// If the length of the input message is less than or equal to the maximum Discord message length, the function returns a slice containing the input message.
	//  Otherwise, the function splits the message by whitespace and iterates over the resulting words. 
	// The function adds each word to a currentMessage string until the length of the currentMessage string plus the length of the next word exceeds the maximum Discord message length.
	// At that point, the function appends the currentMessage string to the messageParts slice and starts a new currentMessage string with the next word. 
	// The function continues this process until all words have been processed and returns the messageParts slice.
	// split the message by whitespace
	words := strings.Fields(message)
	var messageParts []string
	currentMessage := ""
	for _, word := range words {
		if len(currentMessage)+len(word)+1 > discordMaxMessageLength {
			// start a new message if adding the current word exceeds the maximum length
			messageParts = append(messageParts, currentMessage)
			currentMessage = word + " "
		} else {
			// add the current word to the current message
			currentMessage += word + " "
		}
	}
	// add the last message to the list of message parts
	messageParts = append(messageParts, currentMessage)

	return messageParts
}


// The second function, reverseMessages, is used to reverse the order of a slice of openai.ChatCompletionMessage structs. 
// The function takes a pointer to a slice of openai.ChatCompletionMessage structs as input and modifies the slice in place.
func reverseMessages(messages *[]openai.ChatCompletionMessage) {
	length := len(*messages)
	for i := 0; i < length/2; i++ {
		(*messages)[i], (*messages)[length-i-1] = (*messages)[length-i-1], (*messages)[i]
	}
}


// The function iterates over the first half of the slice and swaps each element with its corresponding element from the second half of the slice. This effectively reverses the order of the slice.
// The splitMessage function is used to split long messages into multiple messages that can be sent as Discord messages, while the reverseMessages function is used to reverse the order of a slice of openai.ChatCompletionMessage structs.

