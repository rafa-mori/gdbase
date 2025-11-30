package cli

import (
	"math/rand"
	"os"
	"strings"
)

var banners = []string{
	`
  ______                             __ __
 /      \                           |  \  \
|  ▓▓▓▓▓▓\ ______  _______   ______ | ▓▓\▓▓________  ______
| ▓▓   \▓▓|      \|       \ |      \| ▓▓  \        \/      \
| ▓▓       \▓▓▓▓▓▓\ ▓▓▓▓▓▓▓\ \▓▓▓▓▓▓\ ▓▓ ▓▓\▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓\
| ▓▓   __ /      ▓▓ ▓▓  | ▓▓/      ▓▓ ▓▓ ▓▓ /    ▓▓| ▓▓    ▓▓
| ▓▓__/  \  ▓▓▓▓▓▓▓ ▓▓  | ▓▓  ▓▓▓▓▓▓▓ ▓▓ ▓▓/  ▓▓▓▓_| ▓▓▓▓▓▓▓▓
 \▓▓    ▓▓\▓▓    ▓▓ ▓▓  | ▓▓\▓▓    ▓▓ ▓▓ ▓▓  ▓▓    \\▓▓     \
  \▓▓▓▓▓▓  \▓▓▓▓▓▓▓\▓▓   \▓▓ \▓▓▓▓▓▓▓\▓▓\▓▓\▓▓▓▓▓▓▓▓ \▓▓▓▓▓▓▓`,
}

const Version = "1.0.0"

func GetDescriptions(descriptionArg []string, _ bool) map[string]string {
	var description, banner string
	if descriptionArg != nil {
		if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
			description = descriptionArg[0]
		} else {
			description = descriptionArg[1]
		}
	} else {
		description = ""
	}
	bannerRandLen := len(banners)
	bannerRandIndex := rand.Intn(bannerRandLen)
	banner = banners[bannerRandIndex]
	return map[string]string{"banner": banner, "description": description}
}
