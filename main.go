package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Card struct {
	Number int // шестёрка-туз
	Class  int // 0,1,2,3 - club, diamonds, hearts, spades
}

type Deck []Card

type Player struct {
	Nickname   string // nickname
	Cards      []Card // hand
	BattleCard Card
	ID         int
	State      int // 0 spectator, 1 atacker, 2 defender
}

type Players []Player

type Session struct {
	Deck     Deck    // deck contains not used cards, deck needs to give new cards to players
	Players  Players // list of players
	Turn     int     // number of turn
	Trump    Card    // copy of trump card
	Dumb     Player  // list of won players
	Attacker *Player
	Defender *Player
}

type Stringer interface {
	String() string
}

func (c Card) String() string {
	class := []string{"♣", "♦", "♥", "♠"}
	number := []string{"6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	return fmt.Sprint(number[c.Number] + class[c.Class])
}

func (p Player) String() (s string) {
	for i, e := range p.Cards {
		s += fmt.Sprint("[", i+1, "]", e.String(), " ")
	}
	return
}

func (p Players) String() (s string) {
	for _, e := range p {
		s += fmt.Sprint(e.Nickname, "/", len(e.Cards), ", ")
	}
	return
}

// if game should finish return true
func (s *Session) IsFinish() bool {
	if len(s.Players) == 1 {
		s.Dumb = s.Players[0]
		return true
	}
	return false
}

func (d *Deck) Shuffle() {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(*d), func(i, j int) { (*d)[i], (*d)[j] = (*d)[j], (*d)[i] })
}

func (d *Deck) Create() {
	for i := 0; i != 9; i++ {
		for j := 0; j != 4; j++ {
			card := Card{i, j}
			*d = append(*d, card)
		}
	}
}

// create players by number
func (s *Session) PlayersInit(players int) (err error) {
	randomNicknames := []string{"фембой", "трапик", "жучара", "лох", "гречка"}

	if players < 2 {
		return fmt.Errorf("not enough players")
	} else if players >= 6 {
		return fmt.Errorf("so enough players")
	}

	s.Deck.Create()
	s.Deck.Shuffle()
	s.Trump = s.Deck[len(s.Deck)-1]

	for i := 0; i != players; i++ {
		s.Players = append(
			s.Players,
			Player{
				Cards:    s.Deck[0:6],
				Nickname: randomNicknames[i], //rand.Intn(len(randomNicknames))
				ID:       i})
		s.Deck = s.Deck[6:]
	}
	return nil
}

func (s *Session) Refill(p *Player) {
	if len(p.Cards) < 6 {
		for i := 0; ; i++ {
			if len(s.Deck) == 0 {
				return
			}
			p.Cards = append(p.Cards, s.Deck[0])
			s.Deck = s.Deck[1:]
			if len(p.Cards) == 6 {
				return
			}
		}
	}
}

// if wins defender then true
func (s *Session) Battle() (bool, error) {
	exhaust := []Card{}
	s.Turn += 1

	apc := s.Attacker.BattleCard
	dpc := s.Defender.BattleCard
	s.Attacker.BattleCard = Card{}
	s.Defender.BattleCard = Card{}

	defenderWon := false
	exhaust = append(exhaust, apc, dpc)
	if apc.Class == dpc.Class && apc.Number != dpc.Number {
		if apc.Number < dpc.Number {
			defenderWon = true
		}
	} else if apc.Class != dpc.Class {
		if !(apc.Class == s.Trump.Class || dpc.Class != s.Trump.Class) {
			defenderWon = true
		}
	} else if apc.Class == dpc.Class && apc.Number == dpc.Number {
		return false, fmt.Errorf("doubleCard: ")
	} else {
		return false, fmt.Errorf("unknownError: ")
	}

	s.Refill(s.Attacker)

	if defenderWon {
		s.Refill(s.Defender)
		return true, nil
	} else {
		s.Defender.Cards = append(exhaust, s.Defender.Cards...)
		return false, nil
	}
}

func remove(slice []Card, s int) []Card {
	return append(slice[:s], slice[s+1:]...)
}

// choose card
func (p *Player) GetBattleCard(input string) error {
	inputString := strings.Split(input, "")[0]
	inputNumber, err := strconv.Atoi(inputString)
	if err != nil {
		return err
	}
	p.BattleCard = p.Cards[inputNumber-1]
	p.Cards = remove(p.Cards, inputNumber-1)
	return nil
}

// bot takes card to attack
func (p *Player) BGetBattleCard() error {
	if len(p.Cards) == 0 {
		return fmt.Errorf("indexError: ")
	}
	number := rand.Intn(len(p.Cards))
	p.BattleCard = p.Cards[number]
	p.Cards = remove(p.Cards, number)
	return nil
}

func (s *Session) Update() {
	for i := range s.Players {
		if s.Players[i].State == 1 {
			s.Attacker = &s.Players[i]
		}
		if s.Players[i].State == 2 {
			s.Defender = &s.Players[i]
		}
	}
}

func (p *Players) ChangeState(n int) {
	for i := range *p {
		if (*p)[i].State == 2 {
			(*p)[i].State = 1
			if i+n == len(*p) {
				(*p)[0].State = 2
			} else {
				(*p)[i+n].State = 2
			}
			return
		}
	}
}

func (s Session) Stdout(me int) {
	fmt.Println("-----------")
	fmt.Println("игроки/карт:", s.Players)
	fmt.Println("атакует:", s.Attacker.Nickname)
	fmt.Println("защищается:", s.Defender.Nickname)
	fmt.Println("козыри:", s.Trump)
	fmt.Println("ход:", s.Turn)
	fmt.Println("карт в колоде:", len(s.Deck))
	if p, ok := s.Players.ByID(me); ok {
		fmt.Println("твои карты:", p)
	} else {
		fmt.Println("вы выбыли")
	}

}

func (p *Players) ByID(id int) (*Player, bool) {
	for i := range *p {
		if (*p)[i].ID == id {
			return &(*p)[i], true
		}
	}
	return &Player{}, false
}

func (s *Session) SomeoneGone() (Players, bool) {
	ps := []Player{}
	if len(s.Deck) == 0 {
		for i, e := range s.Players {
			if len(e.Cards) == 0 {
				ps = append(ps, s.Players[i])
				s.Players = append(s.Players[:i], s.Players[i+1:]...)
			}
		}
	}

	if len(ps) != 0 {
		return ps, true
	}
	return ps, false
}

//
func main() {
	var session Session
	err := session.PlayersInit(5) // init N players
	if err != nil {
		panic(err)
	}

	me := 0
	fbot := 1

	session.Attacker, _ = session.Players.ByID(me)
	session.Attacker.State = 1
	session.Defender, _ = session.Players.ByID(fbot)
	session.Defender.State = 2

	fmt.Print("\nвведите свой никнейм: ")
	fmt.Scan(&session.Players[me].Nickname)
	fmt.Println()

	for session.Turn = 1; session.Turn != 100; {
		if session.IsFinish() {
			fmt.Println("игра завершена")
			fmt.Println("дурак -", session.Dumb.Nickname)
			break
		}

		session.Stdout(me)

		if session.Defender.ID == me {
			err = session.Attacker.BGetBattleCard()
			if err != nil {
				panic(err)
			}
			fmt.Println("тебя атакуют этим:", session.Attacker.BattleCard)
			fmt.Println("защищайся")
			fmt.Print("> ")

			input := ""
			fmt.Scan(&input)
			err = session.Defender.GetBattleCard(input)
			if err != nil {
				panic(err)
			}
		} else if session.Attacker.ID == me {
			fmt.Println("атакуй")
			fmt.Print("> ")

			input := ""
			fmt.Scan(&input)
			err = session.Attacker.GetBattleCard(input)
			if err != nil {
				panic(err)
			}

			err = session.Defender.BGetBattleCard()
			if err != nil {
				panic(err)
			}
			fmt.Println("тебя попытались отбить: ", session.Defender.BattleCard)

		} else {
			err = session.Attacker.BGetBattleCard()
			if err != nil {
				panic(err)
			}
			err = session.Defender.BGetBattleCard()
			if err != nil {
				panic(err)
			}
		}

		res, err := session.Battle()
		if err != nil {
			panic(err)
		}
		if res {
			fmt.Println("бито")
		} else {
			fmt.Println("игрок", session.Defender.Nickname, "забрал")
		}

		if gone, yes := session.SomeoneGone(); yes {
			fmt.Println("игроки", gone, "выбыли")
		}

		session.Players.ChangeState(1)
		session.Update()
	}
	fmt.Println("игра окончена всем спасибо")
}
