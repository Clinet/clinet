package services

import (
	"errors"
	"io/ioutil"

	"github.com/JoshuaDoes/json"
)

var (
	errorInvalidKey     = errors.New("invalid key")
	errorInvalidExtra   = errors.New("invalid extra")
	errorInvalidChannel = errors.New("invalid channel")
	errorInvalidMessage = errors.New("invalid message")
	errorInvalidServer  = errors.New("invalid server")
	errorInvalidUser    = errors.New("invalid user")
)

//Storage is designed for stateless interactions, to set it and forget it until it's needed again
type Storage struct {
	Extras   map[string]*StorageObject `json:"extra,omitempty"`
	Channels map[string]*StorageObject `json:"channels,omitempty"`
	Messages map[string]*StorageObject `json:"messages,omitempty"`
	Servers  map[string]*StorageObject `json:"servers,omitempty"`
	Users    map[string]*StorageObject `json:"users,omitempty"`
}

func (s *Storage) LoadFrom(state string) error {
	if s == nil {
		return errors.New("unable to load state into nil storage")
	}

	stateJSON, err := ioutil.ReadFile("states/" + state + ".json")
	if err != nil {
		s = &Storage{
			Extras: make(map[string]*StorageObject),
			Channels: make(map[string]*StorageObject),
			Messages: make(map[string]*StorageObject),
			Servers: make(map[string]*StorageObject),
			Users: make(map[string]*StorageObject),
		}

		stateJSON, err = json.Marshal(s, true)
		if err != nil {
			return err
		}

		return ioutil.WriteFile("states/" + state + ".json", stateJSON, 0644)
	}

	return json.Unmarshal(stateJSON, s)
}

type StorageObject struct {
	Data map[string]interface{} `json:"data,omitempty"`
}
func (so *StorageObject) Get(key string) (interface{}, error) {
	if data, exists := so.Data[key]; exists {
		return data, nil
	}
	return nil, errorInvalidKey
}
func (so *StorageObject) Set(key string, val interface{}) {
	so.Data[key] = val
}
func (so *StorageObject) Del(key string) {
	delete(so.Data, key)
}

func (s *Storage) ExtraGet(extraID, key string) (interface{}, error) {
	if s.Extras == nil {
		s.Extras = make(map[string]*StorageObject)
	}
	if _, exists := s.Extras[extraID]; exists {
		return s.Extras[extraID].Get(key)
	}
	return nil, errorInvalidExtra
}
func (s *Storage) ExtraSet(extraID, key string, val interface{}) {
	if s.Extras == nil {
		s.Extras = make(map[string]*StorageObject)
	}
	if _, exists := s.Extras[extraID]; !exists {
		s.Extras[extraID] = &StorageObject{
			Data: make(map[string]interface{}),
		}
	}
	s.Extras[extraID].Set(key, val)
}
func (s *Storage) ExtraDel(extraID, key string) {
	if s.Extras == nil {
		s.Extras = make(map[string]*StorageObject)
	}
	if _, exists := s.Extras[extraID]; exists {
		s.Extras[extraID].Del(key)
	}
}

func (s *Storage) ChannelGet(channelID, key string) (interface{}, error) {
	if s.Channels == nil {
		s.Channels = make(map[string]*StorageObject)
	}
	if _, exists := s.Channels[channelID]; exists {
		return s.Channels[channelID].Get(key)
	}
	return nil, errorInvalidChannel
}
func (s *Storage) ChannelSet(channelID, key string, val interface{}) {
	if s.Channels == nil {
		s.Channels = make(map[string]*StorageObject)
	}
	if _, exists := s.Channels[channelID]; !exists {
		s.Channels[channelID] = &StorageObject{
			Data: make(map[string]interface{}),
		}
	}
	s.Channels[channelID].Set(key, val)
}
func (s *Storage) ChannelDel(channelID, key string) {
	if s.Channels == nil {
		s.Channels = make(map[string]*StorageObject)
	}
	if _, exists := s.Channels[channelID]; exists {
		s.Channels[channelID].Del(key)
	}
}

func (s *Storage) MessageGet(messageID, key string) (interface{}, error) {
	if s.Messages == nil {
		s.Messages = make(map[string]*StorageObject)
	}
	if _, exists := s.Messages[messageID]; exists {
		return s.Messages[messageID].Get(key)
	}
	return nil, errorInvalidMessage
}
func (s *Storage) MessageSet(messageID, key string, val interface{}) {
	if s.Messages == nil {
		s.Messages = make(map[string]*StorageObject)
	}
	if _, exists := s.Messages[messageID]; !exists {
		s.Messages[messageID] = &StorageObject{
			Data: make(map[string]interface{}),
		}
	}
	s.Messages[messageID].Set(key, val)
}
func (s *Storage) MessageDel(messageID, key string) {
	if s.Messages == nil {
		s.Messages = make(map[string]*StorageObject)
	}
	if _, exists := s.Messages[messageID]; exists {
		s.Messages[messageID].Del(key)
	}
}

func (s *Storage) ServerGet(serverID, key string) (interface{}, error) {
	if s.Servers == nil {
		s.Servers = make(map[string]*StorageObject)
	}
	if _, exists := s.Servers[serverID]; exists {
		return s.Servers[serverID].Get(key)
	}
	return nil, errorInvalidServer
}
func (s *Storage) ServerSet(serverID, key string, val interface{}) {
	if s.Servers == nil {
		s.Servers = make(map[string]*StorageObject)
	}
	if _, exists := s.Servers[serverID]; !exists {
		s.Servers[serverID] = &StorageObject{
			Data: make(map[string]interface{}),
		}
	}
	s.Servers[serverID].Set(key, val)
}
func (s *Storage) ServerDel(serverID, key string) {
	if s.Servers == nil {
		s.Servers = make(map[string]*StorageObject)
	}
	if _, exists := s.Servers[serverID]; exists {
		s.Servers[serverID].Del(key)
	}
}

func (s *Storage) UserGet(userID, key string) (interface{}, error) {
	if s.Users == nil {
		s.Users = make(map[string]*StorageObject)
	}
	if _, exists := s.Users[userID]; exists {
		return s.Users[userID].Get(key)
	}
	return nil, errorInvalidUser
}
func (s *Storage) UserSet(userID, key string, val interface{}) {
	if s.Users == nil {
		s.Users = make(map[string]*StorageObject)
	}
	if _, exists := s.Users[userID]; !exists {
		s.Users[userID] = &StorageObject{
			Data: make(map[string]interface{}),
		}
	}
	s.Users[userID].Set(key, val)
}
func (s *Storage) UserDel(userID, key string) {
	if s.Users == nil {
		s.Users = make(map[string]*StorageObject)
	}
	if _, exists := s.Users[userID]; exists {
		s.Users[userID].Del(key)
	}
}