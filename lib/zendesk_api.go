package lib

import (
  "os"
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
)

// curl -v -u $ZENDESK_USERNAME/token:$ZENDESK_TOKEN $ZENDESK_URL/api/v2/groups/$ZENDESK_GROUP_ID/users.json
// curl -v -u $ZENDESK_USERNAME/token:$ZENDESK_TOKEN $ZENDESK_URL/api/v2/views/$ZENDESK_VIEW/execute.json
// curl -v -u $ZENDESK_USERNAME/token:$ZENDESK_TOKEN $ZENDESK_URL/api/v2/groups/$ZENDESK_GROUP_ID/users.json

type ZendeskUser struct {
  Id int64
  Email string
}

type ZendeskUsers struct {
  Users []ZendeskUser
}

type ZendeskTicket struct {
  Id int64
  Status string
  Url string
  Description string
}

type ZendeskTicketWrapper struct {
  Assignee_Id int64
  Ticket ZendeskTicket
}

type ZendeskView struct {
  Rows []ZendeskTicketWrapper
}

var zendesk_token string = os.Getenv("ZENDESK_TOKEN")
var zendesk_url string = os.Getenv("ZENDESK_URL")
var zendesk_username = os.Getenv("ZENDESK_USERNAME")
var zendesk_group_id = os.Getenv("ZENDESK_GROUP_ID")
var zendesk_view = os.Getenv("ZENDESK_VIEW")
var zendesk_dummy_acct = os.Getenv("ZENDESK_DUMMY_ACCT")


func Fetch_zendesk_users(client *http.Client, zendesk_users *ZendeskUsers, done chan error) {
  path := fmt.Sprintf("%s/api/v2/groups/%s/users.json", zendesk_url, zendesk_group_id)
  content, err := zendesk_api_get(client, path)
  if err != nil {
    done <- err
    return
  }
  err = json.Unmarshal(content, &zendesk_users)
  done <- err
}

func Fetch_zendesk_view_tickets(client *http.Client, zendesk_view_tickets *ZendeskView, done chan error) {
  path := fmt.Sprintf("%s/api/v2/views/%s/execute.json", zendesk_url, zendesk_view)
  content, err := zendesk_api_get(client, path)
  if err != nil {
    done <- err
    return
  }
  err = json.Unmarshal(content, &zendesk_view_tickets)
  done <- err
}

func Format_zendesk_desc(id int64, desc string) string {
    return fmt.Sprintf("%s/tickets/%d\r\n\r\n%s", zendesk_url, id, desc)
}

func zendesk_api_get(client *http.Client, path string) ([]byte, error) {
  req, err := http.NewRequest("GET", path, nil)
  req.SetBasicAuth(fmt.Sprintf("%s/token", zendesk_username), zendesk_token)
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
