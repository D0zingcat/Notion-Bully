# Notion Bully

> Evil hack of auto-invitation machine of [Notion](https://www.notion.so/).


As this project is evil, I would not like to say too much about it. 
Just use it for fun.

## Instructions

1. Prepare a lot of mail accounts
2. Edit `config.toml` for:

```
[[mails]]
domain = "mail.ru"
receive_server = "imap.mail.ru"
receive_port = 993
type = "IMAP"
delimiter = "----"
data = ["abc@mail.ru----abc123"]


[[mails]]
doman = "outlook.com"
receive_server = "imap-mail.outlook.com"
receive_port = 993
type = "IMAP"
delimiter = "----"
# format: "username{delimiter}password"
data = ["abc@hotmail.com----abc123", "def@hotmail.com----def123"]
```

Currently only POP3 is supported, please use the mails supporting POP3 protocol.

3. In Notion, go to Settings & Memebers -> Earn Credit, copy your invitation link.

4. use go run main.go -r {your link} and enjoy.