package cmds

import (
	"strings"
	"math/rand"
)

func init() {
	Commands = append(Commands, &Cmd{
		Handler: cmdHelloDolly,
		Matches: []string{"hellodolly", "hd", "hidolly", "dolly"},
		Description: "This is not just a command, it symbolizes the hope and enthusiasm of an entire generation summed up in two words sung most famously by Louis Armstrong: Hello, Dolly. When executed, you will randomly receive a lyric from Hello, Dolly as your response.",
	})
}

var cmdHelloDollyLyrics = `Hello, Dolly
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

func cmdHelloDolly(ctx *CmdCtx) *CmdResp {
	//Split the lyrics by line into a slice
	lyrics := strings.Split(cmdHelloDollyLyrics, "\n")

	//Choose a random line
	line := rand.Intn(len(lyrics))

	//Return the chosen line
	return makeCmdResp(lyrics[line])
}