package main

import (
	"flag"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type resource struct {
	url    string
	target string
	start  int
	end    int
}

func ruleResource() []resource {
	var res []resource

	// home page
	r1 := resource{
		url:    "http://localhost:8888",
		target: "",
		start:  0,
		end:    0,
	}

	// list page
	r2 := resource{
		url:    "http://localhost:8888/list/{$id}.html",
		target: "{$id}",
		start:  1,
		end:    21,
	}

	// movie page
	r3 := resource{
		url:    "http://localhost:8888/movies/{$id}.html",
		target: "{$id}",
		start:  1,
		end:    1292,
	}

	res = append(res, r1)
	res = append(res, r2)
	res = append(res, r3)
	return res
}

var uaList = []string{
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.3",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/43.4",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 11_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1",
}

func buildURL(res []resource) []string {
	var list []string

	for _, resItem := range res {
		if len(resItem.target) == 0 {
			list = append(list, resItem.url)
		} else {
			for i := resItem.start; i <= resItem.end; i++ {
				urlStr := strings.Replace(resItem.url, resItem.target, strconv.Itoa(i), -1)
				list = append(list, urlStr)
			}
		}
	}
	return list
}

func makeLog(current, refer, ua string) string {
	u := url.Values{}
	u.Set("time", "1")
	u.Set("url", current)
	u.Set("refer", refer)
	u.Set("ua", ua)
	paramsStr := u.Encode()

	logTemplate := "??????"

	log := strings.Replace(logTemplate, "{$paramsStr}", paramsStr, -1)
	log = strings.Replace(log, "{$ua}", ua, -1)
	return log
}

func randInt(min, max int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	if min > max {
		return max
	}
	return r.Intn(max-min) + min
}

func main() {
	// read command line parameter
	// usage:
	//    go run run.go --total=200 --filePath=/thepath
	total := flag.Int("total", 100, "how many rows be created") // read total, defaul value is 100
	filePath := flag.String("filePath", "/User/", "")

	flag.Parse()

	res := ruleResource()
	list := buildURL(res)
	logs := ""

	for i := 0; i < *total; i++ {
		currentURL := list[randInt(0, len(list)-1)]
		referURL := list[randInt(0, len(list)-1)]
		ua := list[randInt(0, len(uaList)-1)]
		logs = logs + makeLog(currentURL, referURL, ua) + "\n"
		// ioutil.WriteFile(*filePath, []byte(logStr), 0644) // overwrite
	}
	fd, _ := os.OpenFile(*filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fd.Write([]byte(logs))
	fd.Close()
}
