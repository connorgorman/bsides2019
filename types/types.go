package types

type Capability struct {
	ContainerID string
	PID int
	Cap string
	Command string
}

type File struct {
	ContainerID string
	Path string
}

type Container struct {
	ID, Name, Pod string
	PID int
	FilePath string
}

type CapabilityMessage struct {
	Container Container
	Capability Capability
}

type FileMessage struct {
	Container Container
	File File
}
