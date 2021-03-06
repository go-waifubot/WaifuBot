package disc

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/Karitham/WaifuBot/database"
	"github.com/Karitham/WaifuBot/query"
	"github.com/diamondburned/arikawa/v2/api"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/utils/sendpart"
)

// Dropper is used to handle the dropping mechanism
type Dropper struct {
	Waifu   map[discord.ChannelID]query.CharStruct
	ChanInc map[discord.ChannelID]uint64
	Mux     *sync.Mutex
}

func (bot *Bot) drop(m *gateway.MessageCreateEvent) {
	var err error

	bot.dropper.Waifu[m.ChannelID], err = query.CharSearchByPopularity(bot.seed.Uint64() % c.MaxCharacterRoll)
	if err != nil {
		log.Println(err)
		return
	}

	// Sanitize the name so it's claimable through discord (some characters have double spaces in their name)
	bot.dropper.Waifu[m.ChannelID].Page.Characters[0].Name.Full =
		strings.Join(strings.Fields(bot.dropper.Waifu[m.ChannelID].Page.Characters[0].Name.Full), " ")

	// get Image
	f, err := http.Get(bot.dropper.Waifu[m.ChannelID].Page.Characters[0].Image.Large)
	if err != nil {
		log.Println(err)
	}
	defer f.Body.Close()

	embedFile := sendpart.File{Name: "stop_reading_that_nerd.png", Reader: f.Body}

	_, err = bot.Ctx.SendMessageComplex(m.ChannelID, api.SendMessageData{
		Embed: &discord.Embed{
			Title:       "CHARACTER DROP !",
			Description: "Can you guess who it is ?\nUse w.claim to get this character for yourself",
			Image:       &discord.EmbedImage{URL: embedFile.AttachmentURI()},
			Footer: &discord.EmbedFooter{
				Text: "This character's initials are " +
					func(name string) (initials string) {
						for _, v := range strings.Fields(name) {
							initials = initials + strings.ToUpper(string(v[0])) + "."
						}
						return
					}(bot.dropper.Waifu[m.ChannelID].Page.Characters[0].Name.Full),
			},
		},
		Files: []sendpart.File{embedFile},
	})
	if err != nil {
		log.Println(err)
	}
}

// Claim a waifu and adds it to the user's database
func (bot *Bot) Claim(m *gateway.MessageCreateEvent, name ...Name) (*discord.Embed, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("if you want to claim a character, use `claim <name>`")
	}

	// Lock because we are reading from the map
	bot.dropper.Mux.Lock()
	defer bot.dropper.Mux.Unlock()
	c, ok := bot.dropper.Waifu[m.ChannelID]

	if !ok {
		return nil, fmt.Errorf("there is no character to claim")
	}

	if !strings.EqualFold(
		strings.Join(name, " "),
		c.Page.Characters[0].Name.Full,
	) {
		return nil, fmt.Errorf("wrong name entered")
	}

	if ok, _ := database.CharID(bot.dropper.Waifu[m.ChannelID].Page.Characters[0].ID).VerifyWaifu(m.Author.ID); ok {
		return nil, fmt.Errorf("%s, you already own %s", m.Author.Username, bot.dropper.Waifu[m.ChannelID].Page.Characters[0].Name.Full)
	}

	// Add to db
	err := database.CharStruct(bot.dropper.Waifu[m.ChannelID]).AddClaimed(m.Author.ID)
	if err != nil {
		return nil, err
	}

	delete(bot.dropper.Waifu, m.ChannelID)

	return &discord.Embed{
		Title: "Claim successful",
		URL:   c.Page.Characters[0].SiteURL,
		Description: fmt.Sprintf(
			"Well done %s you claimed %d\nIt appears in :\n- %s",
			m.Author.Username, c.Page.Characters[0].ID, c.Page.Characters[0].Media.Nodes[0].Title.Romaji,
		),
		Thumbnail: &discord.EmbedThumbnail{
			URL: c.Page.Characters[0].Image.Large,
		},
	}, nil
}
