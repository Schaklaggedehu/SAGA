package automater

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"regexp"
	"strings"
	"time"
)

var f = fmt.Println
var p = log.Println

func RequestExposes(urls []string) {
	f("Requesting exposés...")
	for _, url := range urls {

		var err error

		// create context
		ctxt, cancel := context.WithCancel(context.Background())
		defer cancel()

		// create chrome instance
		c, err := chromedp.New(ctxt)
		if err != nil {
			f(err)
		}

		// run task list
		cp := &captchaProcessor{}
		err = c.Run(ctxt, filloutForm(url, cp))
		if err != nil {
			f(err)
		}
		err = c.Run(ctxt, filloutCapture(cp))
		if err != nil {
			f(err)
		}

		// shutdown chrome
		err = c.Shutdown(ctxt)
		if err != nil {
			f(err)
		}

		// wait for chrome to finish
		err = c.Wait()
		if err != nil {
			f(err)
		}
	}

}

type captchaProcessor struct {
	Captcha string
}

func (c *captchaProcessor) toString() string {
	regexp, _ := regexp.Compile(`\|[\w\|]+\|`)
	center := regexp.FindString(c.Captcha)
	captchaSlice := strings.Split(center, "|")
	return strings.Join(captchaSlice[:], "")
}

func filloutForm(url string, cp *captchaProcessor) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitVisible(`#propContactBtn`, chromedp.ByID),
		chromedp.Click(`#propContactBtn`, chromedp.ByID),
		chromedp.WaitVisible(`#propContactSendBtn`, chromedp.ByID),
		chromedp.SendKeys(`#salutationSelectBoxIt`, "Frau", chromedp.ByID),
		chromedp.SendKeys(`#name`, "Larissa", chromedp.ByID),
		chromedp.SendKeys(`#surname`, "Schmidt", chromedp.ByID),
		chromedp.SendKeys(`#street`, "Rahlstedter Straße", chromedp.ByID),
		chromedp.SendKeys(`#number`, "172", chromedp.ByID),
		chromedp.SendKeys(`#zip`, "22143", chromedp.ByID),
		chromedp.SendKeys(`#city`, "Hamburg", chromedp.ByID),
		chromedp.SendKeys(`#email`, "lar.schmidt@web.de", chromedp.ByID),
		chromedp.InnerHTML(`.five-twelfths>.gw>.one-whole>pre`, &cp.Captcha, chromedp.ByQuery),
		chromedp.Click(`.iCheck-helper`, chromedp.ByQuery),
	}
}
func filloutCapture(cp *captchaProcessor) chromedp.Tasks {

	return chromedp.Tasks{
		chromedp.SendKeys(`#captcha-input`, cp.toString(), chromedp.ByID),

		chromedp.Click(`.btn-close`, chromedp.ByQuery), //TODO deactiate
		//chromedp.Click(	`#propContactSendBtn`, chromedp.ByID),//TODO activate

		chromedp.Sleep(time.Duration(time.Second * 4)),
	}
}
