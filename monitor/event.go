package monitor

import (
	"k8s.io/client-go/kubernetes"
	"strings"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/watch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"io"
	"github.com/pkg/errors"
	"time"
)

type EventWatcher struct {
	K8s       *kubernetes.Clientset
	Namespace string
	BlackList []string
	event     chan Event
}

func (watcher *EventWatcher) Watch() {
	w := func(namespace string, k8s *kubernetes.Clientset) (watch.Interface, error) {
		return k8s.CoreV1().Events(namespace).Watch(metav1.ListOptions{})
	}

	eventWatcher, err := w(watcher.Namespace, watcher.K8s)
	if err != nil {
		watcher.handelErr(err)
	}
	log.Info("new start to monitor " + watcher.Namespace)
	for {
		event, ok := <-eventWatcher.ResultChan()
		if !ok || event.Object == nil {
			log.Info("the channel or Watcher is closed")
			eventWatcher, err = w(watcher.Namespace, watcher.K8s)
			if err != nil {
				watcher.handelErr(err)
				time.Sleep(time.Minute * 5)
			}
			continue
		}

		podEvent := event.Object.(*corev1.Event)

		if event.Type != watch.Added {

			if (podEvent.Type == "Warning" && podEvent.Reason == "Unhealthy") || podEvent.Type == "Killing" || podEvent.Reason == "Killing" {
				if watcher.checkBackList(podEvent) {
					continue
				}

				logger := log.WithFields(log.Fields{
					"pod_name": podEvent.InvolvedObject.Name,
					"Message":  podEvent.Message,
					"reason":   podEvent.Reason,
				})
				logger.Info("warning event in event watcher, eventType: " + event.Type)

				if podEvent.Type == "Killing" || podEvent.Reason == "Killing" {
					e := Event{
						Namespace: watcher.Namespace,
						PodName: podEvent.InvolvedObject.Name,
						Reason:  podEvent.Reason,
						Message: podEvent.Message,
						Error:   nil,
					}
					if containerLogs, err := watcher.getLog(podEvent.InvolvedObject.Name); err != nil {
						e.Error = errors.Wrapf(ContainerLogError{err: err}, "failed to fetch container log in namespace:%s Pod:%s", watcher.Namespace, podEvent.InvolvedObject.Name)
					} else {
						e.Log = containerLogs
					}
					watcher.event <- e
				}
			}
		}
	}
}

func (watcher *EventWatcher) getLog(podName string) ([]io.ReadCloser, error) {
	pod, _ := watcher.K8s.CoreV1().Pods(watcher.Namespace).Get(podName, metav1.GetOptions{})
	var containerLogs []io.ReadCloser

	for _, c := range pod.Status.ContainerStatuses {
		// 抓取container日志
		line := int64(1000) // 定义只抓取前1000行日志
		opts := &corev1.PodLogOptions{
			Container: c.Name,
			TailLines: &line,
		}
		cLog, err := watcher.K8s.CoreV1().Pods(watcher.Namespace).GetLogs(podName, opts).Stream()
		if err != nil {
			return nil, err
		}
		containerLogs = append(containerLogs, cLog)
	}

	return containerLogs, nil
}

func (watcher *EventWatcher) checkBackList(podEvent *corev1.Event) bool {

	for _, p := range watcher.BlackList {
		if strings.Contains(podEvent.InvolvedObject.Name, p) {
			log.Info(podEvent.InvolvedObject.Name + "命中黑名单，不予监控event")
			return true
		}
	}
	return false
}

func(watcher *EventWatcher) handelErr(err error){
	k8sError := K8SWatcherError{err: err}
	e := Event{
		Error: errors.Wrapf(k8sError, "watch event of namespace %s failed", watcher.Namespace),
	}
	watcher.event <- e
	log.Errorf("watch the k8s namespace %s failed", watcher.Namespace)
}