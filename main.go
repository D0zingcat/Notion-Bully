package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"log"
	"math/rand"
	"notion-bully/config"
	"notion-bully/mail"
	"notion-bully/notion"
	"notion-bully/util"
	"strings"
	"time"
)


func main() {
	var c config.Config
	toml.DecodeFile("config.toml", &c)

	r := flag.String("r", "", "plz enter your invitation link")
	//fmt.Print("%+v\n", c)
	//return
	flag.Parse()


	for _, m := range c.Mails {
		//break
		// no error check, cannot recognize if this mail has been registered
		for _, i := range m.Data {
			//p, err := proxy_pool.Crawler(c)
			//if err != nil {
			//	log.Println(err)
			//}
			//fmt.Println(p)
			n := notion.NewNotionJob(c.UserAgents, *r, "")
			ms := strings.Split(i, m.Delimiter)
			log.Print(ms)
			n.VisitReferPage()
			n.ValidateMail(c.Notion.Host + c.Notion.Endpoints.MailCheck, ms[0])
			csrf := n.SendCode(c.Notion.Host + c.Notion.Endpoints.SendTmpPass, ms[0])
			log.Println(csrf)
			// sleep for 5 secs
			time.Sleep(time.Second * 30)
			conn := mail.Conn{
				Addr:        m.ReceiveServer,
				Port:        m.ReceivePort,
				SSLOn:       true,
				ReceiveMode: m.Type,
				User:        ms[0],
				Pass:        ms[1],
			}
			var mail mail.MailBusiness = conn
			code := mail.Extract(mail.Read())
			log.Println(code)
			userId := n.Login(c.Notion.Host + c.Notion.Endpoints.TmpPassLogin, csrf, code)
			n.SubmitTransaction(c.Notion.Host + c.Notion.Endpoints.SubmitTransaction, userId, util.GetOneRandom(c.Timezones).(string),
				util.GetOneRandom(c.FirstNames).(string), util.GetOneRandom(c.LastNames).(string),
				util.GetOneRandom(c.Notion.Src).(string), util.GetOneRandom(c.Notion.Job).(string), util.GetOneRandom(c.Notion.Scope).(string))
			n.ActivateRefer(c.Notion.Host + c.Notion.Endpoints.ActivateRefer)
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			time.Sleep(time.Duration(r.Intn(100)) * time.Second)
		}
	}
}
