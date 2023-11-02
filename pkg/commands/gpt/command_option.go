package gpt

import "fmt"


// The gptCommandOptionType type is an enumeration that represents the different types of command options that can be used by the bot. 
// The enumeration includes options for a prompt, context, context file, model, and temperature. Each option is assigned a unique integer value.
type gptCommandOptionType uint8

const (
	gptCommandOptionPrompt      gptCommandOptionType = 1
	gptCommandOptionContext     gptCommandOptionType = 2
	gptCommandOptionContextFile gptCommandOptionType = 3
	gptCommandOptionModel       gptCommandOptionType = 4
	gptCommandOptionTemperature gptCommandOptionType = 5
)


// The string function is a method on the gptCommandOptionType type that returns a string representation of the option. 
// The function uses a switch statement to return the appropriate string representation based on the value of the option.
func (t gptCommandOptionType) string() string {
	switch t {
	case gptCommandOptionPrompt:
		return "prompt"
	case gptCommandOptionContext:
		return "context"
	case gptCommandOptionContextFile:
		return "context-file"
	case gptCommandOptionModel:
		return "model"
	case gptCommandOptionTemperature:
		return "temperature"
	}
	return fmt.Sprintf("ApplicationCommandOptionType(%d)", t)
}

// The humanReadableString function is another method on the gptCommandOptionType type that returns a human-readable string representation of the option. 
// The function uses a switch statement to return the appropriate human-readable string representation based on the value of the option.
func (t gptCommandOptionType) humanReadableString() string {
	switch t {
	case gptCommandOptionPrompt:
		return "Prompt"
	case gptCommandOptionContext:
		return "Context"
	case gptCommandOptionContextFile:
		return "Context file"
	case gptCommandOptionModel:
		return "Model"
	case gptCommandOptionTemperature:
		return "Temperature"
	}
	return fmt.Sprintf("ApplicationCommandOptionType(%d)", t)
}
