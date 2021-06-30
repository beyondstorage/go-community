package room

import (
	"errors"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Matrix struct {
	hs string

	client *mautrix.Client
}

func NewMatrix(homeserverURL, homeserver, userId, token string) (m *Matrix, err error) {
	client, err := mautrix.NewClient(homeserverURL, id.NewUserID(userId, homeserver), token)
	if err != nil {
		return
	}

	m = &Matrix{
		client: client,
		hs:     homeserver,
	}
	return
}

func (m *Matrix) GetRoom(name string) (roomid string, err error) {
	resp, err := m.client.ResolveAlias(id.NewRoomAlias(name, m.hs))
	if err == nil {
		return resp.RoomID.String(), nil
	}
	var e mautrix.HTTPError
	// If error is not room not found, we should return directly.
	if errors.As(err, &e) && e.IsStatus(404) {
		return "", nil
	}
	return "", err
}

func (m *Matrix) CreateRoom(name string) (roomid string, err error) {
	// Room is not found we should create it.
	resp, err := m.client.CreateRoom(&mautrix.ReqCreateRoom{
		Visibility: "public",
		// The desired room alias local part.
		// If this is included, a room alias will be created and mapped to the newly created room.
		// The alias will belong on the same homeserver which created the room.
		// For example, if this was set to "foo" and sent to the homeserver "example.com" the complete room alias would be #foo:example.com.
		RoomAliasName: name,
		Name:          name,
		Topic:         "",
	})
	if err != nil {
		return
	}
	return resp.RoomID.String(), nil
}

func (m *Matrix) PublicRoom(roomid string) (err error) {
	_, err = m.client.SendStateEvent(
		id.RoomID(roomid),
		event.StateHistoryVisibility,
		"",
		event.HistoryVisibilityEventContent{
			HistoryVisibility: event.HistoryVisibilityWorldReadable,
		},
	)
	if err != nil {
		return err
	}

	_, err = m.client.SendStateEvent(
		id.RoomID(roomid),
		event.StateJoinRules,
		"",
		event.JoinRulesEventContent{
			JoinRule: event.JoinRulePublic,
		},
	)
	if err != nil {
		return err
	}

	_, err = m.client.SendStateEvent(
		id.RoomID(roomid),
		event.StateGuestAccess,
		"",
		event.GuestAccessEventContent{
			GuestAccess: event.GuestAccessCanJoin,
		},
	)
	if err != nil {
		return err
	}
	return nil
}
