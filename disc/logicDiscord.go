package disc

import (
	"bot/config"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/andersfylling/disgord/std"
	"github.com/sirupsen/logrus"
)

// Global Variables to ease working with client/sesion etc
var ctx = context.Background()
var client *disgord.Client
var session disgord.Session
var conf config.ConfJSONStruct

// BotRun the bot and handle events
func BotRun(cf config.ConfJSONStruct) {
	conf = cf
	var log = &logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.ErrorLevel,
	}

	client = disgord.New(disgord.Config{
		BotToken: cf.BotToken,
		Logger:   log,
	})

	defer client.StayConnectedUntilInterrupted(ctx)

	filter, _ := std.NewMsgFilter(ctx, client)
	filter.SetPrefix(cf.Prefix)

	// create a handler and bind it to new message events
	go client.On(disgord.EvtMessageCreate,
		// middleware
		filter.NotByBot,    // ignore bot messages
		filter.HasPrefix,   // read original
		std.CopyMsgEvt,     // read & copy original
		filter.StripPrefix, // write copy

		// handler
		reply, // call reply func
		// specific
	) // handles copy

	fmt.Println("The bot is currently running")
}

func reply(s disgord.Session, data *disgord.MessageCreate) {

	// Parses the message into command / args
	command := strings.Fields(data.Message.Content)[0]
	args := strings.Fields(data.Message.Content)[1:]

	switch {
	case command == "search" || command == "s":
		search(data, args)
	case command == "favourite" || command == "f":
		favourite(data, args)
	case command == "profile" || command == "p":
		profile(data)
	case command == "help" || command == "h":
		help(data)
	case command == "roll" || command == "r":
		roll(data)
	case command == "list" || command == "l":
		list(data, args)
	case command == "invite":
		invite(data)
	default:
		unknown(data)
	}
}