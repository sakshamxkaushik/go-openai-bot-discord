package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	discord "github.com/bwmarrin/discordgo"
)

//defining the struct for a bot, it'll have a session and a router

type Bot struct {
	*discord.Session

	Router *Router
}

// the new bot function that we use in main.go, takes in the API token
// and returns the bot which is based on the bot struct we made above
func NewBot(token string) (*Bot, error) {
	//calls the New function from the discord package and returns session
	session, err := discord.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	//this function is supposed to return a bot based on the struct defined above
	return &Bot{
		Session: session,
		//NewRouter function is in the router.go file in this package
		Router: NewRouter(nil),
	}, nil
}

// the run function is called in the main.go file, takes in guildID and the remove commands
// b is of type Bot struct which means this is a struct method
func (b *Bot) Run(guildID string, removeCommands bool) {
	// IntentMessageContent is required for us to have a conversation in threads without typing any commands
	//this will enable conversations for us in threads without typing any commands
	b.Identify.Intents = discord.MakeIntent(discord.IntentsAllWithoutPrivileged | discord.IntentMessageContent)

	// Add handler takes in a function and that particular function is supposed to
	//take in session and ready handler that logs in as the bot user
	b.AddHandler(func(s *discord.Session, r *discord.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	//handle interaction and handle message are two functions in the router.go file
	//when we say router.HandleInteraction, we're saying that router is actually a struct and
	//Handle interaction is a struct method available to us
	b.AddHandler(b.Router.HandleInteraction)
	b.AddHandler(b.Router.HandleMessage)

	// Run the bot
	err := b.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	// Sync command opens the bot's session and syncs the command with the given GuildID
	// we will write the logic for sync in router file
	//we need to call the sync struct method for router
	err = b.Router.Sync(b.Session, guildID)
	if err != nil {
		panic(err)
	}

	//closes the bot at the end, when this particular function exits.
	//this means we will close the bot session after our interaction and
	//the next time, a different bot session will start
	defer b.Close()
	//now we want to handle graceful shutdown, we will create a channel using the os package
	stop := make(chan os.Signal, 1)
	//signal.Notify is used to send this to the stop channel, we're essentially listening
	//for ctrl+c or something similar, we want to ensure that with ghraceful shutdown, our session
	//is ended and also the commands are unregistered
	//the below function tells the program to listen for two speicif signals -
	//os.interrupt that happens when user presses ctrl+c and syscall.sigterm, which is a termination
	//signal that can be sent by external processes or management tools to request for a graceful shutdown
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	//in our case, we have selected true for remove commands, this means when the bot is stopped
	// the commands will be unregistered
	if removeCommands {
		log.Println("Removing commands...")
		//essentially calling the clearCommands function in the router file
		//takes in the particular session received when creating the bot
		//and the particular guild or the server
		b.Router.ClearCommands(b.Session, guildID)
	}

	log.Println("Gracefully shutting down.")
}
