package main

import (
	"dgo-wrapper/logger"
	"dgo-wrapper/src"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"time"
)

type Conf struct {
	Token string
	Prefix string
}
var Config *Conf

func main() {
	f, _ := ioutil.ReadFile("./bot/config.yml")
	_ = yaml.Unmarshal(f, &Config)

	s, err := dgowrapper.Initialize(&dgowrapper.Options{
		Prefixes: []string{Config.Prefix},
		Token:    Config.Token,
		Intent:   discordgo.IntentsAllWithoutPrivileged,
		Handlers: []interface{}{CreateMessage, Ready},
	})
	if err != nil {
		panic(err)
	}

	err = dgowrapper.NewCommands([]*dgowrapper.Command{
		{
			Name: "ping",
			Args: []*dgowrapper.Argument{
				{
					Name:          "add",
					RequiresValue: true,
				},
				{
					Name: "rm",
					RequiresValue: true,
				},
			},
			Aliases:      []string{"pong"},
			Examples:     []string{"-ping"},
			Descriptions: []string{"standard ping-pong command"},
			Handler: func(ctx *dgowrapper.Context) {
				args, err := ctx.ParseArgs()

				if err == dgowrapper.ErrNoArgsFound {
					ctx.Send("No ags")
					return
				}

				if !args["add"].HasValue {
					ctx.Send("No")
					return
				}

				if err != nil {
					panic(err)
				}
				for k, v := range args {
					fmt.Println(k, v)
				}
				ctx.Send("Pong")
			},
		},
		{
			Name:         "ping2",
			Aliases:      []string{"pong2"},
			Examples:     []string{"-ping"},
			Descriptions: []string{"standard ping-pong command"},
			Handler: func(ctx *dgowrapper.Context) {
				ctx.Send("Pong2")
				collector := dgowrapper.MessageCollector{
					Filter:   []dgowrapper.Filter{dgowrapper.IsNumber},
					Timeout:  5 * time.Second,
					EndAfter: 1,
				}

				err := collector.New(ctx)
				if err != nil {
					ctx.Send(err.Error())
					return
				}

				ctx.Sendf("You said %s", collector.MessagesCollected[0].Message.Content)
			},
		},
	})
	if err != nil {
		panic(err)
	}

	dgowrapper.LogLoadedCommands(dgowrapper.Commands)

	err = dgowrapper.WaitForTerm(s)
	if err != nil {
		panic(err)
	}
}

func CreateMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	msg := strings.Split(strings.ToLower(m.Message.Content)[len(Config.Prefix):], " ")
	if len(msg) <= 0 {
		return
	}
	// Make sure message starts with prefix
	if !strings.HasPrefix(strings.ToLower(m.Message.Content), strings.ToLower(dgowrapper.Prefix)) {
		return
	}

	// Get the command
	cmd, err := dgowrapper.FindCommandOrAlias(dgowrapper.Prefix, m.Content)
	// Can't find command
	if err == dgowrapper.ErrNoCommandOrAliasFound  {
		return
	}

	// Make sure it's in guild
	channel, _ := s.Channel(m.ChannelID)
	if channel.Type == discordgo.ChannelTypeDM {
		return
	}

	ctx := &dgowrapper.Context{
		Session: s,
		Message: m,
		Command: cmd,
		Prefix:  dgowrapper.Prefix,
	}
	cmd.Handler(ctx)
}

func Ready(s *discordgo.Session, e *discordgo.Ready) {
	// Add mention prefix
	dgowrapper.Bot.Prefixes = append(dgowrapper.Bot.Prefixes, []string{
		fmt.Sprintf("<@%s> ", e.User.ID),
		fmt.Sprintf("<@!%s> ", e.User.ID),
	}...)

	logger.Logf("BOT", "%s#%s is ready", logger.Yellow.Sprint(e.User.Username), logger.Yellow.Sprint(e.User.Discriminator))
	err := s.UpdateGameStatus(0, "with yo momma")
	logger.Die(err)
}
