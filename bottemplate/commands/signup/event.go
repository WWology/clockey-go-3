package signup

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
)

const (
	dotaWatchPartyChannel = "738009797932351519"
	csWatchPartyChannel   = "746618267434614804"
	rlWatchPartyChannel   = "1194677990290894989"
	ogStageChannel        = "1186593338300842025"
)

func getHours(seriesLength string) string {
	switch strings.ToLower(seriesLength) {
	case "bo1":
		return "2"
	case "bo2":
		return "3"
	case "bo3":
		return "4"
	case "bo5":
		return "6"
	default:
		return ""
	}
}

var (
	gameModal = discord.ModalCreate{
		CustomID: "game_modal",
		Title:    "Game information",
		Components: []discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.TextInputComponent{
					CustomID:    "name",
					Style:       discord.TextInputStyleShort,
					Label:       "What's the name of the game?",
					Placeholder: "OG vs <opp team name>",
					Required:    true,
				},
			},
			discord.ActionRowComponent{
				discord.TextInputComponent{
					CustomID:    "time",
					Style:       discord.TextInputStyleShort,
					Label:       "What's the scheduled start time for the game?",
					Placeholder: "Insert Unix time here",
					Required:    true,
				},
			},
			discord.ActionRowComponent{
				discord.TextInputComponent{
					CustomID:    "series_length",
					Style:       discord.TextInputStyleShort,
					Label:       "What's the series format of the game?",
					Placeholder: "Bo1 / Bo2 / Bo3 / Bo5 / Bo7",
					Required:    true,
				},
			},
		},
	}
	eventModal = discord.ModalCreate{
		CustomID: "event_modal",
		Title:    "Event information",
		Components: []discord.ContainerComponent{
			discord.ActionRowComponent{
				discord.TextInputComponent{
					CustomID: "name",
					Style:    discord.TextInputStyleShort,
					Label:    "What's the name of the event?",
					Required: true,
				},
			},
			discord.ActionRowComponent{
				discord.TextInputComponent{
					CustomID:    "time",
					Style:       discord.TextInputStyleShort,
					Label:       "What's the scheduled start time for the event?",
					Placeholder: "Insert Unix time here",
					Required:    true,
				},
			},
			discord.ActionRowComponent{
				discord.TextInputComponent{
					CustomID: "hours",
					Style:    discord.TextInputStyleShort,
					Label:    "How many hours is this event?",
					Required: true,
				},
			},
		},
	}
)
var EventCommand = discord.SlashCommandCreate{
	Name:        "event",
	Description: "Create a new event for Gardeners to sign up",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name:        "type",
			Description: "Type of the event",
			Required:    true,
			Choices: []discord.ApplicationCommandOptionChoiceString{
				{
					Name:  "Dota",
					Value: "Dota",
				},
				{
					Name:  "CS",
					Value: "CS",
				},
				{
					Name:  "RL",
					Value: "RL",
				},
				{
					Name:  "Other",
					Value: "Other",
				},
			},
		},
		discord.ApplicationCommandOptionBool{
			Name:        "ping",
			Description: "Should this message ping Gardeners or not",
			Required:    false,
		},
	},
}

func eventModalHandler(
	data discord.SlashCommandInteractionData,
	cmdEvent *handler.CommandEvent,
) func(modalEvent *events.ModalSubmitInteractionCreate) {
	return func(modalEvent *events.ModalSubmitInteractionCreate) {
		ping, ok := data.OptBool("ping")
		if !ok {
			ping = true
		}

		modalData := modalEvent.Data
		var name, unixTimeString, hours, channelID string
		scheduledType := discord.ScheduledEventEntityTypeVoice

		switch data.String("type") {
		case "Dota":
			name = "Dota - " + modalData.Text("name")
			hours = getHours(modalData.Text("series_length"))
			channelID = dotaWatchPartyChannel
		case "CS":
			name = "CS - " + modalData.Text("name")
			hours = getHours(modalData.Text("series_length"))
			channelID = csWatchPartyChannel
		case "RL":
			name = "Rocket League - " + modalData.Text("name")
			hours = "1"
			channelID = rlWatchPartyChannel
		case "Other":
			name = "Other - " + modalData.Text("name")
			hours = modalData.Text("hours")
			channelID = ogStageChannel
			scheduledType = discord.ScheduledEventEntityTypeStageInstance
		}

		unixTimeString = modalData.Text("time")
		unixTimeValue, err := strconv.ParseInt(unixTimeString, 0, 64)
		if err != nil {
			modalEvent.CreateMessage(discord.MessageCreate{
				Content: "Please insert a valid Unix timestamp",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		var replyText string
		if data.String("type") == "Other" {
			replyText = fmt.Sprint("Hey <@&720253636797530203>\n\nI need 1 gardener to work the ", name, " at <t:", unixTimeString, ":F>") +
				"\n\nPlease react below with a <:OGpeepoYes:730890894814740541> to sign up!" +
				fmt.Sprint("\n\nAs this is a ", modalData.Text("series_length"), ", you will be able to add ", hours, " hours of work to invoice for the month")
		} else {
			replyText = fmt.Sprint("Hey <@&720253636797530203>\n\nI need 1 gardener to work the ", name, " at <t:", unixTimeString, ":F>") +
				"\n\nPlease react below with a <:OGpeepoYes:730890894814740541> to sign up!" +
				fmt.Sprint("\n\nYou will be able to add ", hours, " hours of work to your invoice for the month")
		}

		var allowedMentions discord.AllowedMentions
		if ping {
			allowedMentions = discord.AllowedMentions{
				Parse: []discord.AllowedMentionType{
					discord.AllowedMentionTypeRoles,
					discord.AllowedMentionTypeUsers,
				},
			}
		} else {
			allowedMentions = discord.AllowedMentions{
				Parse: []discord.AllowedMentionType{},
			}
		}

		err = modalEvent.CreateMessage(discord.MessageCreate{
			Content:         replyText,
			AllowedMentions: &allowedMentions,
		})
		if err != nil {
			slog.Error("discord response error")
		}

		msg, err := modalEvent.Client().Rest().GetInteractionResponse(
			modalEvent.ApplicationID(),
			modalEvent.Token(),
		)
		if err != nil {
			slog.Error("error fetching message")
		}

		startTime := time.Unix(unixTimeValue, 0)
		channel, err := snowflake.Parse(channelID)
		if err != nil {
			slog.Error("snowflake parsing error")
		}
		modalEvent.Client().Rest().AddReaction(msg.ChannelID, msg.ID, "OGpeepoYes")
		_, err = modalEvent.Client().Rest().CreateGuildScheduledEvent(*cmdEvent.GuildID(),
			discord.GuildScheduledEventCreate{
				Name:               name,
				ScheduledStartTime: startTime,
				ChannelID:          channel,
				PrivacyLevel:       discord.ScheduledEventPrivacyLevelGuildOnly,
				EntityType:         scheduledType,
			})
		if err != nil {
			slog.Error("create event error")
		}
	}
}

func EventCommandHandler(
	data discord.SlashCommandInteractionData,
	cmdEvent *handler.CommandEvent,
) error {
	var err error
	if data.String("type") != "Other" {
		err = cmdEvent.Modal(gameModal)
	} else {
		err = cmdEvent.Modal(eventModal)
	}
	if err != nil {
		return fmt.Errorf("modal response error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	go func() {
		defer cancel()
		bot.WaitForEvent(cmdEvent.Client(), ctx,
			func(modalEvent *events.ModalSubmitInteractionCreate) bool {
				return modalEvent.Data.CustomID == "game_modal" || modalEvent.Data.CustomID == "event_modal"
			},
			eventModalHandler(data, cmdEvent),
			func() {
				_, err := cmdEvent.Client().Rest().CreateMessage(
					cmdEvent.Channel().ID(),
					discord.MessageCreate{
						Content: "Modal timed out",
					})
				if err != nil {
					slog.Error("discord response error", slog.Any("err", err))
				}
			},
		)
	}()

	return nil
}
