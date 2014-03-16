package game

import (
	"encoding/base64"
	"fmt"
	"sort"
	dip "github.com/zond/godip/common"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/kol"
)

type AdminGameState struct {
	Game    *Game
	Phases  Phases
	Members []MemberState
}

func AdminGetGame(c *common.HTTPContext) (err error) {
	gameId, err := base64.URLEncoding.DecodeString(c.Vars()["game_id"])
	if err != nil {
		return
	}
	g := &Game{Id: gameId}
	if err = c.DB().Get(g); err != nil {
		return
	}
	members, err := g.Members(c.DB())
	if err != nil {
		return
	}
	memberStates, err := members.ToStates(c.DB(), g, "")
	if err != nil {
		return
	}
	phases, err := g.Phases(c.DB())
	if err != nil {
		return
	}
	sort.Sort(phases)
	return c.RenderJSON(AdminGameState{
		Game:    g,
		Phases:  phases,
		Members: memberStates,
	})
}

func CreateMessage(c common.WSContext) (err error) {
	// load the  message provided by the client
	message := &Message{}
	c.Data().Overwrite(message)
	if message.Recipients == nil {
		message.Recipients = map[dip.Nation]bool{}
	}

	if message.Body == "" {
		return
	}

	// and the game
	game := &Game{Id: message.GameId}
	if err := c.DB().Get(game); err != nil {
		return err
	}
	// and the member
	sender, err := game.Member(c.DB(), c.Principal())
	if err != nil {
		return
	}

	return SendMessage(c.Diet(), game, sender, message)
}

func DeleteMember(c common.WSContext) error {
	return c.Transact(func(c common.WSContext) error {
		decodedId, err := kol.DecodeId(c.Match()[1])
		if err != nil {
			return err
		}
		game := &Game{Id: decodedId}
		if err := c.DB().Get(game); err != nil {
			return fmt.Errorf("Game not found: %v", err)
		}
		if game.State != common.GameStateCreated {
			return fmt.Errorf("%+v already started", game)
		}
		member := Member{}
		if _, err := c.DB().Query().Where(kol.And{kol.Equals{"GameId", decodedId}, kol.Equals{"UserId", kol.Id(c.Principal())}}).First(&member); err != nil {
			return err
		}
		if err := c.DB().Del(&member); err != nil {
			return err
		}
		left, err := game.Members(c.DB())
		if err != nil {
			return err
		}
		if len(left) == 0 {
			if err := c.DB().Del(game); err != nil {
				return err
			}
		}
		return nil
	})
}

func AddMember(c common.WSContext) error {
	var state GameState
	c.Data().Overwrite(&state)
	return c.Transact(func(c common.WSContext) error {
		game := Game{Id: state.Game.Id}
		if err := c.DB().Get(&game); err != nil {
			return fmt.Errorf("Game not found")
		}
		if game.State != common.GameStateCreated {
			return fmt.Errorf("%+v already started")
		}
		variant, found := common.VariantMap[game.Variant]
		if !found {
			return fmt.Errorf("Unknown variant %v", game.Variant)
		}
		if alreadyMember, err := game.Member(c.DB(), c.Principal()); err != nil {
			return err
		} else if alreadyMember != nil {
			return fmt.Errorf("%+v is already member of %v", alreadyMember, game.Id)
		}
		me := &user.User{Id: kol.Id(c.Principal())}
		if err := c.DB().Get(me); err != nil {
			return err
		}
		if game.Disallows(me) {
			return fmt.Errorf("Is not allowed to join this game due to game settings")
		}
		already, err := game.Members(c.DB())
		if err != nil {
			return err
		}
		if disallows, err := already.Disallows(c.DB(), me); err != nil {
			return err
		} else if disallows {
			return fmt.Errorf("Is not allowed to join this game due to blacklistings")
		}
		if len(already) < len(variant.Nations) {
			member := Member{
				GameId:           state.Game.Id,
				UserId:           kol.Id(c.Principal()),
				PreferredNations: state.Members[0].PreferredNations,
			}
			if err := c.DB().Set(&member); err != nil {
				return err
			}
			if len(already) == len(variant.Nations)-1 {
				if err := game.start(c.Diet()); err != nil {
					return err
				}
				c.Infof("Started %v", game.Id)
			}
		}
		return nil
	})
}

func Create(c common.WSContext) error {
	var state GameState
	c.Data().Overwrite(&state)

	game := &Game{
		Variant:          state.Game.Variant,
		EndYear:          state.Game.EndYear,
		Private:          state.Game.Private,
		SecretEmail:      state.Game.SecretEmail,
		SecretNickname:   state.Game.SecretNickname,
		SecretNation:     state.Game.SecretNation,
		Deadlines:        state.Game.Deadlines,
		ChatFlags:        state.Game.ChatFlags,
		AllocationMethod: state.Game.AllocationMethod,
	}

	if _, found := common.VariantMap[game.Variant]; !found {
		return fmt.Errorf("Unknown variant for %+v", game)
	}

	if _, found := common.AllocationMethodMap[game.AllocationMethod]; !found {
		return fmt.Errorf("Unknown allocation method for %+v", game)
	}

	member := &Member{
		UserId:           kol.Id(c.Principal()),
		PreferredNations: state.Members[0].PreferredNations,
	}
	return c.Transact(func(c common.WSContext) error {
		if err := c.DB().Set(game); err != nil {
			return err
		}
		member.GameId = game.Id
		return c.DB().Set(member)
	})
}
