package storage

type Storage interface {
	Save(filename string, data []byte) (string, error)
}
