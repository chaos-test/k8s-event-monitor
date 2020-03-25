package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	//"vision/pkg/osutil"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"fmt"
)

func init() {
	log.SetOutput(os.Stdout)
	//
	//log.Info("begin to init config")
	////flag, err := osutil.CreateDir(DataDir)
	////if !flag {
	////	panic(err)
	////}
	////ReadConfig("configs/config.json")
	//log.Info("init config done")

}

func main() {
	log.Info("init the kubeconfig")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", "./kubeconfig")
	if err != nil {
		log.Error("cannot init the kubeconfig")
		panic(err.Error())
	}
	log.Info("init the kubeconfig done")
	// create the clientset
	log.Info("init the k8sclient")
	k8s, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Error("cannot init the k8s client")
		panic(err.Error())
	}
	log.Info("init the k8sclient done, now begin to monitor the k8s")
	var tmp []string
	eventChan := NewWatcher(k8s, "371cdh", tmp).Watch()

	for {
		select {
		case event := <- eventChan:
			fmt.Println(event.Reason)
		}
	}

}
