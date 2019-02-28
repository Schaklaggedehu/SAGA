package emailer

import (
	"SAGA_Crawler/resourcer"
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strconv"
)

var f = fmt.Println
var p = log.Println

func SendResultMail(rentals []resourcer.RentalProperty) {
	f("Sending email...")

	messageContent := parseTemplate(rentals)
	mime := "MIME-version: 1.0;\nContent-Type: text/html;charset=\"UTF-8\";\n\n"
	subject := ""
	if resourcer.DEBUG {
	subject = "Subject: SAGA-Bot DEBUGGING\n"
	}else{
	subject = "Subject: SAGA-Bot hat neue Wohnung gefunden!\n"
	}
	from := resourcer.PersonalInfo.From
	to := []string{resourcer.PersonalInfo.To, resourcer.PersonalInfo.From}
	auth := smtp.PlainAuth("", from, resourcer.PersonalInfo.Password, resourcer.PersonalInfo.Server)
	message := []byte(subject + mime + messageContent)
	server := resourcer.PersonalInfo.Server + ":" + strconv.Itoa(resourcer.PersonalInfo.Port)
	err := smtp.SendMail(server, auth, from, to, message)
	if err != nil {
		f(err)
	}
}

func parseTemplate(rentals []resourcer.RentalProperty) string {

	t, err := template.New("template.html").ParseFiles("SAGA_Crawler_settings/template.html")
	if err != nil {
		f(err)
	}

	buffer := new(bytes.Buffer)
	if err = t.Execute(buffer, rentals); err != nil {
		f(err)
	}
	return buffer.String()
}
