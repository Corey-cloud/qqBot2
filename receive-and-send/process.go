package main

import (
	"context"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/dto/message"
	"log"
	"qqBot/mybotgo/mydto"
	"qqBot/myopenapi"
	"strings"
)

// Processor is a struct to process message
type Processor struct {
	Api myopenapi.OpenAPI
}

const (
	CmdWordDragon     = "成语接龙"
	CmdStopWordDragon = "停止接龙"
	CmdExplainWord    = "查看释义"
)
const (
	StopTip           = "欢迎下次使用！"
	ToStartTip        = "接龙还没有开始哦！输入[成语接龙]开始接龙游戏！"
	NotWordTip        = "输入的不是成语哦,再试试吧！"
	NotMatchDragonTip = "输入的成语没有接到上一个成语哦,再试试吧！"
	NormalTip         = "输入[成语接龙]开始游戏！游戏中可回复[查看释义]查看成语含义！"
)

var lastWord string
var play = false

// ProcessMessage is a function to process message
func (p Processor) ProcessMessage(input string, data *dto.WSATMessageData, ws WordsMap) error {
	ctx := context.Background()
	// 获取命令
	cmd := strings.Replace(message.ParseCommand(input).Cmd, "/", "", -1)
	toCreate := &mydto.MessageToCreate{
		Content: NormalTip,
		MessageReference: &mydto.MessageReference{
			MessageID:             data.ID,
			IgnoreGetMessageError: true,
		},
	}
	beginWord := ws.getBeginWord()

	switch cmd {
	case CmdWordDragon:
		play = true
		toCreate.Content = beginWord
		p.sendMsg(ctx, data.ChannelID, toCreate)
		lastWord = beginWord
	case CmdStopWordDragon:
		if play {
			play = false
			toCreate.Content = StopTip
			p.sendMsg(ctx, data.ChannelID, toCreate)
		} else {
			toCreate.Content = ToStartTip
			p.sendMsg(ctx, data.ChannelID, toCreate)
		}
	default:
		if play {
			if ws.isWordLegal(cmd) && ws.isWordDragon(cmd, lastWord) {
				nextWord := ws.getWord(cmd)
				toCreate.Content = ToStartTip
				p.sendMsg(ctx, data.ChannelID, toCreate)
				lastWord = nextWord
			} else if cmd == CmdExplainWord {
				toCreate.Content = ws.getWordMeaning(lastWord)
				p.sendMsg(ctx, data.ChannelID, toCreate)
			} else if !ws.isWordLegal(cmd) {
				toCreate.Content = NotWordTip
				p.sendMsg(ctx, data.ChannelID, toCreate)
			} else if !ws.isWordDragon(cmd, lastWord) {
				toCreate.Content = NotMatchDragonTip
				p.sendMsg(ctx, data.ChannelID, toCreate)
			}
		} else {
			toCreate.Content = NotMatchDragonTip
			p.sendMsg(ctx, data.ChannelID, toCreate)
		}
	}
	return nil
}

// 发送消息
func (p Processor) sendMsg(ctx context.Context, channelID string, toCreate *mydto.MessageToCreate) {
	_, err := p.Api.PostMessage(ctx, channelID, toCreate)
	if err != nil {
		log.Println(err)
	}
}