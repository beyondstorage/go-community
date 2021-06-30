package utils

import (
	"github.com/beyondstorage/go-community/env"
	"os"
	"testing"

	"github.com/beyondstorage/go-community/room"
)

func testProjects(t *testing.T) {
	_, err := Repos(os.Getenv(env.GithubOwner))
	if err != nil {
		t.Fatal(err)
	}
}

func testSync(t *testing.T) {
	rs, err := Repos(os.Getenv(env.GithubOwner))
	if err != nil {
		t.Fatal(err)
	}

	m, err := room.NewMatrix(
		os.Getenv(env.MatrixHomeServerURL),
		os.Getenv(env.MatrixHomeServer),
		os.Getenv(env.MatrixUserId),
		os.Getenv(env.MatrixToken))
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range rs {
		roomid, err := m.GetRoom(name)
		if err != nil {
			t.Fatalf("get room %s: %v", name, err)
		}
		if roomid == "" {
			roomid, err = m.CreateRoom(name)
			if err != nil {
				t.Fatalf("create room %s: %v", name, err)
			}
			t.Logf("room %s created, id %s", name, roomid)
		}

		err = m.PublicRoom(roomid)
		if err != nil {
			t.Fatalf("public room %s: %v", roomid, err)
		}
		t.Logf("room %s public", roomid)
	}
}
