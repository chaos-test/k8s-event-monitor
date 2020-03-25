package monitor

import (
	"io"
	"time"
	"strings"

	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"github.com/pkg/errors"

)

type PodWatcher struct {
	K8s       *kubernetes.Clientset
	Namespace string
	BlackList []string
	event     chan Event
}

func (watcher *PodWatcher) Watch() {
	watchPods := func(namespace string, k8s *kubernetes.Clientset) (watch.Interface, error) {
		return k8s.CoreV1().Pods(namespace).Watch(metav1.ListOptions{})
	}

	podWatcher, err := watchPods(watcher.Namespace, watcher.K8s)
	if err != nil {
		log.Errorf("watch pod of namespace %s failed", watcher.Namespace)
		//return errors.Wrapf(err, "watch pod of namespace %s failed", watcher.Namespace)
	}

	for {
		event, ok := <-podWatcher.ResultChan()
		if !ok || event.Object == nil {
			log.Info("the channel or Watcher is closed")
			podWatcher, err = watchPods(watcher.Namespace, watcher.K8s)
			if err != nil {
				k8sError := K8SWatcherError{err: err}
				e := Event{
					Error: errors.Wrapf(k8sError, "watch pod of namespace %s failed", watcher.Namespace),
				}
				watcher.event <- e
				log.Errorf("watch the k8s namespace %s failed", watcher.Namespace)
				time.Sleep(time.Minute * 5)
			}
			continue
		}

		if event.Type == watch.Error || event.Type == watch.Modified || event.Type == watch.Deleted {
			log.Info("event start")
			pod, _ := event.Object.(*corev1.Pod)

			if watcher.checkBlackList(pod) {
				log.Info(pod.Name + "命中黑名单，不予监控")
				continue
			}

			for _, container := range pod.Status.ContainerStatuses {

				if container.State.Terminated != nil {

					if container.State.Terminated.Reason == "Completed" || container.State.Terminated.Reason == "" {
						continue
					}

					logger := log.WithFields(log.Fields{
						"pod_name":       pod.Name,
						"container_name": container.Name,
						"reason":         container.State.Terminated.Reason,
					})
					logger.Info("container is terminated, the event type is :" + event.Type)
					cLog, err := watcher.getLog(container.Name, pod.Name)

					e := Event{
						PodName: pod.Name,
						Namespace: watcher.Namespace,
						Reason: container.State.Terminated.Reason,
						Message: container.State.Terminated.Message,
						Error: nil,
					}
					if err!=nil{
						e.Error = errors.Wrapf(ContainerLogError{err: err}, "failed to fetch container log in namespace:%s Pod:%s", watcher.Namespace, pod.Name)
					}else{
						e.Log = cLog
					}
					watcher.event <- e
				}
			}
		}
	}
}

func (watcher *PodWatcher) checkBlackList(pod *corev1.Pod) (ok bool) {
	ok = false
	if watcher.BlackList != nil {
		for _, v := range watcher.BlackList {
			if strings.Contains(pod.Name, v) {
				ok = true
				break
			}
		}
	}
	return
}



func (watcher *PodWatcher) getLog(containerName string, podName string) ([]io.ReadCloser, error) {
	// 抓取container日志
	line := int64(1000) // 定义只抓取前1000行日志
	opts := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &line,
	}
	containerLog, err := watcher.K8s.CoreV1().Pods(watcher.Namespace).GetLogs(podName, opts).Stream()
	if err != nil {
		return nil, err
	}
	var logs []io.ReadCloser
	logs = append(logs, containerLog)
	return logs, nil
}
