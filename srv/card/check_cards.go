package card

import (
	"errors"
	"strconv"
	"strings"
)

func str2Cards(cardStr string) (cards []*Card, err error) {

	css := strings.Split(cardStr, "|")
	if len(css) != 5 {
		err = errors.New("扑克牌字符串格式不对")
		return
	}

	for _, cs := range css {
		n, e := strconv.Atoi(cs)
		if e != nil {
			err = errors.New("扑克数值只能在0-52之间")
			return
		}
		cardType := n%4 + 1
		cardValue := n%13 + 1

		cards = append(cards, &Card{
			CardType: cardType,
			CardNo:   cardValue,
			CardId:   n,
		})
	}
	return
}

/***
数字比较： k>q>j>10>9>8>7>6>5>4>3>2>a
花色比较：黑桃>红桃>梅花>方块。
牌型比较：无牛<有牛<牛牛<五花牛<炸弹<五小牛
无牛牌型比较：取其中最大的一张牌比较大小，牌大的赢，大小相同比花色。
有牛牌型比较：取其中最大的一张牌比较大小，牌大的赢，大小相同比花色。
炸弹之间大小比较：取炸弹牌比较大小。
五小牛牌型比较：庄吃闲。
 */
func CmpCards(cardsStr1, cardsStr2 string) (n int, err error) {
	cards1, e := str2Cards(cardsStr1)
	if e != nil {
		err = e
		return
	}
	cards2, e := str2Cards(cardsStr2)
	if e != nil {
		err = e
		return
	}

	px1, _ := getPaixing(cards1)
	px2, _ := getPaixing(cards2)

	// 不同牛之间大小比较
	if px1 > px2 {
		n = 1
		return
	} else if px1 < px2 {
		n = -1
		return
	}

	// 取最大的牌
	maxCard1 := cards1[0]
	maxCard2 := cards1[0]
	for i := 1; i < len(cards1); i++ {
		if cards1[i].CardNo > maxCard1.CardNo { // 点数大
			maxCard1 = cards1[i]
		} else if cards1[i].CardNo == maxCard1.CardNo && cards1[i].CardType > maxCard1.CardType { // 牌型大
			maxCard1 = cards1[i]
		}

		if cards2[i].CardNo > maxCard2.CardNo { // 点数大
			maxCard2 = cards2[i]
		} else if cards2[i].CardNo == maxCard2.CardNo && cards2[i].CardType > maxCard2.CardType { // 牌型大
			maxCard2 = cards2[i]
		}
	}

	// 点数比较大小
	if maxCard1.CardNo > maxCard2.CardNo {
		n=1
		return
	} else if maxCard1.CardNo < maxCard2.CardNo {
		n=-1
		return
	}

	// 点数一样，比花色，花色不可能一样，所以不是大，就是小，我们默认第一个牌小
	n=-1
	if maxCard1.CardType > maxCard2.CardType {
		n=1
	}
	return
}

func GetPaixing(cardStr string) (paixing int, niuCards string, err error) {
	cards, e := str2Cards(cardStr)
	if e != nil {
		err = e
		return
	}

	css := strings.Split(cardStr, "|")
	if len(css) != 5 {
		err = errors.New("扑克牌字符串格式不对")
		return
	}

	paixing, cards = getPaixing(cards)

	for _, v := range cards {
		ts := strconv.Itoa(v.CardId)
		niuCards += ts + "|"
		pos := -1
		for i, s := range css {
			if s == ts {
				pos = i
				break
			}
		}
		css = append(css[0:pos], css[pos+1:]...)
	}
	for _, v := range css {
		niuCards += v + "|"
	}
	niuCards = niuCards[:len(niuCards)-1]
	return
}

func getPaixing(hand_cards []*Card) (paixing int, niu_cards []*Card) {
	niu_cards = make([]*Card, 0)
	if len(hand_cards) != 5 {
		return DouniuType_Meiniu, niu_cards
	}

	//检查是否有特殊牌型
	huapai_num := 0
	total_score := 0
	same_card_num := 0
	var same_card *Card
	for i := 0; i < 5; i++ {
		if hand_cards[i].CardNo > 10 {
			huapai_num ++
		}
		total_score += hand_cards[i].GetScore()
	}
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 5; j++ {
			if hand_cards[i].SameCardNoAs(hand_cards[j]) {
				same_card = hand_cards[i]
				same_card_num++
			}
		}
	}
	if total_score <= 10 {
		for _, hand_card := range hand_cards {
			niu_cards = append(niu_cards, hand_card)
		}
		return DouniuType_Wuxiao, niu_cards
	}
	if same_card_num == 6 {
		for _, hand_card := range hand_cards {
			if hand_card.SameCardNoAs(same_card) {
				niu_cards = append(niu_cards, hand_card)
			}
		}
		return DouniuType_Zhadan, niu_cards
	}
	if huapai_num == 5 {
		for _, hand_card := range hand_cards {
			niu_cards = append(niu_cards, hand_card)
		}
		return DouniuType_Wuhua, niu_cards
	}

	//检查常规牌型
	left_score := 0
	for i := 0; i < 3; i++ {
		cardi_score := hand_cards[i].GetScore()
		for j := i + 1; j < 4; j++ {
			cardj_score := hand_cards[j].GetScore()
			for k := j + 1; k < 5; k++ {
				cardk_score := hand_cards[k].GetScore()
				three_cards_score := cardi_score + cardj_score + cardk_score
				if three_cards_score%10 == 0 {
					left_score = total_score - three_cards_score
					niu_cards = append(niu_cards, hand_cards[i])
					niu_cards = append(niu_cards, hand_cards[j])
					niu_cards = append(niu_cards, hand_cards[k])
					return GetLeftScorePaixing(left_score), niu_cards
				}
			}
		}
	}

	return DouniuType_Meiniu, niu_cards
}

func GetLeftScorePaixing(score int) int {
	if 0 == score {
		return DouniuType_Meiniu
	}

	left_score := score % 10
	switch left_score {
	case 0:
		return DouniuType_Niuniu
	case 1:
		return DouniuType_Niu1
	case 2:
		return DouniuType_Niu2
	case 3:
		return DouniuType_Niu3
	case 4:
		return DouniuType_Niu4
	case 5:
		return DouniuType_Niu5
	case 6:
		return DouniuType_Niu6
	case 7:
		return DouniuType_Niu7
	case 8:
		return DouniuType_Niu8
	case 9:
		return DouniuType_Niu9
	}

	return DouniuType_Meiniu
}

func GetPaixingMultiple(paixing int) int {
	switch paixing {
	case DouniuType_Meiniu:
		return 1
	case DouniuType_Niu7, DouniuType_Niu8:
		return 2
	case DouniuType_Niu9:
		return 3
	case DouniuType_Niuniu:
		return 4
	case DouniuType_Wuhua:
		return 5
	case DouniuType_Zhadan:
		return 6
	case DouniuType_Wuxiao:
		return 7
	default:
		return 1
	}
}

func GetCardsMaxid(hand_cards []*Card) int {
	if len(hand_cards) != 5 {
		return 0
	}

	maxid := hand_cards[0].CardId
	for i := 1; i < 5; i++ {
		if hand_cards[i].CardId > maxid {
			maxid = hand_cards[i].CardId
		}
	}

	return maxid
}
