package commands

import (
	"clockey/bottemplate/commands/signup"

	"github.com/disgoorg/disgo/discord"
)

var Commands = []discord.ApplicationCommandCreate{
	// General Purpose Commands
	version,

	// Signup commands
	signup.EventCommand,
}
