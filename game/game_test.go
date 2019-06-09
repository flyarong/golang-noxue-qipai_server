package game

import "testing"

func TestDoArgs(t *testing.T) {
	g,_:=Games.NewGame(10002)

	g.Start()
	g.SetTimes(10,5)
	g.SetTimes(11,6)
	g.SetTimes(12,7)
	g.SetTimes(13,8)
	g.SetTimes(14,9)

	g.SetScore(10,5)
	g.SetScore(11,6)
	g.SetScore(12,7)
	g.SetScore(13,8)
	g.SetScore(14,9)


	g.ShowCard(10)

	g.CompareCard()

	g.GameOver()
	g.GameOver()
	g.GameOver()


}
