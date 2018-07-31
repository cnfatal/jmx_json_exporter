package utils

import (
	"net/http"
	"log"
	"io/ioutil"
)

func Get(url string) (body []byte)  {
	emptyBody := []byte("{}")
	resp,err := http.Get(url)
	if err != nil {
		log.Print(err.Error() + ",return empty")
		return emptyBody
	}
	if resp.StatusCode != http.StatusOK  {
		log.Print("response not ok,return empty")
		return emptyBody
	}
	body,err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return emptyBody
	}
	return body
}
