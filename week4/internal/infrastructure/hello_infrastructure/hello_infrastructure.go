package hello_infrastructure

type PO struct {
	text string
}

// DB 假设这个是hello的db对象
type DB string

func NewDB() DB {
	return "db"
}
