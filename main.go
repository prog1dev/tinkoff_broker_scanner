package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strings"
  "time"
)

const (
  token              = "token"
  timeout            = time.Second * 3
  topGainersFilepath = "top_gainers_path"
  topLosersFilepath  = "top_losers_path"
)

func main() {
  client := &http.Client{
    Timeout: timeout,
  }

  req, err := http.NewRequest("GET", "https://api-invest.tinkoff.ru/openapi/market/stocks", nil)
  if err != nil {
    log.Fatalf("Can't create http request: %s", err)
  }

  req.Header.Add("Authorization", "Bearer "+token)
  resp, err := client.Do(req)
  if err != nil {
    log.Fatalf("Can't send request: %s", err)
  }

  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    log.Fatalf("bad response code '%s' from '%s'", resp.Status, "https://api-invest.tinkoff.ru/openapi/market/stocks")
  }

  respBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalf("Can't read response: %s", err)
  }

  var stocksResp StocksResponse
  err = json.Unmarshal(respBody, &stocksResp)
  if err != nil {
    log.Fatalf("Can't unmarshal response: '%s' \nwith error: %s", string(respBody), err)
  }

  if strings.ToUpper(stocksResp.Status) != "OK" {
    log.Fatalf("request failed, trackingId: '%s'", stocksResp.TrackingID)
  }

  availableStocks := make(map[string]string)

  for _, v := range stocksResp.Payload.Instruments {
    availableStocks[v.Ticker] = ""
  }

  stocks := topGainers()

  result := make([]string, 0)
  result = append(result, fmt.Sprintf("Name\tTicker\tPrice\tChangePercent\tChangeValue\tVolume\tSector"))

  for _, v := range stocks {
    if _, ok := availableStocks[v.Ticker]; ok {
      result = append(result, fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v", v.Name, v.Ticker, v.Price, v.ChangePercent, v.ChangeValue, v.Volume, v.Sector))
    }
  }

  createFile(result, topGainersFilepath)

  stocks = topLosers()

  result = make([]string, 0)
  result = append(result, fmt.Sprintf("Name\tTicker\tPrice\tChangePercent\tChangeValue\tVolume\tSector"))

  for _, v := range stocks {
    if _, ok := availableStocks[v.Ticker]; ok {
      result = append(result, fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v", v.Name, v.Ticker, v.Price, v.ChangePercent, v.ChangeValue, v.Volume, v.Sector))
    }
  }

  createFile(result, topLosersFilepath)
}

func topGainers() []Stock {
  client := &http.Client{
    Timeout: timeout,
  }

  var jsonStr = []byte(`{"filter":[{"left":"change","operation":"nempty"},{"left":"type","operation":"in_range","right":["stock","dr","fund"]},{"left":"subtype","operation":"in_range","right":["common","","etf","unit","mutual","money","reit","trust"]},{"left":"exchange","operation":"in_range","right":["AMEX","NASDAQ","NYSE"]},{"left":"change","operation":"greater","right":0},{"left":"close","operation":"in_range","right":[2,10000]}],"options":{"active_symbols_only":true,"lang":"en"},"symbols":{"query":{"types":[]},"tickers":[]},"columns":["name","close","change","change_abs","Recommend.All","volume","market_cap_basic","price_earnings_ttm","earnings_per_share_basic_ttm","number_of_employees","sector","description","name","type","subtype","update_mode","pricescale","minmov","fractional","minmove2"],"sort":{"sortBy":"change","sortOrder":"desc"},"range":[0,150]}`)
  req, err := http.NewRequest("POST", "https://scanner.tradingview.com/america/scan", bytes.NewBuffer(jsonStr))

  resp, err := client.Do(req)
  if err != nil {
    log.Fatalf("Can't send request: %s", err)
  }

  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    log.Fatalf("bad response code '%s' from '%s'", resp.Status, "https://scanner.tradingview.com/america/scan")
  }

  respBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalf("Can't read response: %s", err)
  }

  var stocksResp TradingviewRespose
  err = json.Unmarshal(respBody, &stocksResp)
  if err != nil {
    log.Fatalf("Can't unmarshal response: '%s' \nwith error: %s", string(respBody), err)
  }

  log.Printf("stocksResp len: %v", stocksResp.TotalCount)

  stocks := make([]Stock, 0)
  for _, v := range stocksResp.Data {
    stocks = append(stocks, Stock{
      Ticker:        v.D[0].(string),
      Name:          v.D[11].(string),
      Price:         v.D[1].(float64),
      ChangePercent: v.D[2].(float64),
      ChangeValue:   v.D[3].(float64),
      Volume:        v.D[5].(float64),
      // Rating:        v.D[4].(float64),
    })
    // Sector:        v.D[10].(string),
  }

  return stocks
}

func topLosers() []Stock {
  client := &http.Client{
    Timeout: timeout,
  }

  var jsonStr = []byte(`{"filter":[{"left":"change","operation":"nempty"},{"left":"type","operation":"in_range","right":["stock","dr","fund"]},{"left":"subtype","operation":"in_range","right":["common","","etf","unit","mutual","money","reit","trust"]},{"left":"exchange","operation":"in_range","right":["AMEX","NASDAQ","NYSE"]},{"left":"change","operation":"less","right":0},{"left":"close","operation":"in_range","right":[2,10000]}],"options":{"active_symbols_only":true,"lang":"en"},"symbols":{"query":{"types":[]},"tickers":[]},"columns":["name","close","change","change_abs","Recommend.All","volume","market_cap_basic","price_earnings_ttm","earnings_per_share_basic_ttm","number_of_employees","sector","description","name","type","subtype","update_mode","pricescale","minmov","fractional","minmove2"],"sort":{"sortBy":"change","sortOrder":"asc"},"range":[0,150]}`)
  req, err := http.NewRequest("POST", "https://scanner.tradingview.com/america/scan", bytes.NewBuffer(jsonStr))

  resp, err := client.Do(req)
  if err != nil {
    log.Fatalf("Can't send request: %s", err)
  }

  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    log.Fatalf("bad response code '%s' from '%s'", resp.Status, "https://scanner.tradingview.com/america/scan")
  }

  respBody, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatalf("Can't read response: %s", err)
  }

  var stocksResp TradingviewRespose
  err = json.Unmarshal(respBody, &stocksResp)
  if err != nil {
    log.Fatalf("Can't unmarshal response: '%s' \nwith error: %s", string(respBody), err)
  }

  log.Printf("stocksResp len: %v", stocksResp.TotalCount)

  stocks := make([]Stock, 0)
  for _, v := range stocksResp.Data {
    stocks = append(stocks, Stock{
      Ticker:        v.D[0].(string),
      Name:          v.D[11].(string),
      Price:         v.D[1].(float64),
      ChangePercent: v.D[2].(float64),
      ChangeValue:   v.D[3].(float64),
      Volume:        v.D[5].(float64),
      // Rating:        v.D[4].(float64),
    })
    // Sector:        v.D[10].(string),
  }

  return stocks
}

type TradingviewRespose struct {
  Data       []TradingviewStock `json:"data"`
  TotalCount int32              `json:"totalCount"`
}

type TradingviewStock struct {
  S string        `json:"s"`
  D []interface{} `json:"d"`
}

type StocksResponse struct {
  TrackingID string  `json:"trackingId"`
  Status     string  `json:"status"`
  Payload    Payload `json:"payload"`
}

type Stock struct {
  Ticker        string
  Name          string
  Price         float64
  ChangePercent float64
  ChangeValue   float64
  Rating        float64
  Volume        float64
  MarketCap     float64
  Sector        string
}

type Payload struct {
  Instruments []Instrument `json:"instruments"`
  Total       int32        `json:"total"`
}

type Instrument struct {
  Figi              string  `json:"figi"`
  Ticker            string  `json:"ticker"`
  Isin              string  `json:"isin"`
  MinPriceIncrement float32 `json:"minPriceIncrement"`
  Lot               int32   `json:"lot"`
  Currency          string  `json:"currency"`
  Name              string  `json:"name"`
  Type              string  `json:"type"`
}

func createFile(array []string, filepath string) error {
  f, err := os.Create(filepath)
  if err != nil {
    return fmt.Errorf("error creating %s file: %v", filepath, err)
  }
  defer func() {
    if err := f.Close(); err != nil {
      log.Printf("error closing %s file: %v", filepath, err)
    }
  }()

  result := make([]string, 0)
  for _, v := range array {
    if v != "" {
      result = append(result, strings.TrimSpace(v))
    }
  }

  log.Printf("result len: %v", len(result))
  if _, err := f.WriteString(strings.Join(result, "\n")); err != nil {
    return fmt.Errorf("error writing to %s file: %v", filepath, err)
  }

  log.Printf("the file %s has been written\n", filepath)

  return nil
}

func revertSlice(slice []string) {
  for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
    slice[i], slice[j] = slice[j], slice[i]
  }
}
