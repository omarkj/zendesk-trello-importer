package main

import (
  "os"
  "time"
  "fmt"
  "net/url"
  "github.com/garyburd/redigo/redis"
)

var usersRedisUrl = os.Getenv("REDIS_URL")

type User struct {
  Email string
  TrelloUsername string
  LastAssignmentTime string
}

func findAllUsers() ([]User) {
  conn := getConn()
  defer conn.Close()
  keys, err := redis.Strings(conn.Do("KEYS", "*"))
  if err != nil {
    panic(err)
  }
  var users = []User{}
  for _, key := range keys {
    var f []string
    f, err = redis.Strings(conn.Do("HGETALL", key))
    if err != nil {
      panic(err)
    }
    user := User{Email:key}
    for i, v := range f {
      if v == "last_assignment" {
        user.LastAssignmentTime = f[i+1]
      }
      if v == "trello_username" {
        user.TrelloUsername = f[i+1]
      }
    }
    users = append(users, user)
  } 
  return users
}

func findUserByEmail(email string) (*User) {
  users := findAllUsers()
  for _, user := range users {
    if user.Email == email {
      return &user
    }
  }
  return nil
}

func writeLastAssignment(user *User) {
  conn := getConn()
  defer conn.Close()
  t := time.Now().UTC().Unix()
  _, err := conn.Do("HSET", user.Email, "last_assignment", fmt.Sprintf("%d", t))
  if err != nil { panic(err) }
}

func getConn() (redis.Conn) {
  var conn redis.Conn
  parsedUrl, err := url.Parse(usersRedisUrl)
  conn, err = redis.Dial("tcp", parsedUrl.Host)
  if err != nil { panic(err) }
  if parsedUrl.User != nil {
    p,_ := parsedUrl.User.Password()
    _, err = conn.Do("AUTH", p)
    if err != nil { panic(err) }
  }
  return conn
}
