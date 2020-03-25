package monitor

type ContainerLogError struct {
	e error
}

func (err ContainerLogError) Error() string {
	return err.e.Error()
}

type K8SWatcherError struct {
	e error
}

func (err K8SWatcherError) Error() string {
	return err.e.Error()
}

func handelErr(err error){

}


