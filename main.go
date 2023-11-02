package main

import (
	"log"
	"os"

	"github.com/akhilsharma90/go-openai-bot-discord/pkg/bot"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/commands"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/commands/gpt"
	"github.com/akhilsharma90/go-openai-bot-discord/pkg/constants"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/yaml.v2"
)

// creating a struct to work with the discord and open ai keys
type Config struct {
	//there are 2 sub structs in this, one is for discord
	Discord struct {
		//discord bot requires a token, which is present in the yaml file with the name "token"
		//when we write yaml:token, we're specifying the name of the field in the yaml field
		//that field will correspond to this particular field, which is Token and will be binded in runtime
		Token string `yaml:"token"`
		//guild is basically the particular server you want the bot to be present in
		//in our case, we invited the bot to a particular workspace, but you can also
		//specify it from the code itself so we will keep this field so that if we want
		//to put our app into production, we have this capability
		Guild string `yaml:"guild"`
		//when our project shuts down, we either want to remove all the commands that we set
		//for the bot or we want to keep them, this is a boolean value, either true or false
		//and we can set it in our credentials file
		RemoveCommands bool `yaml:"removeCommands"`
	} `yaml:"discord"`
	//this is the other sub struct, for open AI, in the previous one, we mentioned, yaml discord
	//because all of the above values will be under the heading, discord
	OpenAI struct {
		APIKey           string   `yaml:"apiKey"`
		CompletionModels []string `yaml:"completionModels"`
	} `yaml:"openAI"`
	//all of the above values will be under the openAI heading
}

// with this function, you can read config values from the yaml file
// we pass the name of the file in it
func (c *Config) ReadFromFile(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	//unmarshalling function enables us to convert values from yaml to a higher level
	//object such as a golang struct, we need the struct to be able to work in golangf
	//since yaml and json aren't supported by default
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// defining some variables in a group since we need to be able to work with
// the discord bot, open ai client, we will be creating a cache for the messages and need to name it
// also need an ignored channels cache
var (
	discordBot   *bot.Bot
	openaiClient *openai.Client

	gptMessagesCache     *gpt.MessagesCache
	ignoredChannelsCache = make(gpt.IgnoredChannelsCache)
)

func main() {
	// initialize config variable which is of type Config, a struct we have defined above
	config := &Config{}
	//earlier we defined readfromfile as a struct method that's available to belonging to the struct
	//type config and we're now able to access that function with a dot operator placed on config
	//and are able to run that function by passing in the name of the file, we know that that particular
	//function accepts the name of the file as a string
	err := config.ReadFromFile("credentials.yaml")
	//if there's an error reading the credentials file, we will handle that error
	if err != nil {
		log.Fatalf("Error reading credentials.yaml: %v", err)
	}

	// we defined the variable gptmessagescache earlier, we will initiate it with
	//NewMessagesCache function in the gpt package
	gptMessagesCache, err = gpt.NewMessagesCache(constants.DiscordThreadsCacheSize)
	if err != nil {
		log.Fatalf("Error initializing GPTMessagesCache: %v", err)
	}

	// Initialize discord bot by calling the NewBot function from the bot package, that we have created(bot folder)
	//we pass the token from config file, under discord topic
	discordBot, err = bot.NewBot(config.Discord.Token)
	//handle the error if the parameters are invalid
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	//first we will check that in the config file, under the open ai topic, the api key is not empty
	if config.OpenAI.APIKey != "" {
		//if it's not empty, we start a new open ai client by passing the APIKey and assign it
		//to the variable called openaiClient to get the ball rolling
		openaiClient = openai.NewClient(config.OpenAI.APIKey) // initialize OpenAI client first
		//we want to register the commands on the discord bot
		//so we call the register function and assign the parameters to the differnet fields
		//we have 3 commands, so the first thing we register is the chat command, then we register
		//the image command and then the info command
		//commands package is something that we have created (commands folder)
		discordBot.Router.Register(commands.ChatCommand(&commands.ChatCommandParams{
			OpenAIClient:           openaiClient,
			OpenAICompletionModels: config.OpenAI.CompletionModels,
			GPTMessagesCache:       gptMessagesCache,
			IgnoredChannelsCache:   &ignoredChannelsCache,
		}))

		discordBot.Router.Register(commands.ImageCommand(openaiClient))
	}
	discordBot.Router.Register(commands.InfoCommand())

	// Run the bot by passing in values from the config file for guild and remove commands
	//in our case guild is empty but you can set a specific value if required
	discordBot.Run(config.Discord.Guild, config.Discord.RemoveCommands)
}
