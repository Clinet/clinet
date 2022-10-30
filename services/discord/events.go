package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/discordgo-embed"
)

func discordReady(session *discordgo.Session, event *discordgo.Ready) {
	Log.Trace("--- discordReady(", event, ") ---")
	for Discord == nil {
		//Wait for Discord to finish connecting on our end
		if Discord != nil {
			break
		}
	}
	Discord.User = event.User
	Log.Info("Logged into Discord as ", Discord.User, "!")
}

func discordMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	Log.Trace("--- discordMessageCreate(", event, ") ---")
	message, err := session.ChannelMessage(event.ChannelID, event.ID)
	if err != nil {
		Log.Error(message, err)
	}

	cmdResps, err := convoHandler(session, message)
	if err != nil {
		Log.Error(err, cmdResps)
		return
	}

	for i := 0; i < len(cmdResps); i++ {
		if cmdResps[i] == nil {
			continue
		}

		cmdResps[i].OnReady(func(r *cmds.CmdResp) {
			Log.Trace("Response to message for convo: " + r.String())
			r.Context = event.Message
			r.ChannelID = event.ChannelID
				
			msg, err := Discord.MsgSend(r.Message)
			if err != nil {
				Log.Error(err)
				return
			}
			Log.Trace("Sent message: ", msg)
		})
	}
}

func discordInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	Log.Trace("--- discordInteractionCreate(", event, ") ---", event.ID, event.Type, event.GuildID, event.ChannelID, event.Member, event.User, event.Token, event.Version)

	switch event.Type {
	case discordgo.InteractionApplicationCommand:
		cmd := cmds.GetCmd(event.ApplicationCommandData().Name)
		if cmd == nil {
			Log.Error("Unable to find command " + event.ApplicationCommandData().Name)
			return
		}

		eventOpts := event.ApplicationCommandData().Options
		cmdAlias, cmdResps := cmdHandler(cmd, event.Interaction, eventOpts, false)
		for i := 0; i < len(cmdResps); i++ {
			if cmdResps[i] == nil {
				continue
			}

			cmdResps[i].OnReady(func(r *cmds.CmdResp) {
				Log.Trace("Response to interaction for cmd " + cmdAlias + ": " + r.String())
				r.Context = event.Interaction
				
				msg, err := Discord.MsgSend(r.Message)
				if err != nil {
					Log.Error(err)
					return
				}
				Log.Trace("Sent message: ", msg)
			})
		}
	case discordgo.InteractionMessageComponent:
		msgData := event.MessageComponentData()
		if msgData.ComponentType != discordgo.ButtonComponent {
			return
		}

		switch msgData.CustomID {
		case "example1":
			//do something when this button is pressed
		}

		respEmbed := embed.NewGenericEmbed("Feature Name", "Example response embed from feature")
		featureComponents := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "Example 1",
						Style: discordgo.SuccessButton,
						CustomID: "example1",
					},
				},
			},
		}
		resp := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				CustomID: "featureName",
				Embeds: []*discordgo.MessageEmbed{respEmbed},
				Components: featureComponents,
				Flags: discordgo.MessageFlagsEphemeral,
			},
		}

		err := session.InteractionRespond(event.Interaction, resp)
		if err != nil {
			Log.Error(err)
		} else {
			Log.Debug("Responded to button ", msgData.CustomID, " with response: ", fmt.Sprintf("%v", respEmbed))
		}
	}
}
