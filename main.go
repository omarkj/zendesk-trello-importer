package main

import (
  "fmt"
  "regexp"
)

func main() {
  fmt.Println("fetching content...")

  // instantiate Trello and Zendesk objects
  trello := Trello{}
  zendesk := Zendesk{}

  err := trello.populateState()
  if (err!=nil) {
    fmt.Println("Failed fetching Trello state:", err)
  }

  err = zendesk.populateState()
  if (err!=nil) {
    fmt.Println("Failed fetching Zendesk state:", err)
  }

  fmt.Println("syncing tickets...")

  // loop through zendesk tickets
  for _, ticket := range zendesk.Tickets {
    card := findCardForTicket(&trello, ticket.Id) 
    if card != nil {
      fmt.Println(card.Name, "exists")
    } else {
      fmt.Println("creating trello card for ticket #", ticket.Id)
      desc := zendesk.formatCardDesc(ticket.Id, ticket.Description)
      err = trello.createCard(ticket.Id, ticket.Status, desc)
      if err != nil {
        fmt.Println("Failed to create Trello card:", err)
        return
      }
    }
    maybeUpdateCardTitle(&trello, card, ticket.Id, ticket.Status)
    maybeAssignCardOwner(card, ticket.Assignee_Id)
  }

  // loop through trello cards
  re := regexp.MustCompile("Ticket #(\\d+) .*")
  for _, card := range trello.Cards {
    matches := re.FindStringSubmatch(card.Name)
    if len(matches) == 2 {
      maybeDeleteStaleCard(&trello, &zendesk, card.Id, matches[1])
    }
  }
}

func findCardForTicket(trello *Trello, id int64) *TrelloCard {
  for _, card := range trello.Cards {
    match, _ := regexp.MatchString(fmt.Sprintf("Ticket #%d.*", id), card.Name)
    if match {
      return &card
    }
  }
  return nil
}

func maybeUpdateCardTitle(trello *Trello, card *TrelloCard, id int64, status string) {
  newName := fmt.Sprintf("Ticket #%d (%s)", id, status)
  if card.Name != newName {
    fmt.Println("update name")
    trello.updateCardName(card.Id, newName) 
  }
}

func maybeAssignCardOwner(card *TrelloCard, assigneeId int64) {

}

func maybeDeleteStaleCard(trello *Trello, zendesk *Zendesk, cardId string, ticketId string) {
  for _, ticket := range zendesk.Tickets {
    if ticketId == fmt.Sprintf("%d", ticket.Id) {
      return
    }
  }
  fmt.Println("delete card for Ticket #", ticketId)
  trello.deleteCard(cardId)
}
