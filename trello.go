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
func (t *Trello) populateState() (error) {
  done := make(chan error)

  t.asyncFetchMembers(done)
  t.asyncFetchLists(done)
  t.asyncFetchCards(done)

  // wait for API calls to finish
  for i := 0; i < 3; i++ {
    err := <-done
    if err != nil {
      return err
    }
  }
  return nil
} 

func (t *Trello) asyncFetchMembers(done chan error) {
  go func(){
    path := fmt.Sprintf("/1/boards/%s/members?key=%s&token=%s", trelloBoardId, trelloApiKey, trelloApiToken)
    content, err := apiCall("GET", path, nil)
    if err != nil {
      done <- err
      return
    }
    err = json.Unmarshal(content, &t.Members)
    done <- err
  }()
}

func (t *Trello) asyncFetchLists(done chan error) {
  go func(){
    path := fmt.Sprintf("/1/boards/%s/lists?key=%s&token=%s&fields=name", trelloBoardId, trelloApiKey, trelloApiToken)
    content, err := apiCall("GET", path, nil)
    if err != nil {
      done <- err
      return
    }
    err = json.Unmarshal(content, &t.AvailableLists)
    done <- err
  }()
}

func (t *Trello) asyncFetchCards(done chan error) {
  go func(){
    path := fmt.Sprintf("/1/boards/%s/cards?key=%s&token=%s&lists=open&fields=name,idMembers", trelloBoardId, trelloApiKey, trelloApiToken)
    content, err := apiCall("GET", path, nil)
    if err != nil {
      done <- err
      return
    }
    err = json.Unmarshal(content, &t.Cards)
    done <- err
  }()
}

func (t *Trello) createCard(id int64, status string, desc string) (*TrelloCard, error) {
  name := fmt.Sprintf("Ticket #%d (%s)", id, status)
  card := &TrelloCard{Name: name}
  path := fmt.Sprintf("/1/cards?key=%s&token=%s", trelloApiKey, trelloApiToken)
  listId := t.targetList()
  params := url.Values{"idList": {listId}, "name": {name}, "desc": {desc}}
  content, err := apiCall("POST", path, params)
  if err != nil {
    return nil, err
  }
  err = json.Unmarshal(content, &card)
  if err != nil {
    return nil, err
  }
  t.Cards = append(t.Cards, *card)
  return card, nil
}

// TODO: update card name in state
func (t *Trello) updateCardName(cardId string, cardName string) (error) {
  path := fmt.Sprintf("/1/cards/%s/name?key=%s&token=%s", cardId, trelloApiKey, trelloApiToken)
  _, err := apiCall("PUT", path, url.Values{"value": {cardName}})
  return err
}

func (t *Trello) deleteCard(cardId string) (error) {
  path := fmt.Sprintf("/1/cards/%s/closed?key=%s&token=%s", cardId, trelloApiKey, trelloApiToken)
  _, err := apiCall("PUT", path, url.Values{"value": {"true"}})
  return err
}

func (t *Trello) targetList() string {
  for _, list := range t.AvailableLists {
    if list.Name == trelloTargetList {
      return list.Id
    }
  }
  return ""
}

func apiCall(method string, path string, params url.Values) ([]byte, error) {
  var client = http.Client{}
  url := fmt.Sprintf("%s%s", trelloApiUrl, path)
  var form io.Reader = nil
  if params != nil {
    form = strings.NewReader(params.Encode())
  }
  req, err := http.NewRequest(method, url, form)
  if err != nil {
    return nil, err
  }
  req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
