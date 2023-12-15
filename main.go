package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/XiaoMiku01/biliup-go/login"
	"github.com/tidwall/gjson"

	"github.com/AceXiamo/blivedm-go/api"
	"github.com/AceXiamo/blivedm-go/client"
	"github.com/AceXiamo/blivedm-go/message"
	_ "github.com/AceXiamo/blivedm-go/utils"
	log "github.com/sirupsen/logrus"
)

var (
	userInfoApi = "https://api.bilibili.com/x/space/wbi/acc/info?mid="
	userAgent   = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36"
)

func sendDanmaku(roomid int, msg string, cookie CookieType) error {
	dmReq := &api.DanmakuRequest{
		Msg:      msg,
		RoomID:   fmt.Sprint(roomid),
		Bubble:   "0",
		Color:    "16777215",
		FontSize: "25",
		Mode:     "1",
		DmType:   "0",
	}
	var biliJct, sessData string
	for _, v := range cookie.Data.CookieInfo.Cookies {
		if v.Name == "bili_jct" {
			biliJct = v.Value
		} else if v.Name == "SESSDATA" {
			sessData = v.Value
		}
	}
	if biliJct == "" || sessData == "" {
		return errors.New("cookie lost")
	}
	_, err := api.SendDanmaku(dmReq, &api.BiliVerify{
		Csrf:     biliJct,
		SessData: sessData,
	})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func moveCookie() {
	var cookie CookieType
	file, err := os.ReadFile("cookie.json")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(file, &cookie)
	myClient := http.Client{}
	req, _ := http.NewRequest(http.MethodGet, userInfoApi+fmt.Sprint(cookie.Data.Mid), nil)
	req.Header.Add("User-Agent", userAgent)
	resp, err := myClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	var userInfo UserInfoType
	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(b, &userInfo)
	err = os.Rename("cookie.json", "users/cookie_"+userInfo.Data.Name+".json")
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	log.SetLevel(log.InfoLevel)
	err := LoadConfig()
	if err != nil {
		CreateBlankConfigfile()
		log.Fatal("配置文件不存在，以创建空配置文件，请修改！")
	}
	if cfg.RoomId == 0 {
		log.Println("当前配置文件直播间号有误！请修改配置文件！按回车键退出")
		var input string
		fmt.Scanln(&input)
		os.Exit(0)
	}
}

func main() {
	dir, err := os.ReadDir("users")
	if err != nil {
		os.Mkdir("users", 0775)
	}
	_, err = os.Stat("cookie.json")
	if err == nil {
		moveCookie()
	} else if dir == nil {
		login.LoginBili()
		moveCookie()
	}

	for {
		log.Print("是否要继续添加新用户: (y/N) ")
		var input string
		fmt.Scanln(&input)
		if input == "Y" || input == "y" {
			login.LoginBili()
			moveCookie()
		} else {
			break
		}
	}

	dir, _ = os.ReadDir("users")
	for i, v := range dir {
		// 定义正则表达式模式
		regexPattern := `cookie_(.*?)\.json`
		regex := regexp.MustCompile(regexPattern)
		match := regex.FindStringSubmatch(v.Name())
		var username string
		if len(match) > 1 {
			username = match[1]
		} else {
			log.Println("cookie file name error!")
			continue
		}

		// create client
		i := i
		go func() {
			log.Println("正在添加第", i+1, "个用户：", username)
			c := client.NewClient(cfg.RoomId) // 房间号
			cookieStr := ""
			var cookie CookieType
			file, err := os.ReadFile("users/" + v.Name())
			if err != nil {
				log.Fatal(err)
			}
			json.Unmarshal(file, &cookie)
			for _, v := range cookie.Data.CookieInfo.Cookies {
				cookieStr += v.Name + "=" + v.Value + ";"
			}
			c.SetCookie(cookieStr)
			if i == 0 {
				// 弹幕事件
				c.OnDanmaku(func(danmaku *message.Danmaku) {
					if danmaku.Type == message.EmoticonDanmaku {
						log.Printf("[弹幕表情] %s：表情URL： %s\n", danmaku.Sender.Uname, danmaku.Emoticon.Url)
					} else {
						log.Printf("[弹幕] %s：%s\n", danmaku.Sender.Uname, danmaku.Content)
					}
				})
				// 醒目留言事件
				c.OnSuperChat(func(superChat *message.SuperChat) {
					log.Printf("[SC|%d元] %s: %s\n", superChat.Price, superChat.UserInfo.Uname, superChat.Message)
				})
				// 礼物事件
				c.OnGift(func(gift *message.Gift) {
					if gift.CoinType == "gold" {
						log.Printf("[礼物] %s 的 %s %d 个 共%.2f元\n", gift.Uname, gift.GiftName, gift.Num, float64(gift.Num*gift.Price)/1000)
					}
				})
				// 上舰事件
				c.OnGuardBuy(func(guardBuy *message.GuardBuy) {
					log.Printf("[大航海] %s 开通了 %d 等级的大航海，金额 %d 元\n", guardBuy.Username, guardBuy.GuardLevel, guardBuy.Price/1000)
				})
				// 进入直播间事件
				c.RegisterCustomEventHandler("INTERACT_WORD", func(s string) {
					var v message.InteractWord
					data := gjson.Get(s, "data").String()
					json.Unmarshal([]byte(data), &v)
					log.Println(v.Uname, "进入直播间")
				})
			}

			err = c.Start()
			if err != nil {
				log.Fatal(err)
			}
			if cfg.EnterMessage != "" {
				if err := sendDanmaku(cfg.RoomId, cfg.EnterMessage, cookie); err != nil {
					log.Println(err)
				}
			}
		}()
		time.Sleep(time.Second)
	}
	log.Println("start~")
	select {}
}
