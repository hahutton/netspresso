package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//key for sunsun
//var sakey = "M59H+/tfpMLvmZn3fWbHTuqIY1KW2c30MRxRWScNMW4="

//key for orchestra
var sakey = "eQ+D/ow0K7l6nvUjX5W7+/XPe5F1DvcVfdIefxl3Y9M="

//var namespace = "sunsun"
var namespace = "orchestrate"
var name = "pit"

type Event struct {
	At         time.Time
	Region     Location
	Operation  string
	StatusCode int
	Duration   time.Duration
}

func genAuthHeader(host string, ehUrl string, ttl int64) string {
	//TODO
	uriEncoded := url.QueryEscape("https://orchestrate.servicebus.windows.net/")
	fmt.Println("encoded:%s", uriEncoded)
	ttlStr := strconv.FormatInt(ttl, 10)
	toSign := fmt.Sprintf("%s\n%s", uriEncoded, ttlStr)
	fmt.Println()
	fmt.Println(toSign)

	encoding := base64.StdEncoding
	//encodedKey, err := encoding.DecodeString(sakey)
	//if err != nil {
	//	return "bad"
	//}

	h := hmac.New(sha256.New, []byte(sakey))
	h.Write([]byte(toSign))

	authHeader := fmt.Sprintf("SharedAccessSignature sr=%s&sig=%s&se=%s&skn=RootManageSharedAccessKey", uriEncoded, url.QueryEscape(encoding.EncodeToString(h.Sum(nil))), ttlStr)
	fmt.Println(authHeader)
	return authHeader
}

func PutTelemetry(blob interface{}) {
	unixtime := time.Now().UTC().Unix()
	var week int64 = 60 * 60 * 24 * 7
	ttl := unixtime + week
	//ttl = 1498661151
	host := fmt.Sprintf("%s.servicebus.windows.net/pit", namespace)
	accountUrl := fmt.Sprintf("https://%s/messages?api-version=2014-01", host)
	jsonBlob, _ := json.Marshal(blob)
	body := bytes.NewReader(jsonBlob)

	req, err := http.NewRequest("POST", accountUrl, body)
	req.Header.Add("Authorization", genAuthHeader(host, accountUrl, ttl))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("err %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.Status)
	b, _ := ioutil.ReadAll(resp.Body)
	msgResponse := fmt.Sprintf("%v", b)
	fmt.Println("hello telem:")
	fmt.Println(msgResponse)
}
