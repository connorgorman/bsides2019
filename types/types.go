package types

type Capability struct {
	PID     int
	Cap     string
	Command string
}

type File struct {
	ContainerID string
	Path        string
}

type Container struct {
	ID, Name, Pod string
	PID           int
	FilePath      string
}

type ContainerPID struct {
	ID  string
	PID int
}
