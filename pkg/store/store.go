package store

import "errors"

var KeyNotFoundError = errors.New("key not found")

type DB interface {
	Init() error
	Get(key string) (string, error)
	Set(key string, value string) error
	Delete(key string) error
}

type StoreDB struct {
	db DB
}

func NewStore(db DB) *StoreDB {
	return &StoreDB{
		db: db,
	}
}

func (s *StoreDB) Init() error {
	if err := s.db.Init(); err != nil {
		return err
	}
	return nil
}

func (s *StoreDB) Get(key string) (string, error) {
	return s.db.Get(key)
}

func (s *StoreDB) Set(key string, value string) error {
	return s.db.Set(key, value)
}

func (s *StoreDB) Delete(key string) error {
	return s.db.Delete(key)
}
