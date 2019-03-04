package types

type Capability struct {
	PID     int
	Cap     string
	Command string
	Audit   string
}

type File struct {
	ContainerID string
	Path        string
}

type Container struct {
	ID, Name, Pod, Namespace string
	PID                      int
	FilePath                 string
	ReadonlyFS               bool
}

type ContainerPID struct {
	ID  string
	PID int
}

type Network struct {
	PID                   int
	Command, SAddr, DAddr string
	DPort                 int
	Call                  string
}
