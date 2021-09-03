package cmds

import (
	"strings"
	"math/rand"
)

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

func HelloDolly(ctx *CmdCtx) *CmdResp {
	//Split the lyrics by line into a slice
	lyrics := strings.Split(cmdHelloDollyLyrics, "\n")

	//Choose a random line
	line := rand.Intn(len(lyrics))

	//Return the chosen line
	return &CmdResp{Messages: []string{lyrics[line]}}
}