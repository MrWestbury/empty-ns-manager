package main

import (
	"fmt"
	"log"
	"path"

	"github.com/mrwestbury/empty-ns-manager/internal"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	APPNAME = "empty-ns-manager"
)

var (
	APPVERSION = "local"
)

func main() {
	fmt.Printf("%s version: %s\n", APPNAME, APPVERSION)
	var kubecfg string

	if home := homedir.HomeDir(); home != "" {
		kubecfg = path.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubecfg)
	if err != nil {
		fmt.Println("Falling back to in-cluster config")

		config, err = rest.InClusterConfig()

		if err != nil {
			log.Panicf("error with in cluster config: %s", err.Error())
		}
	}

	stop := make(chan struct{})
	defer close(stop)

	app := internal.NewEmptyNSController(config, APPVERSION)
	app.Run(stop)

	fmt.Printf("%s %s controller started\n", APPNAME, APPVERSION)

	<-stop
}
