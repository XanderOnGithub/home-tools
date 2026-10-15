package discord

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

// Posts from the dashboard (decision #59): a message, a poll, or both,
// written on the Send page and posted by the bot in a channel picked
// there. Nothing is stored: the post lives in Discord. Separate from the
// poll feature (#44), which can stay off.
//
// Only the settings page sends these. It's on the LAN with no login
// (#10), like every other setting; there's no slash command, so nobody in
// Discord can make the bot speak.

// Post is one thing to send.
type Post struct {
	ChannelID string `json:"channel_id"`
	// Message is the text, or "" for a poll on its own.
	Message string `json:"message,omitempty"`
	// Everyone pings @everyone (the message gets "@everyone" in front).
	Everyone bool `json:"everyone,omitempty"`
	// Poll is optional: question, answers, single or multiple choice.
	Poll *PostPoll `json:"poll,omitempty"`
}

// PostPoll is a poll on a Post: a question like the scheduled ones, open
// for DurationHours.
type PostPoll struct {
	PollQuestion
	DurationHours int `json:"duration_hours"` // 1–168, like the scheduled poll
}

const maxMessageLen = 2000 // Discord's message limit

// Validate checks p against Discord's limits: a channel, and a message,
// a poll, or both.
func (p Post) Validate() error {
	switch {
	case !snowflake.MatchString(p.ChannelID):
		return fmt.Errorf("%w: pick a channel", ErrInvalid)
	case strings.TrimSpace(p.Message) == "" && p.Poll == nil:
		return fmt.Errorf("%w: write a message or a poll", ErrInvalid)
	case utf8.RuneCountInString(p.text()) > maxMessageLen:
		return fmt.Errorf("%w: a message can be at most %d characters", ErrInvalid, maxMessageLen)
	}
	if p.Poll != nil {
		if err := p.Poll.validate(); err != nil {
			return err
		}
		if d := p.Poll.DurationHours; d < 1 || d > 168 {
			return fmt.Errorf("%w poll: open for 1–168 hours, got %d", ErrInvalid, d)
		}
	}
	return nil
}

// text is the message as Discord gets it.
func (p Post) text() string {
	msg := strings.TrimSpace(p.Message)
	if p.Everyone {
		msg = strings.TrimSpace("@everyone " + msg)
	}
	return msg
}

// message builds what's sent. Mentions are off except @everyone when
// asked for: a pasted "@here" or role mention in the text never pings.
func (p Post) message() *discordgo.MessageSend {
	m := &discordgo.MessageSend{Content: p.text(), AllowedMentions: noPings}
	if p.Everyone {
		m.AllowedMentions = &discordgo.MessageAllowedMentions{Parse: []discordgo.AllowedMentionType{discordgo.AllowedMentionTypeEveryone}}
	}
	if q := p.Poll; q != nil {
		poll := &discordgo.Poll{Question: discordgo.PollMedia{Text: q.Question}, AllowMultiselect: q.Multi, Duration: q.DurationHours}
		for _, a := range q.Answers {
			poll.Answers = append(poll.Answers, discordgo.PollAnswer{Media: &discordgo.PollMedia{Text: a}})
		}
		m.Poll = poll
	}
	return m
}

// errNoChannel: the channel isn't one the bot can see (wrong server,
// deleted, or not a text channel).
var errNoChannel = errors.New("the bot can't see that channel")

// Send posts p and answers with a link to the message.
func (b *Bot) Send(p Post) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	b.mu.RLock()
	connected := b.conn.connected
	b.mu.RUnlock()
	if !connected || b.sess == nil {
		return "", errOffline
	}
	ch, err := b.sess.State.Channel(p.ChannelID)
	if err != nil || (ch.Type != discordgo.ChannelTypeGuildText && ch.Type != discordgo.ChannelTypeGuildNews) {
		return "", errNoChannel
	}
	msg, err := b.sess.ChannelMessageSendComplex(p.ChannelID, p.message())
	if err != nil {
		return "", fmt.Errorf("send: %w", err)
	}
	kind := "message"
	if p.Poll != nil {
		kind = "poll"
	}
	b.log.Info("discord post sent", "channel", ch.Name, "kind", kind, "everyone", p.Everyone)
	return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", ch.GuildID, ch.ID, msg.ID), nil
}
