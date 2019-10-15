package tools

import (
	"github.com/gin-gonic/gin"
	"sort"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"github.com/tidwall/gjson"
	"io/ioutil"
	localcache "github.com/patrickmn/go-cache"
	"time"
	"encoding/xml"
	"fmt"
	"strings"
	"bytes"
)

func CheckToken(c *gin.Context) {
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")
	token := "sunshine"
	l := []string{token, timestamp, nonce}
	sort.Sort(sort.StringSlice(l))

	s := l[0] + l[1] + l[2]
	h := sha1.New()
	h.Write([]byte(s))

	result := hex.EncodeToString(h.Sum(nil))

	if result == signature {
		c.String(200, echostr)
	} else {
		c.String(200, "")
	}
}

type Data struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName,CDATA"`
	FromUserName string   `xml:"FromUserName,CDATA"`
	CreateTime   string   `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType,CDATA"`
	Event        string   `xml:"Event,CDATA"`
	Content      string   `xml:"Content,CDATA"`
	MsgId        string   `xml:"MsgId"`
}

const responseXml = "<xml><ToUserName><![CDATA[%s]]></ToUserName><FromUserName><![CDATA[%s]]></FromUserName><CreateTime>%s</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[%s]]></Content></xml>";

func RobotResponse(c *gin.Context) {
	buf := make([]byte, 1024)
	n, _ := c.Request.Body.Read(buf)

	body := strings.Replace(string(buf[0:n]), " ", "", -1)
	data := Data{}
	err := xml.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println(err)
	}

	if data.MsgType == "text" {
		c.String(200, fmt.Sprintf(responseXml, data.FromUserName, data.ToUserName, data.CreateTime, TulingRobotResponse(data.Content)))
	} else if data.MsgType == "event" && data.Event == "subscribe" {
		c.String(200, fmt.Sprintf(responseXml, data.FromUserName, data.ToUserName, data.CreateTime, "欢迎你，我是图灵机器人，欢迎和我聊天哦！"))
	} else {
		c.String(200, fmt.Sprintf(responseXml, data.FromUserName, data.ToUserName, data.CreateTime, "抱歉，我只能理解文字。"))
	}
}

var c = localcache.New(60*time.Minute, 90*time.Minute)

func GetAccessToken() string {
	accessToken, found := c.Get("access_token")
	if found {
		return accessToken.(string)
	}
	resp, err := http.Get("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=wx21c2f93eb89ace82&secret=ba13ff1a8175dae97ceb26f8b267d788")
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)

	token := gjson.Get(string(result), "access_token").String()
	c.Set("access_token", token, 60*time.Minute)
	return token
}

func TulingRobotResponse(info string) string {
	var jsonStr = []byte(fmt.Sprintf(`{"key":"4b31d0b70a214467b1b822169aaae3c7","info":"%s"}`, info))
	resp, err := http.Post("http://www.tuling123.com/openapi/api", "application/json", bytes.NewBuffer(jsonStr))
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)

	text := gjson.Get(string(result), "text").String()
	return text
}

type Focus struct {
	Title    string `json:"title"`
	Url      string `json:"url"`
	ImageUrl string `json:"imageUrl"`
}

func TouTiaoFocus() []Focus {
	news := make([]Focus, 0)

	resp, err := http.Get("https://www.toutiao.com/api/pc/focus/");
	if err != nil || resp.StatusCode != http.StatusOK {
		return news
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)

	focus := gjson.Get(string(result), "data.pc_feed_focus")
	focus.ForEach(func(key, value gjson.Result) bool {
		title := value.Get("title").String();
		imageUrl := value.Get("image_url").String();
		url := fmt.Sprintf("https://www.toutiao.com%s", value.Get("display_url").String())
		f := Focus{title, url, imageUrl}
		news = append(news, f)
		return true
	})
	return news
}
