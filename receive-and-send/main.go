package main

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"path"
	"qqBot/mybotgo"
	"qqBot/mybotgo/websocket"
	"qqBot/mytoken"
	"runtime"
	"strings"
	"time"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/dto/message"
	"github.com/tencent-connect/botgo/event"
)

// 消息处理器，持有 openapi 对象
var processor Processor

func main() {

	// 获取配置文件中的 appId 和 token 信息
	appId, token, err := getConfigInfo("../config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	botToken := mytoken.BotToken(appId, token)

	// 沙箱
	//api := mybotgo.NewSandboxOpenAPI(botToken).WithTimeout(3 * time.Second)
	// 正式
	api := mybotgo.NewOpenAPI(botToken).WithTimeout(3 * time.Second)

	ctx := context.Background()
	// 获取 websocket 信息
	wsInfo, err := api.WS(ctx, nil, "")
	if err != nil {
		log.Fatal(err)
	}

	processor = Processor{Api: api}

	words := getWordsFromFile()

	// websocket.RegisterResumeSignal(syscall.SIGUSR1)
	// 根据不同的回调，生成 intents
	intent := websocket.RegisterHandlers(
		// at 机器人事件
		ATMessageEventHandler(words),
	)

	err = mybotgo.NewSessionManager().Start(wsInfo, botToken, &intent)
	if err != nil {
		log.Fatal(err)
	}
}

// ATMessageEventHandler 实现处理 at 消息的回调
func ATMessageEventHandler(words map[string]string) event.ATMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSATMessageData) error {
		input := strings.ToLower(message.ETLInput(data.Content))
		return processor.ProcessMessage(input, data, words)
	}
}

// 获取配置文件中的信息
func getConfigInfo(fileName string) (uint64, string, error) {
	// 获取当前go程调用栈所执行的函数的文件和行号信息
	// 忽略pc和line
	_, filePath, _, ok := runtime.Caller(1)

	if !ok {
		log.Fatal("runtime.Caller(1) 读取失败")
	}
	file := fmt.Sprintf("%s/%s", path.Dir(filePath), fileName)
	var conf struct {
		AppID uint64 `yaml:"appid"`
		Token string `yaml:"token"`
	}
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Print("ioutil.ReadFile() 读取失败")
		return 0, "", err
	}
	if err = yaml.Unmarshal(content, &conf); err != nil {
		log.Print("yaml.Unmarshal(content, &conf) 读取失败")
		return 0, "", err
	}
	return conf.AppID, conf.Token, nil
}