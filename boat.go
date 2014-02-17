/*
 * GoBoat, the Boat that goes.
 */

package main

import (
  "fmt"
  "flag"
  "strings"

  "crypto/tls"
  "github.com/thoj/go-ircevent"
)

func run_network(net string, cfg *NetworkConfig, db_chan chan *LoggerEvent, quit_chan chan string) {
  // initialize: nick, username
  ircobj := irc.IRC(cfg.Nick, cfg.Username)

  // ssl
  ircobj.UseTLS = cfg.UseSSL
  ircobj.TLSConfig = &tls.Config{InsecureSkipVerify: true}

  // spit out everything to stdout
  ircobj.VerboseCallbackHandler = true

  // go go go
  ircobj.Connect(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))

  // connected to network. identify with nickserv, and maybe join some
  // channels.
  ircobj.AddCallback("001", func(event *irc.Event) {
    // identify with NickServ
    if cfg.IdentPass != "" {
      ircobj.Privmsgf("nickserv", "identify %v %v", cfg.IdentNick, cfg.IdentPass)
    }

    // if we're not waiting for a vhost to change, just join channels
    // now
    if !cfg.WaitVHost {
      // join channels
      for _, channel := range cfg.Channel {
        ircobj.Join(channel)
      }
    }

    // identify as an oper on connect
    if cfg.OperPass != "" {
      ircobj.SendRawf("OPER %v %v", cfg.OperNick, cfg.OperPass)
    }
  })

  // hostname changed
  ircobj.AddCallback("396", func(event *irc.Event) {
    // this should probably be toggled to fire only once and never
    // again somehow.
    if cfg.WaitVHost {
      for _, channel := range cfg.Channel {
        ircobj.Join(channel)
      }
    }
  })

  // try to fix a netsplit
  ircobj.AddCallback("NOTICE", func(event *irc.Event) {
    if cfg.OperPass == "" { return }
    if cfg.AutoSconn == false { return }

    if event.Arguments[0] != "*" { return }
    if event.Source != cfg.Host { return }
    if strings.Index(event.Raw, "Netsplit") < 0 { return }

    // try to get the split host
    start := strings.Index(event.Raw, "<->") + 4
    end   := strings.Index(event.Raw[start:], " ")
    host  := event.Raw[start:start+end]

    ircobj.SendRawf("CONNECT %v", host)
  })

  // log stuff
  ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {
    db_chan <- &LoggerEvent{Event:event, Network:net}
  })

  // hi responder
  ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {
    target := event.Arguments[0]
    if target == cfg.Nick {
      target = event.Nick
    }

    if m := event.Message(); m == "hi" {
      ircobj.Privmsgf(target, "hi, %v", event.Nick)
    }
  })

  // search logs
  ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {
    // ignore non-pms
    if event.Arguments[0] != cfg.Nick { return }

  })

  // looooooooper
  ircobj.Loop()

  // tell someone we're done
  quit_chan <- net
}

func main() {
  // load config options
  config_file := flag.String("config", "config.ini", "Location of config file")
  flag.Parse()
  config := load_config(*config_file)

  // set up channel for logging
  db_chan, db_collector, db_cleanup := db_logger(config.Logger.Driver, config.Logger.Source)
  defer db_cleanup()
  go db_collector()

  // fire up a bunch of channels for each network
  quit_chans := make(map[string](chan string))
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

  // all the channels have quit. nothing to do :|
}
