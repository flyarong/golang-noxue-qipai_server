package card

import (
	"reflect"
	"testing"
)

func TestGetPaixing(t *testing.T) {

	var cards []*Card

	card := &Card{
		CardType: CardType_Meihua,
		CardNo:   3,
		CardId:   1,
	}
	cards = append(cards, card)

	card = &Card{
		CardType: CardType_Fangpian,
		CardNo:   7,
		CardId:   2,
	}
	cards = append(cards, card)

	card = &Card{
		CardType: CardType_Heitao,
		CardNo:   13,
		CardId:   3,
	}
	cards = append(cards, card)

	card = &Card{
		CardType: CardType_Hongtao,
		CardNo:   1,
		CardId:   4,
	}
	cards = append(cards, card)

	card = &Card{
		CardType: CardType_Meihua,
		CardNo:   7,
		CardId:   5,
	}
	cards = append(cards, card)

	paixing, cs := getPaixing(cards)
	print(paixing, cs)
}

func TestGetPaixing2(t *testing.T) {
	px, cards, err := GetPaixing("48|16|18|23|1")
	print(px, " ", cards, err)
}

func TestTTTT(t *testing.T) {
	var a interface{}
	c := 1
	a = c
	i := reflect.TypeOf(a)
	print(i.Kind(), i.Kind() == reflect.Array)
}
