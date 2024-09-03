package storage

import (
	"errors"
	"sync"
)

var (
	errNowSuchUrl        = errors.New("unknown url")
	errNotUniqueShortUrl = errors.New("not unique short url")
	errNotUniqueOrigUrl  = errors.New("not unique original url")
)

type Inmem struct {
	urlList  map[string]string
	origUrls map[string]int
	lock     sync.Mutex
}

func NewInmem() *Inmem {
	memstor := &Inmem{}
	memstor.urlList = make(map[string]string)
	memstor.origUrls = make(map[string]int)
	return memstor
}

func (storage *Inmem) SaveData(og, short string) (string, error) {
	storage.lock.Lock()
	defer storage.lock.Unlock()
	if err := storage.uniqueCheck(og, short); err != nil {
		return "", err
	}
	storage.origUrls[og] = 1
	storage.urlList[short] = og
	return short, nil
}

func (storage *Inmem) GetData(short string) (string, error) {
	storage.lock.Lock()
	defer storage.lock.Unlock()
	og, ok := storage.urlList[short]
	if !ok {
		return "", errNowSuchUrl
	}
	return og, nil
}

func (storage *Inmem) uniqueCheck(og, short string) error {
	if _, ok := storage.origUrls[og]; ok {
		return errNotUniqueOrigUrl
	}
	if _, ok := storage.urlList[short]; ok {
		return errNotUniqueShortUrl
	}
	return nil
}
