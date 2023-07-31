package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sData struct {
	Ingress []networkingv1.Ingress
}

func main() {
	kubeconfig := flag.String("kubeconfig", filepath.Join(os.ExpandEnv("$HOME"), ".kube", "config"), "Path to kubeconfig file")
	namespace := flag.String("n", "", "Namespace to read data from")

	if len(os.Args) < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s [flags] <template.txt>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	config, err := restConfig(*kubeconfig)
	if err != nil {
		log.Fatalf("Loading kubernetes config: %v", err)
	}

	k8s, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Creating kubernetes client: %v", err)
	}

	ingresses, err := k8s.NetworkingV1().Ingresses(*namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Listing ingresses: %v", err)
	}

	data := K8sData{
		Ingress: ingresses.Items,
	}

	templatePath := os.Args[1]
	templateBytes, err := os.ReadFile(templatePath)
	if err != nil {
		log.Fatalf("Reading template from %q: %v", templatePath, err)
	}

	tmpl, err := template.New("kube").Funcs(sprig.TxtFuncMap()).Funcs(tplFuncs()).Parse(string(templateBytes))
	if err != nil {
		log.Fatalf("Parsing template from %q: %v", templatePath, err)
	}

	err = tmpl.Execute(os.Stdout, &data)
	if err != nil {
		log.Fatalf("Executing template: %v", err)
	}
}

func restConfig(kubeconfig string) (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, err
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err == nil {
		return config, err
	}

	return nil, err
}

func tplFuncs() template.FuncMap {
	return template.FuncMap{
		"deref": func(ptr reflect.Value) reflect.Value {
			return ptr.Elem()
		},
		"asDict": func(mss map[string]string) map[string]interface{} {
			out := make(map[string]interface{}, len(mss))
			for k, v := range mss {
				out[k] = v
			}

			return out
		},
	}
}
