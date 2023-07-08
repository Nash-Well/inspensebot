package database

import (
	"encoding/json"
	tele "gopkg.in/telebot.v3"
)

type UserCache struct {
	ListPage          int
	ListMessageID     int
	ActionsMessageID  int
	CategoryPage      int
	CategoryMessageID int
	MediaMessageID    int
	ShareMessageID    int
}

func (user *User) Exists(key string) bool {
	m := make(map[string]any)
	json.Unmarshal(user.Cache, &m)
	_, ok := m[key]
	return ok
}

func (user *User) UpdateCache(key string, value any) {
	m := make(map[string]any)
	json.Unmarshal(user.Cache, &m)
	m[key] = value
	data, _ := json.Marshal(m)
	user.Cache = data
}

func (user *User) DeleteFromCache(key string) {
	m := make(map[string]any)
	json.Unmarshal(user.Cache, &m)
	delete(m, key)
	data, _ := json.Marshal(m)
	user.Cache = data
}

func (user *User) GetCache() (cache UserCache) {
	json.Unmarshal(user.Cache, &cache)
	return
}

func (user User) ListMessage() *tele.Message {
	return &tele.Message{
		ID:   user.GetCache().ListMessageID,
		Chat: &tele.Chat{ID: user.ID},
	}
}

func (user User) ListActionsMessage() *tele.Message {
	return &tele.Message{
		ID:   user.GetCache().ActionsMessageID,
		Chat: &tele.Chat{ID: user.ID},
	}
}

func (user User) CategoryMessage() *tele.Message {
	return &tele.Message{
		ID:   user.GetCache().CategoryMessageID,
		Chat: &tele.Chat{ID: user.ID},
	}
}

func (user User) MediaMessage() *tele.Message {
	return &tele.Message{
		ID:   user.GetCache().MediaMessageID,
		Chat: &tele.Chat{ID: user.ID},
	}
}

func (user User) ShareMessage() *tele.Message {
	return &tele.Message{
		ID:   user.GetCache().ShareMessageID,
		Chat: &tele.Chat{ID: user.ID},
	}
}
