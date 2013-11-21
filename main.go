package main

import (
  "fmt"
  "regexp"
  "net/http"
  "github.com/JacobVorreuter/zendesk-trello-importer/lib"
)

// HTTP client
var client = &http.Client{}

// global trello objects
var trello_lists []lib.TrelloList
var trello_members []lib.TrelloMember
var trello_cards []lib.TrelloCard

// global zendesk objects
var zendesk_users lib.ZendeskUsers
var zendesk_view_tickets lib.ZendeskView

func main() {
  fmt.Println("refreshing data...")
  fetch_content_from_apis()

  fmt.Println("syncing tickets...") 

  // loop through zendesk tickets
  for _, ticket_wrapper := range zendesk_view_tickets.Rows {
    card := match_trello_card_from_zendesk_id(ticket_wrapper.Ticket.Id)
    if card != nil {
      fmt.Println(card.Name, "exists")
    } else {
      fmt.Println("creating trello card for ticket #", ticket_wrapper.Ticket.Id)
      var err error
      desc := lib.Format_zendesk_desc(ticket_wrapper.Ticket.Id, ticket_wrapper.Ticket.Description)
      card, err = lib.Create_trello_card(client,
                                         ticket_wrapper.Ticket.Id,
                                         ticket_wrapper.Ticket.Status,
                                         desc,
                                         lib.Get_trello_list_id(&trello_lists))
      if err != nil {
        fmt.Println("\nerr: ", err)
        return
      }
    }
    maybe_update_title(card, ticket_wrapper.Ticket.Id, ticket_wrapper.Ticket.Status)
    maybe_assign_trello_board_member(card, ticket_wrapper.Assignee_Id)
  }

  // loop through trello cards
  re := regexp.MustCompile("Ticket #(\\d+) .*")
  for _, card := range trello_cards {
    matches := re.FindStringSubmatch(card.Name)
    if len(matches) == 2 {
      maybe_delete_stale_card(&card, matches[1])
    }
  }
}

// asynchronously populate global objects
// through trello and zendesk REST APIs
func fetch_content_from_apis() {
  done := make(chan error)

  go lib.Fetch_trello_board_lists(client, &trello_lists, done)
  go lib.Fetch_trello_board_members(client, &trello_members, done)
  go lib.Fetch_trello_board_cards(client, &trello_cards, done)
  go lib.Fetch_zendesk_users(client, &zendesk_users, done)
  go lib.Fetch_zendesk_view_tickets(client, &zendesk_view_tickets, done)

  // wait for API calls to finish
  for i := 0; i < 5; i++ {
    err := <-done
    if err != nil {
      fmt.Println(err)
      return
    }
  }
}

func match_trello_card_from_zendesk_id(id int64) *lib.TrelloCard {
  for _, card := range trello_cards {
    match, _ := regexp.MatchString(fmt.Sprintf("Ticket #%d.*", id), card.Name)
    if match {
      return &card
    }
  }
  return nil
}

func maybe_update_title(card *lib.TrelloCard, id int64, status string) {
  newName := fmt.Sprintf("Ticket #%d (%s)", id, status)
  if card.Name != newName {
    fmt.Println("update name")
    lib.Update_trello_card_name(client, card.Id, newName) 
  } 
}

func maybe_assign_trello_board_member(card *lib.TrelloCard, assigneeId int64) {

}

func maybe_delete_stale_card(card *lib.TrelloCard, ticket_id string) {
  for _, wrapper := range zendesk_view_tickets.Rows {
    if ticket_id == fmt.Sprintf("%d", wrapper.Ticket.Id) {
      return
    }
  }
  fmt.Println("delete card for Ticket #", ticket_id)
  lib.Delete_trello_card(client, card.Id)
}
