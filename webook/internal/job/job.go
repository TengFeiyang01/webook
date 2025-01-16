package job

type Job interface {
	Name() string
	Run() error
}
