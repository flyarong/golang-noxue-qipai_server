package domain

import "qipai/enum"

type ReqCreateRoom struct {
	Players   int            `json:"players"`
	Score     enum.ScoreType `json:"score"`
	Pay       enum.PayType   `json:"pay"`
	Count     int            `json:"count"`
	StartType enum.StartType `json:"start"`
	Times     int            `json:"times"`
}
