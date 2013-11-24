package main

import (
  "os"
  "time"
  "fmt"
  "strconv"
  "net/url"
  "github.com/garyburd/redigo/redis"
)

var usersRedisUrl = os.Getenv("REDIS_URL")

type User struct {
  Email string
  TrelloUsername string
  LastAssignmentTime string
}

func findAllUsers() ([]User, error) {
  conn, err := getConn()
  if err != nil {
    return nil, err
  }
  defer conn.Close()
  var keys []string
  keys, err = redis.Strings(conn.Do("KEYS", "*"))
  if err != nil {
    return nil, err
  }
  var users = []User{}
  for _, key := range keys {
    var f []string
    f, err = redis.Strings(conn.Do("HGETALL", key))
    if err != nil {
      return nil, err
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
  return users, nil
}

func findUserByEmail(email string) (*User, error) {
  users, err := findAllUsers()
  if err != nil {
    return nil, err
  }
  for _, user := range users {
    if user.Email == email {
      return &user, nil
    }
  }
  return nil, nil
}

func writeLastAssignment(user string) (error) {
  conn, err := getConn()
  if err != nil {
    return err
  }
  defer conn.Close()
  t := time.Now().UTC().Unix()
  _, err = conn.Do("SET", user, fmt.Sprintf("%d", t))
  return err
}

func getLastAssignment(user string) (int64, error) {
  conn, err := getConn()
  if err != nil {
    return 0, err
  }
  defer conn.Close()
  var t string
  t, err = redis.String(conn.Do("GET", user))
  if err != nil {
    return 0, err
  }
  var i int64
  i, err = strconv.ParseInt(t, 10, 64)
  if err != nil {
    return 0, err
  }
  return i, err
}

func getConn() (redis.Conn, error) {
  var conn redis.Conn
  parsedUrl, err := url.Parse(usersRedisUrl)
  conn, err = redis.Dial("tcp", parsedUrl.Host)
  if err != nil {
    return nil, err
  }

  if parsedUrl.User != nil {
    p,_ := parsedUrl.User.Password()
    _, err = conn.Do("AUTH", p)
    if err != nil {
      return nil, err
    }
  }
  return conn, nil
}
