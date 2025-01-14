package postgresql

//go:generate mockgen -source=writer.go -destination=mock/mock_reader.go
//go:generate gofumpt -s -w mock/mock_writer.go
type IReaderStore interface {
}
