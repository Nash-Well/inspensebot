package database

import (
	"encoding/json"
	tele "gopkg.in/telebot.v3"
)

type UserCache struct {
	CategoryPage    int
	CategoryMessage int
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

func (user User) CategoryMessage() *tele.Message {
	return &tele.Message{
		ID:   user.GetCache().CategoryMessage,
		Chat: &tele.Chat{ID: user.ID},
	}
}