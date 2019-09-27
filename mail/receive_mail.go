package mail

import (
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/taknb2nch/go-pop3"
	"log"
	"regexp"
	"strconv"
)

type MailBusiness interface {
	Read() string
	// currently only support IMAP due to getting subject is quite easy
	Extract(subject string) string
}

type Conn struct {
	Addr        string
	Port        int
	SSLOn       bool
	ReceiveMode string
	User        string
	Pass        string
}

func (c Conn) Extract(sub string) string {
	exp := regexp.MustCompile(`.*?"(.*?)"`)
	re := exp.FindStringSubmatch(sub)
	if len(re) > 1 {
		return re[1]
	} else {
		log.Fatal("fail to match code")
	}
	return ""
}

func (c Conn) Read() string {
	switch c.ReceiveMode {
	case "POP3":
		if err := pop3.ReceiveMail(c.Addr+":"+strconv.Itoa(c.Port), c.User, c.Pass,
			func(number int, uid, data string, err error) (bool, error) {
				log.Printf("%d, %s\n", number, uid)
				log.Printf("%v\n", data)
				// implement your own logic here
				return false, nil
			}); err != nil {
			log.Fatalf("%v\n", err)
		}
	case "IMAP":
		cli, err := client.DialTLS(c.Addr+":"+strconv.Itoa(c.Port), nil)
		if err != nil {
			log.Fatal(err)
		}
		defer cli.Logout()
		if err := cli.Login(c.User, c.Pass); err != nil {
			log.Fatal(err)
		}
		// List mailboxes
		//mailboxes := make(chan *imap.MailboxInfo, 10)
		//done := make(chan error, 1)
		//go func () {
		//	done <- cli.List("", "*", mailboxes)
		//}()
		//for m := range mailboxes {
		//	log.Println("* " + m.Name)
		//}
		//if err := <-done; err != nil {
		//	log.Fatal(err)
		//}
		mbox, err := cli.Select("INBOX", false)
		if err != nil {
			log.Fatal(err)
		}
		if mbox.Messages == 0 {
			log.Fatal("There's no message")
		}
		seqSet := new(imap.SeqSet)
		// suppose only one new(from notion) mail in INBOX
		seqSet.AddNum(mbox.Messages)
		// Get the whole message body
		var section imap.BodySectionName
		items := []imap.FetchItem{section.FetchItem()}
		messages := make(chan *imap.Message, 1)
		go func() {
			if err := cli.Fetch(seqSet, items, messages); err != nil {
				log.Fatal(err)
			}
		}()
		msg := <-messages
		if msg == nil {
			log.Fatal("Server didn't returned message")
		}
		r := msg.GetBody(&section)
		if r == nil {
			log.Fatal("Server didn't returned message body")
		}
		// Create a new mail reader
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}
		// Print some info about the message
		header := mr.Header
		//if date, err := header.Date(); err == nil {
		//	log.Println("Date:", date)
		//}
		//if from, err := header.AddressList("From"); err == nil {
		//	log.Println("From:", from)
		//}
		//if to, err := header.AddressList("To"); err == nil {
		//	log.Println("To:", to)
		//}
		if subject, err := header.Subject(); err == nil {
			log.Println("Subject:", subject)
			return subject
		}

		// Process each message's part
		//for {
		//	p, err := mr.NextPart()
		//	if err == io.EOF {
		//		break
		//	} else if err != nil {
		//		log.Fatal(err)
		//	}
		//
		//	switch h := p.Header.(type) {
		//	case *mail.InlineHeader:
		//		// This is the message's text (can be plain-text or HTML)
		//		b, _ := ioutil.ReadAll(p.Body)
		//		log.Println("Got text: %v", string(b))
		//	case *mail.AttachmentHeader:
		//		// This is an attachment
		//		filename, _ := h.Filename()
		//		log.Println("Got attachment: %v", filename)
		//	}
		//}

		//log.Println("Flags for INBOX:", mbox.Flags)
		//from := uint32(1)
		//to := mbox.Messages
		//if mbox.Messages > 3 {
		//	// We're using unsigned integers here, only substract if the result is > 0
		//	from = mbox.Messages - 3
		//}
		//seqset := new(imap.SeqSet)
		//seqset.AddRange(from, to)
		//
		//messages := make(chan *imap.Message, 10)
		//done = make(chan error, 1)
		//go func() {
		//	done <- cli.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
		//}()
		//
		//log.Println("Last 4 messages:")
		//for msg := range messages {
		//	log.Println("* " + msg.Envelope.Subject)
		//	log.Println(msg.GetBody())
		//}
		//
		//if err := <-done; err != nil {
		//	log.Fatal(err)
		//}

	default:
		log.Fatalf("%v mode not supported yet!", c.ReceiveMode)
	}
	return ""
}
