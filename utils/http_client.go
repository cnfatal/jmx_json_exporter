package utils

import (
	"net/http"
	"log"
	"io/ioutil"
)

func Get(url string) (body []byte,err error)  {
	resp,err := http.Get(url)
	if err != nil {
		log.Print(err)
		return
	}
	body,err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return
	}
	return body,nil
}
