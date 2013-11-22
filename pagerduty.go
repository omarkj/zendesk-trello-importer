package main

import (
  "strings"
  "fmt"
  "os"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "time"
)

var pagerdutyApiUrl string = os.Getenv("PAGERDUTY_URL")
var pagerdutyApiKey string = os.Getenv("PAGERDUTY_API_KEY")
var pagerdutyRotationIds []string = strings.Split(os.Getenv("PAGERDUTY_ROTATION_IDS"), ",")

type PagerdutyUser struct {
  Id string
  Name string
  Email string
}

type ScheduleEntry struct {
  User PagerdutyUser
}

type ScheduleEntries struct {
  Entries []ScheduleEntry
}

type Pagerduty struct {
  OnCall []PagerdutyUser
}

// blocking function that fetches on-call schedules
func (p *Pagerduty) populateState() (error) {
  var err error
  for _, rotationId := range pagerdutyRotationIds {
    err = p.fetchScheduleEntries(rotationId)
    if err != nil {
      return err
    }
  }
  return nil
} 

func (p *Pagerduty) fetchScheduleEntries(rotationId string) (error) {
  since := time.Now().UTC().Add(time.Duration(-10)*time.Minute).Format("2006-01-02T15:04:05Z")
  until := time.Now().UTC().Format("2006-01-02T15:04:05Z")
  path := fmt.Sprintf("/api/v1/schedules/%s/entries?overflow=true&since=%s&until=%s", rotationId, since, until)
  content, err := pagerdutyApiGet(path)
  if err != nil {
    return err
  }
  wrapper := ScheduleEntries{}
  err = json.Unmarshal(content, &wrapper)
  if err != nil {
    return err
  }
  for _, entry := range wrapper.Entries {
    p.OnCall = append(p.OnCall, entry.User)
  }
  return nil
}

func pagerdutyApiGet(path string) ([]byte, error) {
  var client = http.Client{}
  url := fmt.Sprintf("%s%s", pagerdutyApiUrl, path)
  authToken := fmt.Sprintf("Token token=%s", pagerdutyApiKey)

  req, err := http.NewRequest("GET", url, nil)
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Authorization", authToken) 
  resp, err := client.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }
  return body, nil
}
