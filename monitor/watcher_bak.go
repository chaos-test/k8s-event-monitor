package monitor
//
//import (
//	"k8s.io/client-go/kubernetes"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/watch"
//	"k8s.io/api/core/v1"
//	"strings"
//	log "github.com/sirupsen/logrus"
//	"io"
//	"bytes"
//	"io/ioutil"
//	"time"
//	"path"
//	"vision/pkg/email"
//	. "vision/configs"
//	//"os"
//	//"encoding/json"
//	//"net/http"
//)
//
//type Event struct {
//	Reason string
//	Log io.ReadCloser
//}
//
//func NewWatcher(k8s *kubernetes.Clientset, namespace string, blackList []string) *Watcher{
//	return &Watcher{
//		K8s: k8s,
//		Namespace:namespace,
//		BlackList:blackList,
//		event: make(chan Event, 1),
//	}
//}
//
//
//type Watcher struct {
//	K8s       *kubernetes.Clientset
//	Namespace string
//	BlackList []string
//	event chan Event
//
//}
//
//func (watcher *Watcher) watchEvent() {
//	w := func(namespace string, k8s *kubernetes.Clientset) (watch.Interface, error) {
//		return k8s.CoreV1().Events(namespace).Watch(metav1.ListOptions{})
//	}
//
//	eventWatcher, err := w(watcher.Namespace, watcher.K8s)
//	if err != nil {
//		panic(err)
//	}
//	log.Info("new start to monitor " + watcher.Namespace)
//	for {
//		event, ok := <-eventWatcher.ResultChan()
//		if !ok || event.Object == nil {
//			log.Info("the channel or Watcher is closed")
//			eventWatcher, err = w(watcher.Namespace, watcher.K8s)
//			if err != nil {
//				log.Error("watch the k8s ns failed")
//				panic(err)
//			}
//			continue
//		}
//
//		podEvent := event.Object.(*v1.Event)
//
//		if event.Type != watch.Added {
//
//			if (podEvent.Type == "Warning" && podEvent.Reason == "Unhealthy") || podEvent.Type == "Killing" ||  podEvent.Reason == "Killing"{
//
//				checkBackList := func() bool {
//
//					for _, p := range watcher.BlackList {
//						if strings.Contains(podEvent.InvolvedObject.Name, p) {
//							log.Info(podEvent.InvolvedObject.Name + "命中黑名单，不予监控event")
//							return true
//						}
//					}
//					return false
//				}
//				if checkBackList() {
//					continue
//				}
//
//				logger := log.WithFields(log.Fields{
//					"pod_name": podEvent.InvolvedObject.Name,
//					"Message":  podEvent.Message,
//					"reason":   podEvent.Reason,
//				})
//				logger.Info("warning event in event watcher, eventType: " + event.Type)
//				var subject string
//				//var level string
//
//				var body strings.Builder
//				body.WriteString("POD名称：")
//				body.WriteString(podEvent.InvolvedObject.Name)
//				body.WriteString("\n")
//				body.WriteString("异常原因：")
//				body.WriteString(podEvent.Message)
//				body.WriteString("\n")
//				body.WriteString("异常信息: ")
//				body.WriteString(podEvent.Message)
//
//				if podEvent.Type == "Killing" ||  podEvent.Reason == "Killing"{
//					subject = "容器由于异常被kill"
//					//level = "Exception"
//
//
//					filePath, err := watcher.saveLog("app", podEvent.InvolvedObject.Name)
//					if err != nil {
//						log.Error(err)
//						continue
//					}
//
//					err = email.Send(StableConf.Monitor.Emails, subject, body.String(), filePath)
//					if err != nil {
//						err = email.Send(StableConf.Monitor.Emails, subject, body.String())
//						log.Error("send email failed", err)
//					}
//				} else {
//					subject = "容器出现Warning"
//					//level = "Warning"
//				}
//
//				//type event struct {
//				//	Type int
//				//	Content struct {
//				//		Level   string `json:"level"`
//				//		PodName string `json:"pod_name"`
//				//		Reason  string `json:"reason"`
//				//		Message string `json:"message"`
//				//	}
//				//}
//				//e := new(event)
//				//e.Type = 1
//				//e.Content.PodName = podEvent.InvolvedObject.Name
//				//e.Content.Reason = podEvent.Reason
//				//e.Content.Message = podEvent.Message
//				//e.Content.Level = level
//
//				//jsons, err := json.Marshal(e)
//				//if err != nil {
//				//	log.Error("parse event json error", err)
//				//}
//				//callback(string(jsons))
//			}
//		}
//	}
//}
//
//func (watcher *Watcher) watchPod() {
//	watchPods := func(namespace string, k8s *kubernetes.Clientset) (watch.Interface, error) {
//		return k8s.CoreV1().Pods(namespace).Watch(metav1.ListOptions{})
//	}
//
//	podWatcher, err := watchPods(watcher.Namespace, watcher.K8s)
//	if err != nil {
//		log.Error("初始化k8s watcher 失败：", err)
//		panic(err)
//	}
//
//	for {
//		event, ok := <-podWatcher.ResultChan()
//		if !ok || event.Object == nil {
//			log.Info("the channel or Watcher is closed")
//			podWatcher, err = watchPods(watcher.Namespace, watcher.K8s)
//			if err != nil {
//				log.Error("初始化k8s watcher 失败：", err)
//				panic(err)
//			}
//			continue
//		}
//
//		//if event.Type == watch.Error || event.Type == watch.Modified|| event.Type == watch.Deleted {
//		if event.Type == watch.Error || event.Type == watch.Modified || event.Type == watch.Deleted {
//			pod, ok := event.Object.(*v1.Pod)
//
//			if !ok {
//				log.Error("unexpected type")
//			}
//
//			if watcher.checkBlackList(pod) {
//				log.Info(pod.Name + "命中黑名单，不予监控")
//				continue
//			}
//
//			for _, container := range pod.Status.ContainerStatuses {
//
//				if container.State.Terminated != nil {
//
//					if container.State.Terminated.Reason == "Completed" || container.State.Terminated.Reason == "" {
//						continue
//					}
//
//					logger := log.WithFields(log.Fields{
//						"pod_name":       pod.Name,
//						"container_name": container.Name,
//						"reason":         container.State.Terminated.Reason,
//					})
//					logger.Info("container is terminated, the event type is :" + event.Type)
//
//					filePath, err := watcher.saveLog(container.Name, pod.Name)
//					if err != nil {
//						log.Error(err)
//						continue
//					}
//
//					subject := "容器异常"
//					var body strings.Builder
//					body.WriteString("POD名称：")
//					body.WriteString(pod.Name)
//					body.WriteString("\n")
//					body.WriteString("容器名称：")
//					body.WriteString(container.Name)
//					body.WriteString("\n")
//					body.WriteString("异常原因：")
//					body.WriteString(container.State.Terminated.Reason)
//					body.WriteString("\n")
//					body.WriteString("error code: ")
//					body.WriteString(string(container.State.Terminated.ExitCode))
//					body.WriteString("\n")
//					body.WriteString("message: ")
//					body.WriteString(container.State.Terminated.Message)
//
//					err = email.Send(StableConf.Monitor.Emails, subject, body.String(), filePath)
//					if err != nil {
//						log.Error("send email failed", err)
//					}
//					type event struct {
//						Type int
//						Content struct {
//							ContainerName string `json:"container_name"`
//							Level         string `json:"level"`
//							PodName       string `json:"pod_name"`
//							Reason        string `json:"reason"`
//							Message       string `json:"message"`
//						}
//					}
//					e := new(event)
//					e.Type = 1
//					e.Content.PodName = pod.Name
//					e.Content.Reason = container.State.Terminated.Reason
//					e.Content.ContainerName = container.Name
//					e.Content.Level = "Exception"
//
//					//jsons, err := json.Marshal(e)
//					//if err != nil {
//					//	log.Error(err)
//					//}
//					//callback(string(jsons))
//
//				}
//			}
//		}
//	}
//}
//
//func (watcher *Watcher) Watch() {
//	go watcher.watchEvent()
//	go watcher.watchPod()
//
//	for {
//		time.Sleep(time.Hour)
//	}
//
//}
//
////func callback(json string) {
////	url := os.Getenv("webhook")
////	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(json))
////	if err != nil {
////		log.Error("send callback faild", err)
////	}
////	if _, err = http.DefaultClient.Do(req); err != nil {
////		log.Error("send callback faild", err)
////	}
////}
//
//func (watcher *Watcher) checkBlackList(pod *v1.Pod) (ok bool) {
//	ok = false
//	if watcher.BlackList != nil {
//		for _, v := range watcher.BlackList {
//			if strings.Contains(pod.Name, v) {
//				ok = true
//				break
//			}
//		}
//	}
//	return
//}
//
//func (watcher *Watcher) saveLog(containerName string, podName string) (string, error) {
//	// 创建存储日志的目录
//	ok, err := CreateDir(LogDir)
//	if !ok {
//		return "", err
//	}
//
//	// 抓取container日志
//	line := int64(1000) // 定义只抓取前1000行日志
//	opts := &v1.PodLogOptions{
//		Container: containerName,
//		TailLines: &line,
//	}
//	containerLog, err := watcher.K8s.CoreV1().Pods(watcher.Namespace).GetLogs(podName, opts).Stream()
//	if err != nil {
//		return "", err
//	}
//
//	defer containerLog.Close()
//	buf := new(bytes.Buffer)
//	_, err = io.Copy(buf, containerLog)
//	if err != nil {
//		return "", err
//	}
//
//	var logName strings.Builder
//	logName.WriteString(podName)
//	logName.WriteString("_")
//	logName.WriteString(containerName)
//	logName.WriteString(time.Now().Format("2006-01-02 15:04:05"))
//	logName.WriteString(".log")
//
//	filePath := path.Join(LogDir, logName.String())
//	err = ioutil.WriteFile(filePath, buf.Bytes(), 0666)
//
//	if err != nil {
//		return "", err
//	}
//
//	return filePath, nil
//
//}
//
//func Watch(k8s *kubernetes.Clientset) {
//	for _, ns := range StableConf.Monitor.Namespaces {
//		//log.Info("new start to monitor " + ns)
//		podWatcher := Watcher{Namespace: ns, K8s: k8s, BlackList: StableConf.Monitor.BlackList}
//		go podWatcher.Watch()
//	}
//}
