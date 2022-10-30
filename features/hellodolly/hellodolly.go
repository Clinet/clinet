package hellodolly

import (
	"math/rand"
	"strings"

	"github.com/Clinet/clinet/cmds"
)

var Cmds []*cmds.Cmd

func init() {
	Cmds = []*cmds.Cmd{
		cmds.NewCmd("hellodolly", "Responds with a random lyric from Louis Armstrong's Hello, Dolly", handleHelloDolly),
	}
}

var lyrics = `Hello, Dolly
Well, hello, Dolly
It's so nice to have you back where you belong
You're lookin' swell, Dolly
I can tell, Dolly
You're still glowin', you're still crowin'
You're still goin' strong
We feel the room swayin'
While the band's playin'
One of your old favourite songs from way back when
So, take her wrap, fellas
Find her an empty lap, fellas
Dolly'll never go away again
Hello, Dolly
Well, hello, Dolly
It's so nice to have you back where you belong
You're lookin' swell, Dolly
I can tell, Dolly
You're still glowin', you're still crowin'
You're still goin' strong
We feel the room swayin'
While the band's playin'
One of your old favourite songs from way back when
Golly, gee, fellas
Find her a vacant knee, fellas
Dolly'll never go away
Dolly'll never go away
Dolly'll never go away again`

func handleHelloDolly(ctx *cmds.CmdCtx) *cmds.CmdResp {
	//Split the lyrics by line into a slice
	lines := strings.Split(lyrics, "\n")

	//Choose a random line
	line := rand.Intn(len(lines))

	//Return the chosen line
	return cmds.NewCmdRespEmbed("Hello Dolly", lines[line])
}