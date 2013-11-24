package main

import (
  "fmt"
  "regexp"
)

func runAsImporter() {
  fmt.Println("fetching content...")

  var err error
  trello := Trello{}
  zendesk := Zendesk{}
  pagerduty := Pagerduty{}

  err = trello.populateState()
  if (err!=nil) {
    fmt.Println("failed fetching Trello state:", err)
    return
  }

  err = zendesk.populateState()
  if (err!=nil) {
    fmt.Println("Failed fetching Zendesk state:", err)
    return
  }

  err = pagerduty.populateState()
  if (err!=nil) {
    fmt.Println("Failed fetching PagerDuty state:", err)
    return
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
      card, err = trello.createCard(ticket.Id, ticket.Status, desc)
      if err != nil {
        fmt.Println("Failed to create Trello card:", err)
        return
      }
    }
    maybeUpdateCardTitle(&trello, card, ticket.Id, ticket.Status)
    maybeAssignCardOwner(&trello, &zendesk, card, ticket.Assignee_Id)
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

func maybeAssignCardOwner(trello *Trello, zendesk * Zendesk, card *TrelloCard, assigneeId int64) {
  zendeskUser := zendesk.findUser(assigneeId)
  if zendeskUser != nil {
    // lookup user mapping in redis
    user, err := findUserByEmail(zendeskUser.Email)
    if err != nil {
      return
    }
    trelloMember := trello.findMember(user.TrelloUsername)
    if trelloMember != nil {
      // assign trello member to card 
      err = trello.assignMember(card.Id, trelloMember.Id)
      if err != nil {
        return
      }
    }
  }
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
