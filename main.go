package main

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"wechatDataBackup/pkg/wechat"

	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	// log output format
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	// 定义命令行参数
	resPath := flag.String("path", "", "微信数据路径，例如：E:\\scoop\\persist\\wechatDataBackup\\User\\wxid_xxx")
	chatroomName := flag.String("name", "", "聊天对象的昵称")
	outputPath := flag.String("output", "messages.csv", "输出文件路径")
	flag.Parse()

	// 检查必需参数
	if *resPath == "" || *chatroomName == "" {
		flag.Usage()
		return
	}

	logJack := &lumberjack.Logger{
		Filename:   "./app.log",
		MaxSize:    5,
		MaxBackups: 1,
		MaxAge:     30,
		Compress:   false,
	}
	defer logJack.Close()

	multiWriter := io.MultiWriter(logJack, os.Stdout)
	log.SetOutput(multiWriter)
	log.Println("====================== wechatDataBackup ======================")

	// 从路径中提取 prefix
	prefix := "\\User\\" + getLastPathComponent(*resPath)
	provider, err := wechat.CreateWechatDataProvider(*resPath, prefix)
	if err != nil {
		log.Println("CreateWechatDataProvider failed:", *resPath)
		return
	}

	var chatRoomId string
	for _, user := range provider.ContactList.Users {
		if user.NickName == *chatroomName {
			chatRoomId = user.UserName
			break
		}
	}

	if chatRoomId == "" {
		log.Printf("找不到昵称为 %s 的聊天对象\n", *chatroomName)
		return
	}

	messages, err := provider.WeChatGetMessageListByTime(chatRoomId, 0, 50000, wechat.Message_Search_Backward)
	if err != nil {
		log.Println("WeChatGetMessageListByTime failed:", err)
		return
	}

	// output messages to csv
	csvFile, err := os.Create(*outputPath)
	if err != nil {
		log.Println("Create csv file failed:", err)
		return
	}
	defer csvFile.Close()
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.Write([]string{"timestamp", "msgNum", "username"})
	for _, message := range messages.Rows {
		name := message.UserInfo.ReMark
		if name == "" {
			name = message.UserInfo.NickName
		}
		if message.IsSender == 1 {
			name = provider.SelfInfo.NickName
		}
		if name == "" {
			continue
		}
		csvWriter.Write([]string{strconv.FormatInt(message.CreateTime, 10), "1", name})
	}
	csvWriter.Flush()
}

// 获取路径的最后一个组件
func getLastPathComponent(path string) string {
	// 将路径分割，取最后一个部分
	components := strings.Split(strings.ReplaceAll(path, "\\", "/"), "/")
	if len(components) > 0 {
		return components[len(components)-1]
	}
	return ""
}
