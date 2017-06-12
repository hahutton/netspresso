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
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	//"github.com/mediocregopher/radix.v2/redis"
	"github.com/mediocregopher/radix.v2/pool"
)

var masterkey = "SFiUcrzsuV6Ofq3HOP2IYeftiC3Qn1n6nCwDQMRbKThGBiY65sx9FsC47PE6LHmydGfChOJ0pN0tdbT22kOlng=="
var apiversion = "2017-02-22"
var client = &http.Client{}

var p *pool.Pool

type (
	Location struct {
		Name     string `json:"name"`
		Endpoint string `json:"databaseAccountEndpoint"`
	}

	Regions struct {
		Writable []Location `json:"writableLocations"`
		Readable []Location `json:"readableLocations"`
	}

	Database struct {
		Id string `json:"id"`
	}

	Dbs struct {
		List []Database `json:"Databases"`
	}

	Query struct {
		SQL    string   `json:"query"`
		Params []string `json:"parameters"`
	}
)

var writeRegion Location
var readRegions []Location

func init() {
	var err error
	p, err = pool.New("tcp", "localhost:6379", 20)
	if err != nil {
		panic("no redis")
	}
}

func authHeader(verb string, resourceType string, resourceLink string, datetime string) (string, error) {
	verbLower := strings.ToLower(verb)
	datetimeLower := strings.ToLower(datetime)

	body := fmt.Sprintf("%s\n%s\n%s\n%v\n\n", verbLower, resourceType, resourceLink, datetimeLower)
	encoding := base64.StdEncoding
	encodedKey, err := encoding.DecodeString(masterkey)
	if err != nil {
		fmt.Println("Error decoding key", err)
		return "", err
	}

	h := hmac.New(sha256.New, encodedKey)
	h.Write([]byte(body))

	return url.QueryEscape(fmt.Sprintf("type=master&ver=1.0&sig=%s", encoding.EncodeToString(h.Sum(nil)))), nil
}

func SetRegions() {
	var regions = CurrentWriteRegion()
	writeRegion = regions.Writable[0]
	readRegions = regions.Readable
}

func CurrentWriteRegion() Regions {
	verb := "GET"
	datetime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	auth, _ := authHeader(verb, "", "", datetime)

	accountUrl := fmt.Sprintf("https://%s.documents.azure.com/", cosmosAccountName)
	req, err := http.NewRequest("GET", accountUrl, nil)
	req.Header.Add("Authorization", auth)
	req.Header.Add("x-ms-version", apiversion)
	req.Header.Add("x-ms-date", datetime)

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		fmt.Println("err %v", err)
	}

	var regions Regions
	err = json.NewDecoder(resp.Body).Decode(&regions)
	if err != nil {
		fmt.Println("error")
	}

	return regions
}

func ListDbs() {
	verb := "GET"
	datetime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	auth, _ := authHeader(verb, "dbs", "", datetime)

	accountUrl := fmt.Sprintf("%sdbs", readRegions[RandIntn(len(readRegions))].Endpoint)
	req, err := http.NewRequest("GET", accountUrl, nil)
	req.Header.Add("Authorization", auth)
	req.Header.Add("x-ms-version", apiversion)
	req.Header.Add("x-ms-date", datetime)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("err %v", err)
	}

	defer resp.Body.Close()

	var dbs Dbs

	err = json.NewDecoder(resp.Body).Decode(&dbs)
	if err != nil {
		fmt.Println("error dbs")
	}

	fmt.Println(dbs.List[0].Id)

}

func Generator(wg *sync.WaitGroup) {
	defer wg.Done()
	var exit bool
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		select {
		case <-c:
			exit = true
		}
	}()

	for exit != true {
		RunWeightedRandom()
	}

}

func PutDocument() {
	verb := "POST"
	datetime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	auth, _ := authHeader(verb, "docs", "dbs/helo/colls/helo1", datetime)

	accountUrl := fmt.Sprintf("%sdbs/helo/colls/helo1/docs", writeRegion.Endpoint)
	machineDoc := InsertMachineDocument()
	body, _ := json.Marshal(machineDoc)
	document := bytes.NewReader(body)

	req, err := http.NewRequest("POST", accountUrl, document)
	req.Header.Add("Authorization", auth)
	req.Header.Add("x-ms-version", apiversion)
	req.Header.Add("x-ms-date", datetime)
	req.Header.Add("x-ms-documentdb-partitionkey", fmt.Sprintf("[\"%s\"]", machineDoc.Name))
	req.Header.Add("Content-Type", "application/json")

	telem := Event{At: time.Now(), Region: writeRegion, Operation: verb}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("err %v", err)
		return
	}
	defer resp.Body.Close()

	telem.StatusCode = resp.StatusCode
	telem.Duration = time.Since(telem.At) / 1000000 //Want millis not nanos

	err = p.Cmd("RPUSH", "partrowkeys", strings.Join([]string{machineDoc.Name, machineDoc.Id}, ":")).Err
	if err != nil {
		fmt.Println("err %v", err)
		return
	}

	b, _ := ioutil.ReadAll(resp.Body)

	fmt.Printf("Put%s\n", b)
	fmt.Printf("Put%s\n", telem)
	PutTelemetry(telem)
}

func ReadRemote() {
	verb := "GET"
	readRegion := readRegions[RandIntn(len(readRegions))]

	conn, err := p.Get()
	if err != nil {
		fmt.Println(err)
	}
	defer p.Put(conn)
	keyCount, _ := conn.Cmd("LLEN", "partrowkeys").Int()
	choice := RandIntn(keyCount)
	key := conn.Cmd("LINDEX", "partrowkeys", choice).String()
	keys := strings.Split(key, ":")

	datetime := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
	auth, _ := authHeader(verb, "docs", "dbs/helo/colls/helo1", datetime)

	accountUrl := fmt.Sprintf("%sdbs/helo/colls/helo1/docs", readRegion.Endpoint)
	queryDoc := Query{"SELECT * FROM helo1 WHERE (helo1.Name = @partition AND helo1.Id = @row)", []string{keys[0], keys[1]}}
	body, _ := json.Marshal(queryDoc)
	document := bytes.NewReader(body)

	req, err := http.NewRequest(verb, accountUrl, document)
	req.Header.Add("Authorization", auth)
	req.Header.Add("x-ms-version", apiversion)
	req.Header.Add("x-ms-date", datetime)
	req.Header.Add("x-ms-partitionkey", fmt.Sprintf("[\"%s\"]", keys[0]))
	req.Header.Add("Content-Type", "application/query+json")

	telem := Event{At: time.Now(), Region: readRegion, Operation: verb}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("err %v", err)
		return
	}
	defer resp.Body.Close()

	telem.StatusCode = resp.StatusCode
	telem.Duration = time.Since(telem.At) / 1000000 //Want millis not nanos

	b, _ := ioutil.ReadAll(resp.Body)

	fmt.Printf("Read%s\n", b)
	fmt.Printf("Telem%s\n", telem)
	PutTelemetry(telem)
}
