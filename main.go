package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/piquette/finance-go/chart"
	"github.com/piquette/finance-go/datetime"
	_ "github.com/piquette/finance-go/quote"
	"github.com/valyala/fasthttp"
)

//Open:226974.578125 Low:223636.984375 High:226974.578125 Close:224198.234375 AdjClose:0 Volume:0 Timestamp:1629241200
type ChartType struct {
	Open      int64
	Low       int64
	High      int64
	Close     int64
	Volume    int
	Timestamp int64
}

//load server over port
func main() {
	fasthttphandler()
}

//load html default template
func loadTemplate(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/html")
	tmpl := template.Must(template.ParseFiles("layout.html"))
	tmpl.Execute(ctx, nil)
}

//getUpdateChartValues get chart values on condition
func getUpdateChartValues(ctx *fasthttp.RequestCtx) {

	var data []ChartType
	var q *chart.Iter
	stockname := ctx.QueryArgs().Peek("sname")
	startdatetime := string(ctx.QueryArgs().Peek("sdate"))
	enddatetime := string(ctx.QueryArgs().Peek("edate"))
	params := &chart.Params{
		Symbol:   string(stockname),
		Interval: datetime.FiveMins,
	}
	// fmt.Println("startdatetme ", startdatetime, enddatetime)
	if len(startdatetime) > 0 {
		s, _ := strconv.ParseInt(startdatetime, 10, 64)
		sdate := time.Unix(s, 0)
		params.Start = datetime.New(&sdate)

	}

	if len(enddatetime) > 0 {
		e, _ := strconv.ParseInt(enddatetime, 10, 64)
		edate := time.Unix(e, 0)
		params.End = datetime.New(&edate)
	}

	//retry 5 times if err after retry send null
	for i := 0; i < 5; i++ {
		q = chart.Get(params)
		if q.Err() != nil {
			fmt.Printf("Failed to get data , err : %s , Retrying sleeping for 5 sec \n", q.Err().Error())
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	for q.Next() {
		high := q.Bar().High.IntPart()
		low := q.Bar().Low.IntPart()
		volume := q.Bar().Volume
		open := q.Bar().Open.IntPart()
		close := q.Bar().Close.IntPart()
		timestamp := int64(q.Bar().Timestamp)
		diff := time.Now().Minute() % 5
		if time.Now().Add(-time.Duration(diff)*time.Minute).Unix() < timestamp {
			continue
		}

		data = append(data, ChartType{Low: low, High: high, Volume: volume, Timestamp: timestamp, Open: open, Close: close})
	}
	ctx.SetContentType("text/json")
	body, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("jSON ENCODE FAILED : %s\n", err.Error())
		fmt.Fprintf(ctx, "")
		return
	}
	fmt.Fprintf(ctx, string(body))
}

//create api server to listen
func fasthttphandler() {
	fmt.Println("open given link into browser  to see the graph : http://127.0.0.1:8080/")
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/":
			loadTemplate(ctx)
		case "/getUpdatedChart":
			getUpdateChartValues(ctx)
		}
	}
	fasthttp.ListenAndServe(":8080", m)
}
