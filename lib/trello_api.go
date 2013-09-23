package lib

import (
  "os"
  "io"
  "strings"
  "fmt"
  "net/url"
  "net/http"
  "io/ioutil"
  "encoding/json"
)

// https://trello.com/1/authorize?key=$TRELLO_API_KEY&name=Go%20Zendesk%20Importer&expiration=never&response_type=token&scope=read,write
// curl -v "https://api.trello.com/1/boards/$TRELLO_BOARD_ID/members?key=$TRELLO_API_KEY&token=$TRELLO_API_TOKEN"
// curl -v "https://api.trello.com/1/boards/$TRELLO_BOARD_ID/cards?key=$TRELLO_API_KEY&token=$TRELLO_API_TOKEN&lists=open&fields=name,idMembers"

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

var trello_api_host string = "https://api.trello.com"
var trello_api_key string = os.Getenv("TRELLO_API_KEY")
var trello_api_token string = os.Getenv("TRELLO_API_TOKEN")
var trello_board_id string = os.Getenv("TRELLO_BOARD_ID")
var trello_list string = os.Getenv("TRELLO_LIST")


func Get_trello_list_id(trello_lists *[]TrelloList) string {
  for _, list := range *trello_lists {
    if list.Name == trello_list {
      return list.Id
    }
  }
  return ""
}

func Fetch_trello_board_cards(client *http.Client, trello_cards *[]TrelloCard, done chan error) {
  path := fmt.Sprintf("/1/boards/%s/cards?key=%s&token=%s&lists=open&fields=name,idMembers", trello_board_id, trello_api_key, trello_api_token)
  content, err := trello_api_method(client, "GET", path, nil)
  if err != nil {
    done <- err
    return
  }
  err = json.Unmarshal(content, &trello_cards)
  done <- err
}

func Fetch_trello_board_members(client *http.Client, trello_members *[]TrelloMember, done chan error) {
  path := fmt.Sprintf("/1/boards/%s/members?key=%s&token=%s", trello_board_id, trello_api_key, trello_api_token)
  content, err := trello_api_method(client, "GET", path, nil)
  if err != nil {
    done <- err
    return
  }
  err = json.Unmarshal(content, &trello_members)
  done <- err
}

func Fetch_trello_board_lists(client *http.Client, trello_lists *[]TrelloList, done chan error) {
  path := fmt.Sprintf("/1/boards/%s/lists?key=%s&token=%s&fields=name", trello_board_id, trello_api_key, trello_api_token)
  content, err := trello_api_method(client, "GET", path, nil)
  if err != nil {
    done <- err
    return
  }
  err = json.Unmarshal(content, &trello_lists)
  done <- err
}

func Create_trello_card(client *http.Client, id int64, status string, desc string, listId string) (*TrelloCard, error) {
  name := fmt.Sprintf("Ticket #%d (%s)", id, status)
  card := TrelloCard{Name: name}
  path := fmt.Sprintf("/1/cards?key=%s&token=%s", trello_api_key, trello_api_token)
  params := url.Values{"idList": {listId}, "name": {name}, "desc": {desc}}
  content, err := trello_api_method(client, "POST", path, params)
  if err != nil {
    return nil, err
  }
  err = json.Unmarshal(content, &card)
  if err != nil {
    return nil, err
  }
  return &card, nil
}

func Update_trello_card_name(client *http.Client, cardId string, cardName string) (error) {
  path := fmt.Sprintf("/1/cards/%s/name?key=%s&token=%s", cardId, trello_api_key, trello_api_token)
  _, err := trello_api_method(client, "PUT", path, url.Values{"value": {cardName}})
  return err
}

func Delete_trello_card(client *http.Client, cardId string) (error) {
  path := fmt.Sprintf("/1/cards/%s/closed?key=%s&token=%s", cardId, trello_api_key, trello_api_token)
  _, err := trello_api_method(client, "PUT", path, url.Values{"value": {"true"}})
  return err
}

func trello_api_method(client *http.Client, method string, path string, params url.Values) ([]byte, error) {
  url := fmt.Sprintf("%s%s", trello_api_host, path)
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
