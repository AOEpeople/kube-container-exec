package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	remotecommandconsts "k8s.io/apimachinery/pkg/util/remotecommand"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/exec"
)

// KubeClient is our configured client to access kubernetes
type KubeClient struct {
	namespace  string
	clientset  *kubernetes.Clientset
	kubeconfig clientcmd.ClientConfig
	restconfig *rest.Config
}

// KubeClientFromConfig loads a new KubeClient from the usual configuration
// (KUBECONFIG env param / selfconfigured in kubernetes)
func KubeClientFromConfig() (*KubeClient, error) {
	var client = new(KubeClient)
	var err error

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	configOverrides := &clientcmd.ConfigOverrides{}

	client.kubeconfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	client.restconfig, err = client.kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	client.clientset, err = kubernetes.NewForConfig(client.restconfig)
	if err != nil {
		return nil, err
	}

	client.namespace, _, err = client.kubeconfig.Namespace()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// GetPodByFilter returns the first pod found by a list with the given filter
// similiar to kubectl get pods
func (kc *KubeClient) GetPodByFilter(filter string) (*apiv1.Pod, error) {
	podClient := kc.clientset.CoreV1().Pods(kc.namespace)
	pods, err := podClient.List(metav1.ListOptions{
		LabelSelector: filter,
	})

	if err != nil {
		return nil, err
	}

	var pod apiv1.Pod
	for _, p := range pods.Items {
		if p.Status.Phase == apiv1.PodRunning {
			pod = p
			break
		}
	}

	return &pod, nil
}

// ExecInPod takes a pod and a container to find the correct container,
// then executes the commands in this container
// similiar to kubectl exec
func (kc *KubeClient) ExecInPod(pod *apiv1.Pod, container string, commands []string) error {
	restClient := kc.clientset.CoreV1().RESTClient()

	req := restClient.Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		Param("container", container).
		Param("stdout", "true").
		Param("stderr", "true")

	for _, command := range commands {
		req.Param("command", command)
	}

	executor, err := remotecommand.NewExecutor(kc.restconfig, http.MethodPost, req.URL())
	if err != nil {
		return err
	}

	err = os.Stdout.Sync()
	if err != nil {
		return err
	}
	err = os.Stderr.Sync()
	if err != nil {
		return err
	}

	err = executor.Stream(remotecommand.StreamOptions{
		SupportedProtocols: remotecommandconsts.SupportedStreamingProtocols,
		Stdin:              nil,
		Stdout:             os.Stdout,
		Stderr:             os.Stderr,
		Tty:                false,
		TerminalSizeQueue:  nil,
	})

	return err
}

func main() {
	var (
		filter     = os.Getenv("FILTER")
		container  = os.Getenv("CONTAINER")
		kubeconfig = os.Getenv("KUBECONFIG")
	)

	err := flag.Set("logtostderr", "true")
	if err != nil {
		glog.Fatal(err)
	}

	flag.StringVar(&filter, "filter", filter, "filter to apply to pod listing")
	flag.StringVar(&container, "container", container, "container to execute command in")
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "kubeconfig path")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "%s [flags] command params -foo -bar=a\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()

	// set KUBECONFIG from the flag to pass thru to clientcmd package.
	err = os.Setenv("KUBECONFIG", kubeconfig)
	if err != nil {
		glog.Fatal(err)
	}

	commands := flag.Args()

	if filter == "" {
		glog.Exit(`no pod-"filter" set`)
	}

	if container == "" {
		glog.Exit(`no "container" set`)
	}

	glog.V(4).Infof("Executing %v in container '%s' matching pod '%s'", commands, container, filter)

	client, err := KubeClientFromConfig()
	if err != nil {
		glog.Fatal(err)
	}

	pod, err := client.GetPodByFilter(filter)
	if err != nil {
		glog.Fatal(err)
	}

	if pod == nil {
		glog.Fatal("No (running) pod found")
	}

	glog.V(4).Infof("Using pod %s", pod.Name)

	err = client.ExecInPod(pod, container, commands)

	if err != nil {
		if err, ok := err.(exec.ExitError); ok {
			glog.Error(err)
			os.Exit(err.ExitStatus())
		}
		glog.Fatal(err)
	}
}
