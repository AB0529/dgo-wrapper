package dgowrapper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/AB0529/dgo-wrapper/logger"
	"github.com/bwmarrin/discordgo"
)

// TODO:
// - Command flag system ✅
// - Message collector
// - Reaction collector

// Options the options passed into the bot
type Options struct {
	// Token the bot token
	Token string
	// Prefixes the prefixes the bot accepts
	Prefixes []string
	// Intents the intents for the bot
	Intent discordgo.Intent
	// Handlers event handler functions
	Handlers []interface{}
	// ID the bot id
	ID string
}

// Context context about the bot passed to each command
type Context struct {
	// Session the Discord session
	Session *discordgo.Session
	// Message the message event that triggered the command
	Message *discordgo.MessageCreate
	// LastMessage the last message sent
	LastMessage chan *discordgo.MessageCreate
	// Command the command which was ran
	Command *Command
	// Prefix current prefix for the bot
	Prefix string
	// Bot the options for the bot
	Bot Options
}

// Argument argument for a command
type Argument struct {
	// Name name of the argument
	Name string
	// Value the value after the argument
	Value string
	// RequiresValue determines if requires a value or not
	RequiresValue bool
	// HasValue determines if value is passed in or not
	HasValue bool
	// Panic throw an error or return empty value
	Panic bool
}

// Command represents a command for the bot
type Command struct {
	// Name the name for the command
	Name string
	// Args any args the command takes
	Args []*Argument
	// Aliases any aliases for the command
	Aliases []string
	// Examples examples on how to use the command
	Examples []string
	// Descriptions descriptions on the command
	Descriptions []string
	// Handler the handler function for the command
	Handler func(*Context)
}

// Filter a filter function used in a message collector
type Filter func(v interface{}) error

// MessageCollector represents the options for the message collection
type MessageCollector struct {
	// MessagesCollected the amount of messages collected
	MessagesCollected []*discordgo.MessageCreate
	// Filter functions to run on each message
	Filter []Filter
	// Timeout the duration of the message collector
	Timeout time.Duration
	// EndAfter end collector after n amount of messages
	EndAfter int
	Done     chan bool
}

var (
	ErrEmptyToken            = errors.New("bot token cannot be an empty value")
	ErrEmptyPrefixes         = errors.New("must have at least one prefix")
	ErrEmptyIntents          = errors.New("must specify intents")
	ErrEmptyHandlers         = errors.New("must specify handlers")
	ErrNoPrefixFound         = errors.New("no prefix found in message content")
	ErrNoCommandOrAliasFound = errors.New("no command or alias found from message content")
	ErrNoArgsToParse         = errors.New("command has no arguments to parse")
	ErrNoArgsFound           = errors.New("no args matching command found")
	ErrNoValueForArg         = errors.New("no value found for required argument")
	ErrFilterFailed          = errors.New("filter did not work on value")
	// Commands holds the commands for the bot
	Commands = make(map[string]*Command)
	// Aliases holds the aliases for each command
	Aliases = make(map[string]*Command)
	// Bot empty options for the bot to start off
	Bot = &Options{}
	// Prefix the selected prefix (default: "!!!")
	Prefix = "!!!"
	// LastMessage last message channel
	LastMessage = make(chan *discordgo.MessageCreate)
	// LastReactionAdd last add reaction chan
	LastReactionAdd = make(chan *discordgo.MessageReaction)
	// LastReactionRemove last reaction removed chan
	LastReactionRemove = make(chan *discordgo.MessageReaction)
)

func checkOptions(opt *Options) error {
	// Check if token is empty
	if opt.Token == "" {
		return ErrEmptyToken
	}
	// Check if prefixes is empty
	if len(opt.Prefixes) <= 0 {
		return ErrEmptyPrefixes
	}
	// Check if intents are empty
	if opt.Intent == discordgo.IntentsNone {
		return ErrEmptyIntents
	}
	// Check for empty handlers
	if len(opt.Handlers) <= 0 {
		return ErrEmptyHandlers
	}

	return nil
}

// Initialize will initialize a new bot
func Initialize(opt *Options) (*discordgo.Session, error) {
	err := checkOptions(opt)
	if err != nil {
		return nil, err
	}
	dg, err := discordgo.New("Bot " + opt.Token)
	if err != nil {
		return nil, err
	}

	// Inject ready handler stuff
	dg.AddHandler(func(s *discordgo.Session, e *discordgo.Ready) {
		// Set bot id
		Bot.ID = e.User.ID
	})
	// Inject message handler stuff
	// @depercated doesn't actually inject
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore bots
		if m.Author.Bot {
			return
		}
		// Ignore no content messages
		if m.Content == "" {
			return
		}
		// Select prefix from list of prefixes
		Prefix, err = SelectPrefix(m.Content)
		if err == ErrNoPrefixFound {
			LastMessage <- m
			return
		}
	})
	// Inject reaction add handler stuff
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
		LastReactionAdd <- m.MessageReaction
	})
	// Inject reaction remove handler stuff
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
		LastReactionRemove <- m.MessageReaction
	})

	for _, handler := range opt.Handlers {
		dg.AddHandler(handler)
	}
	dg.Identify.Intents = opt.Intent

	err = dg.Open()
	if err != nil {
		return nil, err
	}

	Bot = opt

	return dg, err
}

// WaitForTerm waits for term signal and terminates the bot
func WaitForTerm(dg *discordgo.Session) error {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	err := dg.Close()
	return err
}

// NewCommand adds a new command to the bot
func NewCommand(cmd *Command) error {
	Commands[cmd.Name] = cmd

	// For each alias, create a new alias command
	for _, alias := range cmd.Aliases {
		// Make sure there's no duplicate aliases
		if _, ok := Commands[alias]; ok {
			return fmt.Errorf("alias '%s' already exists for another command", alias)
		}
		Aliases[alias] = cmd
	}

	return nil
}

// NewCommands adds multiple commands to the bot
func NewCommands(cmds []*Command) error {
	for _, cmd := range cmds {
		err := NewCommand(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// LogLoadedCommands logs all commands into stdout
func LogLoadedCommands(cmds map[string]*Command) {
	for _, cmd := range cmds {
		logger.Logf("CMD", "%s (%s) loaded", logger.Green.Sprint(cmd.Name), logger.Yellow.Sprint(len(cmd.Aliases)))
	}
}

// New starts a new instance of a message collector
func (collector *MessageCollector) New(ctx *Context)  error {
	// Use timeout instead of channel
	if collector.Timeout >= time.Second * 0 {
		// Create timeout context
		c, cancel := context.WithTimeout(context.Background(), collector.Timeout)
		defer cancel()

		sel:
		select {
			case msg := <-LastMessage:
				// Cancel collector
				if msg.Timestamp.After(ctx.Message.Timestamp) && msg.Author.ID == ctx.Message.Author.ID && strings.ToLower(msg.Content) == "c" {
					return errors.New("collector canceled")
				}

				if len(collector.Filter) <= 0 {
					if msg.Timestamp.After(ctx.Message.Timestamp) {
						if collector.EndAfter == 1 {
							collector.MessagesCollected = append(collector.MessagesCollected, msg)
							return nil
						}
					} else {
						goto sel
					}
				}

				for _, f := range collector.Filter {
					// Run filter on message
					err := f(msg.Content)

					if msg.Timestamp.After(ctx.Message.Timestamp) {
						if err == nil {
							if collector.EndAfter == 1 {
								collector.MessagesCollected = append(collector.MessagesCollected, msg)
								return nil
							}
							collector.MessagesCollected = append(collector.MessagesCollected, msg)
							goto sel
						}
						return err
					} else {
						goto sel
					}
				}
			case <-c.Done():
				if collector.EndAfter == 0 {
					return nil
				}
				return c.Err()
		}
	}

	return nil
}
