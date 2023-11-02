package dalle

import (
	"fmt"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/constants"
	discord "github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

const (
	imageDefaultSize = openai.CreateImageSize256x256

	imagePriceSize256x256   = 0.016
	imagePriceSize512x512   = 0.018
	imagePriceSize1024x1024 = 0.02
)

// The code block contains two functions: priceForResponse and imageCreationUsageEmbedFooter.


// The priceForResponse function takes an integer n and a string size as input and returns a float64 value representing the price
// for generating the specified number of images at the specified size. The function uses a switch statement to determine the price based
// on the specified size. If the specified size is not recognized, the function returns 0.
func priceForResponse(n int, size string) float64 {
	switch size {
	case openai.CreateImageSize256x256:
		return float64(n) * imagePriceSize256x256
	case openai.CreateImageSize512x512:
		return float64(n) * imagePriceSize512x512
	case openai.CreateImageSize1024x1024:
		return float64(n) * imagePriceSize1024x1024
	}

	return 0
}


// The imageCreationUsageEmbedFooter function takes a string size and an integer number as input and returns a discord.MessageEmbedFooter object. 
// The discord.MessageEmbedFooter object represents the footer of a message embed and contains information about the image creation usage. 
// The function first creates a string extraInfo containing the size and number of images. 
// The function then calls the priceForResponse function to determine the cost of generating the specified number of images at the specified size. 
// If the cost is greater than 0, the function appends the cost to the extraInfo string. 
// The function then returns a discord.MessageEmbedFooter object containing the extraInfo string and a URL to an icon.
func imageCreationUsageEmbedFooter(size string, number int) *discord.MessageEmbedFooter {
	extraInfo := fmt.Sprintf("Size: %s, Images: %d", size, number)
	price := priceForResponse(number, size)
	if price > 0 {
		extraInfo += fmt.Sprintf("\nGeneration Cost: $%g", price)
	}
	return &discord.MessageEmbedFooter{
		Text:    extraInfo,
		IconURL: constants.OpenAIBlackIconURL,
	}
}
