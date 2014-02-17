/*
 * GoBoat, the Boat that goes.
 */

package main

import (
  "time"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

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
