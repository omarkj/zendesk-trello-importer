package main

import (
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

// blocking function that fetches users and tickets
// through asynchronous REST API calls
func (z *Zendesk) populateState() (error) {
  done := make(chan error)

  z.asyncFetchUsers(done)
  z.asyncFetchTickets(done)

  // wait for API calls to finish
  for i := 0; i < 2; i++ {
    err := <-done
    if err != nil {
      return err
    }
  }
  return nil
} 

func (z *Zendesk) asyncFetchUsers(done chan error) {
  go func(){
    path := fmt.Sprintf("%s/api/v2/groups/%s/users.json", zendeskApiUrl, zendeskGroupId)
    content, err := zendeskApiGet(path)
    if err != nil {
      done <- err
      return
    }
    wrapper := ZendeskUsers{}
    err = json.Unmarshal(content, &wrapper)
    if err == nil {
      for _, user := range wrapper.Users {
        z.Users = append(z.Users, user)
      }
    }
    done <- err
  }()
}

func (z *Zendesk) asyncFetchTickets(done chan error) {
  go func(){
    path := fmt.Sprintf("%s/api/v2/views/%s/execute.json", zendeskApiUrl, zendeskView)
    content, err := zendeskApiGet(path)
    if err != nil {
      done <- err
      return
    }
    wrapper := ZendeskView{}
    err = json.Unmarshal(content, &wrapper)
    if err == nil {
      for _, ticketWrapper := range wrapper.Rows {
        ticketWrapper.Ticket.Assignee_Id = ticketWrapper.Assignee_Id
        z.Tickets = append(z.Tickets, ticketWrapper.Ticket)
      }
    }
    done <- err
  }()
}

func (z *Zendesk) formatCardDesc(id int64, desc string) string {
  return fmt.Sprintf("%s/tickets/%d\r\n\r\n%s", zendeskApiUrl, id, desc)
}

func zendeskApiGet(path string) ([]byte, error) {
  var client = http.Client{}
  req, err := http.NewRequest("GET", path, nil)
  req.SetBasicAuth(fmt.Sprintf("%s/token", zendeskUsername), zendeskToken)
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
