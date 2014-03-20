package main

import (
  "io"
  "strings"
  "fmt"
  "os"
  "net/url"
  "net/http"
  "io/ioutil"
  "encoding/json"
)

var trelloApiUrl string = "https://api.trello.com"
var trelloApiKey string = os.Getenv("TRELLO_API_KEY")
var trelloApiToken string = os.Getenv("TRELLO_API_TOKEN")
var trelloBoardId string = os.Getenv("TRELLO_BOARD_ID")
var trelloTargetList string = os.Getenv("TRELLO_LIST")

type TrelloList struct {
  Id string
  Name string
}

type TrelloMember struct {
  Id string
  Username string
}

type TrelloCard struct {
  Id string
  Name string
  IdMembers []string
}

type Trello struct {
  AvailableLists []TrelloList
  Members []TrelloMember
  Cards []TrelloCard
}

// blocking function that fetches members, lists and cards
// through multiple asynchronous REST API calls
func (t *Trello) populateState() {
  done := make(chan bool)

  t.asyncFetchMembers(done)
  t.asyncFetchLists(done)
  t.asyncFetchCards(done)

  // wait for API calls to finish
  for i := 0; i < 3; i++ { <-done }
} 

func (t *Trello) asyncFetchMembers(done chan bool) {
  go func(){
    path := fmt.Sprintf("/1/boards/%s/members?key=%s&token=%s", trelloBoardId, trelloApiKey, trelloApiToken)
    content := apiCall("GET", path, nil)
    err := json.Unmarshal(content, &t.Members)
    if err != nil { panic(err) }
    done <- true
  }()
}

func (t *Trello) asyncFetchLists(done chan bool) {
  go func(){
    path := fmt.Sprintf("/1/boards/%s/lists?key=%s&token=%s&fields=name", trelloBoardId, trelloApiKey, trelloApiToken)
    content := apiCall("GET", path, nil)
    err := json.Unmarshal(content, &t.AvailableLists)
    if err != nil { panic(err) }
    done <- true
  }()
}

func (t *Trello) asyncFetchCards(done chan bool) {
  go func(){
    path := fmt.Sprintf("/1/boards/%s/cards?key=%s&token=%s&lists=open&fields=name,idMembers", trelloBoardId, trelloApiKey, trelloApiToken)
    content := apiCall("GET", path, nil)
    err := json.Unmarshal(content, &t.Cards)
    if err != nil { panic(err) }
    done <- true
  }()
}

func (t *Trello) findMember(username string) (*TrelloMember) {
  for _, m := range t.Members {
    if m.Username == username {
      return &m
    }
  }
  return nil
}

func (t *Trello) findMemberById(id string) *TrelloMember {
  for _, m := range t.Members {
    if m.Id == id {
      return &m
    }
  }
  return nil
}

func (t *Trello) createCard(id int64, status string, desc string) (*TrelloCard) {
  name := fmt.Sprintf("Ticket #%d (%s)", id, status)
  card := &TrelloCard{Name: name}
  path := fmt.Sprintf("/1/cards?key=%s&token=%s", trelloApiKey, trelloApiToken)
  listId := t.targetList()
  params := url.Values{"idList": {listId}, "name": {name}, "desc": {desc}}
  content := apiCall("POST", path, params)
  err := json.Unmarshal(content, &card)
  if err != nil { panic(err) }
  t.Cards = append(t.Cards, *card)
  return card
}

// TODO: update card name in state
func (t *Trello) updateCardName(cardId string, cardName string) {
  path := fmt.Sprintf("/1/cards/%s/name?key=%s&token=%s", cardId, trelloApiKey, trelloApiToken)
  apiCall("PUT", path, url.Values{"value": {cardName}})
}

func (t *Trello) deleteCard(cardId string) {
  path := fmt.Sprintf("/1/cards/%s/closed?key=%s&token=%s", cardId, trelloApiKey, trelloApiToken)
  apiCall("PUT", path, url.Values{"value": {"true"}})
}

func (t *Trello) assignMember(cardId string, memberId string) {
  path := fmt.Sprintf("/1/cards/%s/idMembers?key=%s&token=%s", cardId, trelloApiKey, trelloApiToken)
  apiCall("PUT", path, url.Values{"value": {memberId}})
}

func (t *Trello) targetList() string {
  for _, list := range t.AvailableLists {
    if list.Name == trelloTargetList {
      return list.Id
    }
  }
  return ""
}

func apiCall(method string, path string, params url.Values) ([]byte) {
  var client = http.Client{}
  url := fmt.Sprintf("%s%s", trelloApiUrl, path)
  var form io.Reader = nil
  if params != nil {
    form = strings.NewReader(params.Encode())
  }
  req, err := http.NewRequest(method, url, form)
  if err != nil {
    panic(err)
  }
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    panic(err)
  }
  return body
}
