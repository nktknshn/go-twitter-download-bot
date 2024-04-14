package bot

import (
	"sync"
)

// not implemented

type UserStorageJSON struct {
	Filepath string
	Users    map[int64]*UserData
	mutex    *sync.Mutex
}

func NewUserStorageJSON(filepath string) *UserStorageJSON {
	panic("not implemented")
	return nil
}

func (us *UserStorageJSON) Load() error {
	return nil
}

func (us *UserStorageJSON) Save() error {
	return nil
}

func (us *UserStorageJSON) Get(userID int64) *UserData {
	return nil
}

func (us *UserStorageJSON) Set(userID int64, data *UserData) error {
	return nil
}

func (us *UserStorageJSON) IncQueries(userID int64) error {
	return nil
}
