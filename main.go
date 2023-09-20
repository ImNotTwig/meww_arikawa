package main

import (
	"context"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/api/cmdroute"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/goccy/go-json"
)

var commands = []api.CreateCommandData{
	{
		Name:        "ping",
		Description: "Ping!",
	},
	{
		Name:        "set_join_log_channel",
		Description: "Used to set the Join/Leave Log Channel.", Options: []discord.CommandOption{
			&discord.ChannelOption{
				OptionName:  "log_channel",
				Description: "The channel you want to set for logging member joining/leaving",
				Required:    true,
			},
		}},
}

// This type is for the JoinLogChannelList json data,
// it's used for keeping track of what all the guilds' log servers are
type JoinLogChannelList struct {
	JoinLogChannelList map[string]string `json:JoinLogChannelList`
}

func main() {
	r := cmdroute.NewRouter()
    config := GetConfig()

	// Ping Command
	r.AddFunc("ping", func(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
		return &api.InteractionResponseData{Content: option.NewNullableString("Pong!")}
	})

	// Getting the JoinLogChannelList json file
	var jsonFile *os.File

	// checking if the JoinLogChannelList.json file exists or not, and if it doesnt, we create it
	if _, err := os.Stat("./data/JoinLogChannelList.json"); err == nil {
		// file exists
		jsonFile, _ = os.Open("./data/JoinLogChannelList.json")
	} else {
		os.Create("./data/JoinLogChannelList.json")
		jsonFile, _ = os.Open("./data/JoinLogChannelList.json")
	}

	// unmarshaling the file into a struct so we can actually read it
	var logList JoinLogChannelList
	byteJson, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteJson, &logList)
    jsonFile.Close()

	r.AddFunc("set_join_log_channel", func(ctx context.Context, data cmdroute.CommandData) *api.InteractionResponseData {
		logList.JoinLogChannelList[data.Event.GuildID.String()] = strings.Trim(data.CommandInteractionOption.Options[0].Value.String(), "\"")
		marshaledList, err := json.Marshal(logList)
		if err != nil {
			return &api.InteractionResponseData{Content: option.NewNullableString("ERROR: Could not write JoinLogChannelList to file, please message twigthecat, or more preferably, make an issue at https://github.com/ImNotTwig/meww_arikawa")}
        }
        jsonFile, _ = os.OpenFile("./data/JoinLogChannelList.json", os.O_WRONLY, os.ModeAppend)
        n, err := jsonFile.Write(marshaledList)
        jsonFile.Close()
        if err != nil {
            log.Println(err)
            log.Println(n)
        }
		return &api.InteractionResponseData{Content: option.NewNullableString(data.CommandInteractionOption.Options[0].Value.String() + " has been set as the Join Log Channel.")}
	})

	s := state.New("Bot " + config.Tokens.Discord)
	s.AddIntents(gateway.IntentGuilds)
	s.AddInteractionHandler(r)

	// Fires when a user joins a server
	s.AddHandler(func(g *gateway.GuildMemberRemoveEvent) {
		// TODO: ask server to set a log channel/ specifically a leaving and joining log channel
	})

	// getting the data for "me", aka the bot
	me, err := s.Session.Me()
	if err != nil {
		log.Fatalln("Can't get self Session data:", err)
	}

	// Print when we are logged in
	s.AddHandler(func(g *gateway.ReadyEvent) {
		log.Println("Joined as: ", me.Username)
	})

	// The handler for when we join Guilds
	// it checks if we actually just joined, or if the bot is just coming online
	// if we are actually joining, it sends a joining message to the servers system logging channel
	s.AddHandler(func(g *gateway.GuildCreateEvent) {
		memberMe, err := s.Member(g.Guild.ID, me.ID)
		if err != nil {
			log.Fatalln("error finding self in guild "+g.Name, err)
		}
		// check if the GuildCreate time is at least 5 seconds older than the JoinedTime, so we can differentiate the bot starting up,
		// and when we actually join a server
		if !time.Now().After(memberMe.Joined.Time().Add(time.Second * 5)) {
			s.SendMessage(g.Guild.SystemChannelID, `You've invited meww to your server, Please set a leave/join log channel, and a other logs channel. 
use the commands ~SetJoinLogChannel and ~SetOtherLogChannel . Slash Commands are also available.`)
		}
	})

	if err := cmdroute.OverwriteCommands(s, commands); err != nil {
		log.Fatalln("cannot update commands:", err)
	}

	if err := s.Connect(context.TODO()); err != nil {
		log.Println("cannot connect:", err)
	}
}
