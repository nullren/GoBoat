/*
 * GoBoat, the Boat that goes.
 */

package main

import (
  "log"
  "fmt"
  "flag"
  "errors"

  "code.google.com/p/gcfg"

  "crypto/tls"
  "github.com/thoj/go-ircevent"

  "time"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

// types

type NetworkConfig struct {
  Host       string
  Port       int
  UseSSL     bool
  Nick       string
  Username   string
  Channel    []string
}

type LoggerConfig struct {
  Driver string
  Source string
}

type Config struct {
  Logger LoggerConfig
  General NetworkConfig
  Network map[string]*NetworkConfig
}

type LoggerEvent struct {
  Event *irc.Event
  Network string
}

// fail a bitch

func fail(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

// load and set defaults

func load_config(file string) Config {
  var cfg Config
  err := gcfg.ReadFileInto(&cfg, file)
  fail(err)

  //if strings.ToLower(cfg.Logger.Driver) != "sqlite3" { }

  // use some sane defaults
  for net,v := range cfg.Network {
    // must have a server to connect to
    if v.Host == "" {
      fail(errors.New(fmt.Sprintf("load_config: no host given in network %s", net)))
    }

    // must have a nickname
    if v.Nick == "" {
      if cfg.General.Nick == "" {
        fail(errors.New(fmt.Sprintf("load_config: no nickname given in network %s", net)))
      } else { v.Nick = cfg.General.Nick }
    }

    // defaults
    if v.Port     == 0  { v.Port     = 6667 }
    if v.Username == "" {
      if cfg.General.Username == "" {
        v.Username = "GoBoat"
      } else {
        v.Username = cfg.General.Username
      }
    }
  }

  return cfg
}

func db_logger(driver, source string) (chan *LoggerEvent, func(), func()) {
  // set up a channel
  c := make(chan *LoggerEvent)

  db, err := sql.Open(driver, source)
  fail(err)

  create_table := `
      create table if not exists log (date integer, network text, nick text, channel text, message text);
      `
  _, err = db.Exec(create_table)
  fail(err)

  insert_message := `
      insert into log values (?, ?, ?, ?, ?);
      `
  stmt, err := db.Prepare(insert_message)
  fail(err)

  // goroutine to put away message events
  collector := func() {
    for {
      le := <-c
      event, network := le.Event, le.Network
      _, err = stmt.Exec(time.Now().Unix(), network, event.Nick, event.Arguments[0], event.Message())
      fail(err)
    }
  }

  // figure out how to defer these to when an error occurs
  cleanup := func() {
    db.Close()
    stmt.Close()
  }

  return c, collector, cleanup
}

func run_network(net string, cfg *NetworkConfig, db_chan chan *LoggerEvent, quit_chan chan string) {
  // initialize: nick, username
  ircobj := irc.IRC(cfg.Nick, cfg.Username)

  // ssl
  ircobj.UseTLS = cfg.UseSSL
  ircobj.TLSConfig = &tls.Config{InsecureSkipVerify: true}

  // spit out everything to stdout
  //ircobj.VerboseCallbackHandler = true

  // go go go
  ircobj.Connect(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))

  // join channels
  ircobj.AddCallback("001", func(event *irc.Event) {
    for _, channel := range cfg.Channel {
      ircobj.Join(channel)
    }
  })

  // log stuff
  ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {
    db_chan <- &LoggerEvent{Event:event, Network:net}
  })

  // hi responder
  ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {
    if m := event.Message(); m == "hi" {
      ircobj.Privmsgf(event.Arguments[0], "hi, %v", event.Nick)
    }
  })

  // looooooooper
  ircobj.Loop()

  // tell someone we're done
  quit_chan <- net
}

func main() {
  config_file := flag.String("config", "config.ini", "Location of config file")

  flag.Parse()

  config := load_config(*config_file)

  // set up channel for logging
  db_chan, db_collector, db_cleanup := db_logger(config.Logger.Driver, config.Logger.Source)
  defer db_cleanup()
  go db_collector()

  // a bunch of channels for each network
  quit_chans := make(map[string](chan string))

  // fire them up
  for net, cfg := range config.Network {
    quit_chans[net] = make(chan string)
    go run_network(net, cfg, db_chan, quit_chans[net])
  }

  // listen to the quit_chans
  quit_chan := make(chan string)
  for net, _ := range config.Network {
    go func() { quit_chan <- <-quit_chans[net] }()
  }

  for _, _ = range config.Network {
    fmt.Printf("OMGOMG: %s DISCONNECTED!", <-quit_chan)
  }
}
