package main

import (
  "io"
  "strings"
  "os"
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
)

var zendeskToken string = os.Getenv("ZENDESK_TOKEN")
var zendeskApiUrl string = os.Getenv("ZENDESK_URL")
var zendeskUsername = os.Getenv("ZENDESK_USERNAME")
var zendeskGroupId = os.Getenv("ZENDESK_GROUP_ID")
var zendeskView = os.Getenv("ZENDESK_VIEW")
var zendeskDummyAcct = os.Getenv("ZENDESK_DUMMY_ACCT")

type ZendeskUser struct {
  Id int64
  Email string
}

type ZendeskUsers struct {
  Users []ZendeskUser
}

type ZendeskTicket struct {
  Id int64
  Assignee_Id int64
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

type Zendesk struct {
  Users []ZendeskUser
  Tickets []ZendeskTicket 
}

type ZendeskUpdateResponse struct {
  Ticket interface{}
}

type ZendeskAssigneeMsg struct {
  Assignee_Id int64
}

type ZendeskUpdateMsg struct {
  Ticket ZendeskAssigneeMsg
}

// blocking function that fetches users and tickets
// through asynchronous REST API calls
func (z *Zendesk) populateState() {
  done := make(chan bool)

  z.asyncFetchUsers(done)
  z.asyncFetchTickets(done)

  // wait for API calls to finish
  for i := 0; i < 2; i++ { <-done }
} 

func (z *Zendesk) asyncFetchUsers(done chan bool) {
  go func(){
    path := fmt.Sprintf("%s/api/v2/groups/%s/users.json", zendeskApiUrl, zendeskGroupId)
    content := zendeskApiGet(path)
    wrapper := ZendeskUsers{}
    err := json.Unmarshal(content, &wrapper)
    if err != nil { panic(err) }
    for _, user := range wrapper.Users {
      z.Users = append(z.Users, user)
    }
    done <- true
  }()
}

func (z *Zendesk) asyncFetchTickets(done chan bool) {
  go func(){
    path := fmt.Sprintf("%s/api/v2/views/%s/execute.json", zendeskApiUrl, zendeskView)
    content := zendeskApiGet(path)
    wrapper := ZendeskView{}
    err := json.Unmarshal(content, &wrapper)
    if err != nil { panic(err) }
    for _, ticketWrapper := range wrapper.Rows {
      ticketWrapper.Ticket.Assignee_Id = ticketWrapper.Assignee_Id
      z.Tickets = append(z.Tickets, ticketWrapper.Ticket)
    }
    done <- true
  }()
}

func (z *Zendesk) formatCardDesc(id int64, desc string) string {
  return fmt.Sprintf("%s/tickets/%d\r\n\r\n%s", zendeskApiUrl, id, desc)
}

func (z *Zendesk) findUser(id int64) *ZendeskUser {
  for _, user := range z.Users {
    if user.Id == id {
      return &user
    }
  }
  return nil
}

func (z *Zendesk) updateTicketOwner(id int64) {
  path := fmt.Sprintf("/api/v2/tickets/%d.json", id)
  b, err := json.Marshal(ZendeskUpdateMsg{ZendeskAssigneeMsg{id}})
  content := zendeskApiCall("PUT", path, b)
  resp := &ZendeskUpdateResponse{}
  err = json.Unmarshal(content, &resp)
  if err != nil { panic(err) }
}

func zendeskApiGet(path string) ([]byte) {
  var client = http.Client{}
  req, err := http.NewRequest("GET", path, nil)
  req.SetBasicAuth(fmt.Sprintf("%s/token", zendeskUsername), zendeskToken)
  resp, err := client.Do(req)
  if err != nil { panic(err) }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil { panic(err) }
  return body
}

func zendeskApiCall(method string, path string, b []byte) ([]byte) {
  var client = http.Client{}
  url := fmt.Sprintf("%s%s", zendeskApiUrl, path)
  var form io.Reader = nil
  if b != nil {
    form = strings.NewReader(string(b))
  }
  req, err := http.NewRequest(method, url, form)
  if err != nil {
    panic(err)
  }
  req.SetBasicAuth(fmt.Sprintf("%s/token", zendeskUsername), zendeskToken)
  req.Header.Set("Content-Type", "Content-Type: application/json")
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
