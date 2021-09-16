package dgowrapper

import (
	"dgo-wrapper/logger"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strings"
)

// Send sends a simple message
func (ctx *Context) Send(content string) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, content)
	logger.Warn(err)
	return m
}

// Sendf sends a formatted message
func (ctx *Context) Sendf(content string, a ...interface{}) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageSend(ctx.Message.ChannelID, fmt.Sprintf(content, a...))
	logger.Warn(err)
	return m
}

// Embed creates a simple embed with a description and a random color
func (ctx *Context) Embed(content string) *discordgo.Message {
	msg, err := ctx.Session.ChannelMessageSendComplex(ctx.Message.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Color:       rand.Intn(10000000),
			Description: content,
		},
	})
	logger.Warn(err)
	return msg
}

// Err sends a error message to the same channel as the command
func (ctx *Context) Err(content interface{}) {
	// Handle strings being passed by creating an error type
	if content != nil {
		if e, ok := content.(string); ok {
			content = errors.New(e)
		}
		ctx.Embedf(":x: | Uh oh, something **went wrong**!\n```css\n%s\n```", content)
	}
}

// Edit edits a message with a new content
func (ctx *Context) Edit(m *discordgo.Message, content string) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageEdit(m.ChannelID, m.ID, content)
	logger.Warn(err)
	return m
}

// Editf edits a message with a new formatted content
func (ctx *Context) Editf(m *discordgo.Message, content string, a ...interface{}) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageEdit(m.ChannelID, m.ID, fmt.Sprintf(content, a...))
	logger.Warn(err)
	return m
}

// Embedf creates a simple embed with a formatted description and a random color
func (ctx *Context) Embedf(content string, a ...interface{}) *discordgo.Message {
	msg, err := ctx.Session.ChannelMessageSendComplex(ctx.Message.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Color:       rand.Intn(10000000),
			Description: fmt.Sprintf(content, a...),
		},
	})
	logger.Warn(err)
	return msg
}

// EditEmbed edits a embed with a new content
func (ctx *Context) EditEmbed(m *discordgo.Message, content string) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Embed: &discordgo.MessageEmbed{
			Color:       rand.Intn(10000000),
			Description: content,
		},
		ID:      m.ID,
		Channel: ctx.Message.ChannelID,
	})
	logger.Warn(err)
	return m
}

// EditEmbedf edits a embed with a new formatted content
func (ctx *Context) EditEmbedf(m *discordgo.Message, content string, a ...interface{}) *discordgo.Message {
	m, err := ctx.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Embed: &discordgo.MessageEmbed{
			Color:       rand.Intn(10000000),
			Description: fmt.Sprintf(content, a...),
		},
		ID:      m.ID,
		Channel: ctx.Message.ChannelID,
	})
	logger.Warn(err)
	return m
}

// SendCommandHelp properly formats and shows the help page of a command
func (ctx *Context) SendCommandHelp() {
	_, err := ctx.Session.ChannelMessageSendComplex(ctx.Message.ChannelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Description: fmt.Sprintf("`%s%s` Command Help", ctx.Bot.Prefixes[0], ctx.Command.Name),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "📜 | Description",
					Value: fmt.Sprintf("```css\n%s\n```", strings.Join(ctx.Command.Descriptions, "\n")),
				},
				{
					Name:  "🤖 | Example",
					Value: fmt.Sprintf("```css\n%s\n```", strings.Join(ctx.Command.Examples, "\n")),
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text:    fmt.Sprintf("Aliases: %s", strings.Join(ctx.Command.Aliases, " | ")),
				IconURL: ctx.Message.Message.Author.AvatarURL("512x512"),
			},
			Color: rand.Intn(10000000),
		},
	})
	logger.Warn(err)
}

// CmdRan the name of the command or alias that was ran
func (ctx *Context) CmdRan() string {
	msg := strings.Split(
		strings.Replace(strings.ToLower(ctx.Message.Content), strings.ToLower(ctx.Prefix), "", 1), " ")
	return msg[0]
}

// ParseArgs parses arguments for the command
func (ctx *Context) ParseArgs() (map[string]*Argument, error) {
	if len(ctx.Command.Args) <= 0 {
		return nil, ErrNoArgsToParse
	}
	foundArgs := make(map[string]*Argument)

	// Remove prefix + command
	cArgs := strings.Replace(strings.ToLower(ctx.Message.Content), ctx.Prefix+ctx.CmdRan(), "", 1)
	args := strings.Split(cArgs, " ")

	for i, arg := range args {
		// Make sure
		for _, flag := range ctx.Command.Args {
			if strings.ToLower(arg) == flag.Name {
				// Add flag without required value
				if !flag.RequiresValue {
					foundArgs[flag.Name] = &Argument{
						Name:     flag.Name,
						HasValue: true,
					}
				}
				// Add flag with value
				if flag.RequiresValue {
					// Pass next element ass value for flag
					if i+1 >= len(args) {
						if flag.Panic {
							return nil, ErrNoValueForArg
						}
						foundArgs[flag.Name] = &Argument{
							Name:          flag.Name,
							Value:         "",
							RequiresValue: flag.RequiresValue,
							HasValue:      false,
						}
						continue
					}
					foundArgs[flag.Name] = &Argument{
						Name:          flag.Name,
						Value:         args[i+1],
						RequiresValue: flag.RequiresValue,
						HasValue:      true,
					}
				}
			}
		}
	}

	if len(foundArgs) <= 0 {
		return nil, ErrNoArgsFound
	}

	return foundArgs, nil
}
