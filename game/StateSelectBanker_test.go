package game

import (
	"fmt"
	"testing"
	"time"
)

func TestCreateCard(t *testing.T) {
	var cards []int
	var n1 = 9

	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond*100)
		result := createCard(cards, n1)
		v := result[3]%13 + 1 + result[4]%13 + 1
		if v != n1 {
			t.Error("错误")
			return
		}
		var tResult []int
		for i := 0; i < 5; i++ {
			tResult = append(tResult,result[i]%13 + 1)
		}
		fmt.Println(result,tResult, v)

	}
}
