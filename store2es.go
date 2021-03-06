package main

import (
  // "os"
	//"bufio"
  "fmt"
  "time"
	//"encoding/base64"
	//"net/http"
	//"github.com/gorilla/mux"
	"strings"
  "strconv"
  "encoding/json"
  "github.com/kellydunn/golang-geo"
  "github.com/xuyu/goredis"
  "gopkg.in/olivere/elastic.v2"
)

func isJSON(s string) bool {
    var js map[string]interface{}
    return json.Unmarshal([]byte(s), &js) == nil

}

func FetchFromRedis(messages chan string, RedisClient *goredis.Redis) {

	data, _ := RedisClient.RPop("Tqueue")

	if data != nil {
		//if isJSON(string(data)) {
			fmt.Println("Data: ", string(data))
			messages <- string(data)
		//} else {
		//	fmt.Println("Data not in JSON format...")
		//}
	}
}

func Store2ES(ElasticClient *elastic.Client, val string) {
	index := time.Now().Format("2006.01.02")

	res, err := ElasticClient.Index().
	Index("t4u-" + index).
	Type("gps").
	Id(time.Now().Format("20060102150405")).
	BodyString( val ).
	Do()

	if err != nil {
		fmt.Println("After ES insert:", err)
		// file.WriteString(("Error: " + val + "\n"))
	} 

	fmt.Println("ES response - status:", res.Created, ", id:", res.Id, ", index:", res.Index,", type:", res.Type)
}

func StoreDistance(val string) string {
	fmt.Println("Current val: ", val)

	dataPoints := strings.Split(val, "},")

	coordinates := []*geo.Point{}

	 total := 0.000
	
	 for i := range dataPoints {
	 	values := strings.Split(dataPoints[i], ",")

	 	lat,_ := strconv.ParseFloat(values[3], 64)
	 	lng,_ := strconv.ParseFloat(values[5], 64)

    //fmt.Println(lat, " ", lng)
	  pt := geo.NewPoint(lat, lng)
	 	coordinates = append(coordinates, pt)

	 	if i > 0 {
	 		total += coordinates[i-1].GreatCircleDistance(coordinates[i])
	 	}

	 	//fmt.Println("\n\nResult[", i, "]: \n", dataPoints[i])
		
	 }

	 dist := ", \"distance\":" + strconv.FormatFloat(total, 'f', 15, 64)

	 result_final := strings.Join(dataPoints, "},")
	 result_final = result_final[0:len(result_final)-1] + dist + "}" 
	
	// fmt.Println("great circle distance:", total)
	 fmt.Println("\n\nResult_final:   ", result_final)

	 return result_final

}

func main() {

	ElasticClient, err := elastic.NewClient()
	if err != nil {
		fmt.Println( "ES start error: ", err)
	} else {
		fmt.Println( "Connection to ES server established successfully...")
	}

	RedisClient, err2 := goredis.Dial(&goredis.DialConfig{Address: "0.0.0.0:6379"})
	if err2 != nil {
		fmt.Println( "Redis start error: ", err2)
	} else {
		fmt.Println( "Connection to redis-server established successfully...")
	}

	// file, err3 := os.OpenFile("index.txt", os.O_APPEND|os.O_WRONLY, 0666)
 //  if err3 != nil {
	// 	fmt.Println( "Redis start error: ", err3)
	// } else {
	// 	fmt.Println( "File opened successfully...")
	// }

	messages := make(chan string)

	go func() {
		for {
			FetchFromRedis(messages, RedisClient)
		}
	}()

	for {
		select {
			case val := <- messages:
				//val = StoreDistance(val)
				Store2ES(ElasticClient, val)
		}
	}
}
