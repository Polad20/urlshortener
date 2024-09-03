package storage

type Storage interface {
	SaveData(og, short string) (string, error)
	GetData(short string) (string, error)
}