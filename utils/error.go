package utils

import (
	"errors"
	"fmt"
	log2 "github.com/ethereum/go-ethereum/log"
	"gopkg.in/gomail.v2"
	"log"
	"reflect"
	"strconv"
)

var (
	ErrTooManyRecords     = errors.New("collector: too many records")
	ErrInvalidKey         = errors.New("etherscan: invalid api key")
	ErrEtherscanRateLimit = errors.New("etherscan: rate limit")
	ErrGethRespTooLarge   = errors.New("geth: response too large")
	ErrConnection         = errors.New("read: connection reset by peer")
	ErrChainbaseTooWide   = "block range is too wide"
	ErrNetwork429         = errors.New("request too much: status code 429")
)

const PhalcopnBaseUrl = "https://explorer.phalcon.xyz/tx/%s/%s"

func SendMail(subject string, body string) error {
	//定义收件人
	mailTo := []string{
		"hxhsia@163.com",
	}

	err := sendMail(mailTo, subject, body)
	if err != nil {
		log.SetPrefix("SendMail()")
		log2.Error("failed to send mail", "err", err)
		return err
	}
	return nil
}

func sendMail(mailTo []string, subject string, body string) error {
	//定义邮箱服务器连接信息，如果是网易邮箱 pass填密码，qq邮箱填授权码

	//mailConn := map[string]string{
	//  "user": "xxx@163.com",
	//  "pass": "your password",
	//  "host": "smtp.163.com",
	//  "port": "465",
	//}

	mailConn := map[string]string{
		"user": "hxhsia@163.com",
		"pass": "UBNNRQTVMMTEEUZG",
		"host": "smtp.163.com",
		"port": "465",
	}

	port, _ := strconv.Atoi(mailConn["port"]) //转换端口类型为int

	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(mailConn["user"], "hxh")) //这种方式可以添加别名，即“XX官方”
	//说明：如果是用网易邮箱账号发送，以下方法别名可以是中文，如果是qq企业邮箱，以下方法用中文别名，会报错，需要用上面此方法转码
	//Multichain.SetHeader("From", "FB Sample"+"<"+mailConn["user"]+">") //这种方式可以添加别名，即“FB Sample”， 也可以直接用<code>Multichain.SetHeader("From",mailConn["user"])</code> 读者可以自行实验下效果
	//Multichain.SetHeader("From", mailConn["user"])
	m.SetHeader("To", mailTo...)    //发送给多个用户
	m.SetHeader("Subject", subject) //设置邮件主题
	m.SetBody("text/html", body)    //设置邮件正文

	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])
	err := d.DialAndSend(m)
	return err

}

func GenPhalconTxUrl(chain, hash string) string {
	url := fmt.Sprintf(PhalcopnBaseUrl, chain, hash)
	return url
}

func IsEmpty(s interface{}) bool {
	v := reflect.ValueOf(s)

	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return true
		}
		return IsEmpty(v.Elem().Interface())

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if !field.IsZero() {
				return false
			}
		}
		return true

	default:
		return false
	}
}

func PrintErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
