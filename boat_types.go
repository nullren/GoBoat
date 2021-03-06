/*
 * GoBoat, the Boat that goes.
 */

package main

import (
  "github.com/nullren/go-ircevent"
)

type NetworkConfig struct {
  Host       string
  Port       int
  UseSSL     bool
  UseSocks   bool
  SocksHost  string
  SocksPort  int
  Nick       string
  Username   string
  Realname   string
  Channel    []string
  IdentPass  string
  IdentNick  string
  WaitVHost  bool
  OperPass   string
  OperNick   string
  AutoSconn  bool
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
