package monitor

import (
	"k8s.io/client-go/kubernetes"
	//log "github.com/sirupsen/logrus"
	"io"
)

type Event struct {
	Namespace string
	PodName string
	Reason  string
	Log     []io.ReadCloser
	Message string
	Error error
}

func NewWatcher(k8s *kubernetes.Clientset, namespace string, blackList []string) *Watcher {
	return &Watcher{
		K8s:       k8s,
		Namespace: namespace,
		BlackList: blackList,
		event:     make(chan Event, 100),
	}
}

type Watcher struct {
	K8s       *kubernetes.Clientset
	Namespace string
	BlackList []string
	event     chan Event
}

func (watcher *Watcher) Watch() chan Event{

	eWatcher := EventWatcher{
		K8s: watcher.K8s,
		Namespace:watcher.Namespace,
		BlackList:watcher.BlackList,
		event:watcher.event,
	}

	pWatcher := PodWatcher{
		K8s: watcher.K8s,
		Namespace:watcher.Namespace,
		BlackList:watcher.BlackList,
		event:watcher.event,
	}

	go eWatcher.Watch()
	go pWatcher.Watch()

	return watcher.event
}


