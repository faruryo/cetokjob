package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/faruryo/cetokjob/cetokjob"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	jobGenerator *cetokjob.JobGenerator
)

func usage() {
	fmt.Printf("usage: cetokjob [path ...]\n")
	flag.PrintDefaults()
}

func parse() []cetokjob.JobConfig {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
		os.Exit(1)
	}

	var jobConfigs []cetokjob.JobConfig
	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			fmt.Printf("os.Stat error : %s\n", err)
			os.Exit(1)
		case dir.IsDir():
			fmt.Printf("Directory not supported\n")
			os.Exit(1)
		default:
			jc, err := cetokjob.LoadJobConfig(path)
			if err != nil {
				fmt.Printf("Failed load %s : %s\n", path, err)
				os.Exit(1)
			}
			jobConfigs = append(jobConfigs, jc...)
		}
	}
	fmt.Printf("load job config : %#v\n", jobConfigs)

	return jobConfigs
}

func getCurrentNamespace() (string, error) {
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
	}

	return "", errors.New("Failed current namespace")
}

func prepareClientset() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func main() {
	jobConfigs := parse()

	currentNamespace, err := getCurrentNamespace()
	if err != nil {
		fmt.Printf("Failed to get current namespace : %s\n", err)
		os.Exit(1)
	}

	clientset, err := prepareClientset()
	if err != nil {
		fmt.Printf("Failed to prepare clientset : %s\n", err)
		os.Exit(1)
	}

	jobGenerator = cetokjob.NewJobGenerator(currentNamespace, clientset, jobConfigs)

	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		fmt.Printf("Failed to create cloudevents client, %s\n", err)
		os.Exit(1)
	}

	log.Fatal(c.StartReceiver(context.Background(), receiver))
}
