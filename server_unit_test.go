package docker

import (
	"github.com/dotcloud/docker/utils"
	"testing"
	"time"
)

func TestPools(t *testing.T) {
	srv := &Server{
		pullingPool: make(map[string]struct{}),
		pushingPool: make(map[string]struct{}),
	}

	err := srv.poolAdd("pull", "test1")
	if err != nil {
		t.Fatal(err)
	}
	err = srv.poolAdd("pull", "test2")
	if err != nil {
		t.Fatal(err)
	}
	err = srv.poolAdd("push", "test1")
	if err == nil || err.Error() != "pull test1 is already in progress" {
		t.Fatalf("Expected `pull test1 is already in progress`")
	}
	err = srv.poolAdd("pull", "test1")
	if err == nil || err.Error() != "pull test1 is already in progress" {
		t.Fatalf("Expected `pull test1 is already in progress`")
	}
	err = srv.poolAdd("wait", "test3")
	if err == nil || err.Error() != "Unknown pool type" {
		t.Fatalf("Expected `Unknown pool type`")
	}

	err = srv.poolRemove("pull", "test2")
	if err != nil {
		t.Fatal(err)
	}
	err = srv.poolRemove("pull", "test2")
	if err != nil {
		t.Fatal(err)
	}
	err = srv.poolRemove("pull", "test1")
	if err != nil {
		t.Fatal(err)
	}
	err = srv.poolRemove("push", "test1")
	if err != nil {
		t.Fatal(err)
	}
	err = srv.poolRemove("wait", "test3")
	if err == nil || err.Error() != "Unknown pool type" {
		t.Fatalf("Expected `Unknown pool type`")
	}
}

func TestLogEvent(t *testing.T) {
	srv := &Server{
		events:    make([]utils.JSONMessage, 0, 64),
		listeners: make(map[string]chan utils.JSONMessage),
	}

	srv.LogEvent("fakeaction", "fakeid", "fakeimage")

	listener := make(chan utils.JSONMessage)
	srv.Lock()
	srv.listeners["test"] = listener
	srv.Unlock()

	srv.LogEvent("fakeaction2", "fakeid", "fakeimage")

	if len(srv.events) != 2 {
		t.Fatalf("Expected 2 events, found %d", len(srv.events))
	}
	go func() {
		time.Sleep(200 * time.Millisecond)
		srv.LogEvent("fakeaction3", "fakeid", "fakeimage")
		time.Sleep(200 * time.Millisecond)
		srv.LogEvent("fakeaction4", "fakeid", "fakeimage")
	}()

	setTimeout(t, "Listening for events timed out", 2*time.Second, func() {
		for i := 2; i < 4; i++ {
			event := <-listener
			if event != srv.events[i] {
				t.Fatalf("Event received it different than expected")
			}
		}
	})
}

// FIXME: this is duplicated from integration/commands_test.go
func setTimeout(t *testing.T, msg string, d time.Duration, f func()) {
	c := make(chan bool)

	// Make sure we are not too long
	go func() {
		time.Sleep(d)
		c <- true
	}()
	go func() {
		f()
		c <- false
	}()
	if <-c && msg != "" {
		t.Fatal(msg)
	}
}
