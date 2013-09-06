package main

import (
  "fmt"
  "regexp"
  "net/http"
)

var client = &http.Client{}

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
      desc := fmt.Sprintf("%s/tickets/%d\r\n\r\n%s", zendesk_url, ticket_wrapper.Ticket.Id, ticket_wrapper.Ticket.Description)
      card, err = create_trello_card(ticket_wrapper.Ticket.Id,
                                     ticket_wrapper.Ticket.Status,
                                     desc,
                                     get_trello_list_id())
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

func fetch_content_from_apis() {
  done := make(chan error)

  go fetch_trello_board_lists(done)
  go fetch_trello_board_members(done)
  go fetch_trello_board_cards(done)
  go fetch_zendesk_users(done)
  go fetch_zendesk_view_tickets(done)

  // wait for API calls to finish
  for i := 0; i < 5; i++ {
    err := <-done
    if err != nil {
      fmt.Println(err)
      return
    }
  }
}

func match_trello_card_from_zendesk_id(id int64) *TrelloCard {
  for _, card := range trello_cards {
    match, _ := regexp.MatchString(fmt.Sprintf("Ticket #%d.*", id), card.Name)
    if match {
      return &card
    }
  }
  return nil
}

func maybe_update_title(card *TrelloCard, id int64, status string) {
  newName := fmt.Sprintf("Ticket #%d (%s)", id, status)
  if card.Name != newName {
    fmt.Println("update name")
    update_trello_card_name(card.Id, newName) 
  } 
}

func maybe_assign_trello_board_member(card *TrelloCard, assigneeId int64) {

}

func maybe_delete_stale_card(card *TrelloCard, ticket_id string) {
  for _, wrapper := range zendesk_view_tickets.Rows {
    if ticket_id == fmt.Sprintf("%d", wrapper.Ticket.Id) {
      return
    }
    fmt.Println("delete card for Ticket #", ticket_id)
    delete_trello_card(card.Id)
  }
}
