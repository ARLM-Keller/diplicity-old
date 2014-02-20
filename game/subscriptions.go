package game

import (
	"encoding/base64"
	"sort"

	"github.com/zond/diplicity/common"
	"github.com/zond/diplicity/user"
	"github.com/zond/kcwraps/kol"
	"github.com/zond/kcwraps/subs"
)

type MemberState struct {
	*Member
	User *user.User
}

type GameState struct {
	*Game
	Members []MemberState
	Phase   *Phase
}

type GameStates []GameState

func (self GameStates) Len() int {
	return len(self)
}

func (self GameStates) Less(i, j int) bool {
	return self[j].Game.CreatedAt.Before(self[i].Game.CreatedAt)
}

func (self GameStates) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func SubscribeCurrent(c subs.Context) error {
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(kol.Equals{"UserId", kol.Id(c.Principal())})
	s.Call = func(i interface{}, op string) error {
		members := i.([]*Member)
		states := GameStates{}
		for _, member := range members {
			if op == common.DeleteType {
				states = append(states, GameState{
					Game:    &Game{Id: member.GameId},
					Members: []MemberState{MemberState{Member: member}},
				})
			} else {
				game := &Game{Id: member.GameId}
				if err := s.DB().Get(game); err != nil {
					return err
				}
				phase, err := game.LastPhase(c.DB())
				if err != nil {
					return err
				}
				members, err := game.Members(c.DB())
				if err != nil {
					return err
				}
				states = append(states, GameState{
					Game:    game,
					Members: members.toStates(c, game, c.Principal()),
					Phase:   phase,
				})
			}
		}
		if op == subs.FetchType || len(states) > 0 {
			sort.Sort(states)
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(&Member{})
}

func SubscribeGame(c subs.Context) error {
	base64DecodedId, err := base64.URLEncoding.DecodeString(c.Match()[1])
	if err != nil {
		return err
	}
	s := c.Pack().New(c.Match()[0])
	s.Call = func(i interface{}, op string) error {
		game := i.(*Game)
		members, err := game.Members(c.DB())
		if err != nil {
			return err
		}
		isMember := members.Contains(c.Principal())
		if !game.Private || isMember {
			phase, err := game.LastPhase(c.DB())
			if err != nil {
				return err
			}
			return s.Send(GameState{
				Game:    game,
				Members: members.toStates(c, game, c.Principal()),
				Phase:   phase,
			}, op)
		} else if op == subs.FetchType {
			return s.Send(GameState{}, op)
		}
		return nil
	}
	return s.Subscribe(&Game{Id: base64DecodedId})
}

func SubscribeMessages(c subs.Context) (err error) {
	base64DecodedId, err := base64.URLEncoding.DecodeString(c.Match()[1])
	if err != nil {
		return err
	}
	game := &Game{Id: base64DecodedId}
	if err = c.DB().Get(game); err != nil {
		return
	}
	member, err := game.Member(c.DB(), c.Principal())
	if err != nil {
		return
	}
	memberId := member.Id.String()
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(kol.Equals{"GameId", base64DecodedId})
	s.Call = func(i interface{}, op string) error {
		messages := i.([]*Message)
		result := Messages{}
		for _, message := range messages {
			if message.Recipients[memberId] {
				result = append(result, *message)
			}
		}
		if op == subs.FetchType || len(result) > 0 {
			sort.Sort(result)
			return s.Send(result, op)
		}
		return nil
	}
	return s.Subscribe(&Message{})
}

func SubscribeOpen(c subs.Context) error {
	s := c.Pack().New(c.Match()[0])
	s.Query = s.DB().Query().Where(kol.And{kol.Equals{"Closed", false}, kol.Equals{"Private", false}})
	s.Call = func(i interface{}, op string) error {
		games := i.([]*Game)
		states := GameStates{}
		isMember := false
		me := user.EnsureUser(c.DB(), c.Principal())
		for _, game := range games {
			if game.Disallows(me) {
				break
			}
			members, err := game.Members(c.DB())
			if err != nil {
				return err
			}
			if disallows, err := members.Disallows(c.DB(), me); err != nil {
				return err
			} else if disallows {
				break
			}
			isMember = false
			for _, m := range members {
				if string(m.UserId) == c.Principal() {
					isMember = true
					break
				}
			}
			if !isMember {
				phase, err := game.LastPhase(c.DB())
				if err != nil {
					return err
				}
				states = append(states, GameState{
					Game:    game,
					Members: members.toStates(c, game, c.Principal()),
					Phase:   phase,
				})
			}
		}
		if op == subs.FetchType || len(states) > 0 {
			sort.Sort(states)
			return s.Send(states, op)
		}
		return nil
	}
	return s.Subscribe(&Game{})
}
