package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	kaimera "github.com/kaimera-ai/kaimera/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ProxyServer struct {
	client client.Client
	logger logr.Logger
}

func New(client client.Client, logger logr.Logger) *ProxyServer {
	return &ProxyServer{
		client: client,
		logger: logger,
	}
}

func (server *ProxyServer) Start(address string) error {

	// server.address = address
	err := http.ListenAndServe(address, server)
	if err != nil {
		return err
	}

	return nil
}

func (server *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("Path invoked is %s", path)
	// Parse path; path format should be /[your-namespace]/[your-modeldeploymentname]/api/v1/chat...

	modelDeploymentStringParts := strings.Split(path, "/")
	if len(modelDeploymentStringParts) < 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	namespace := modelDeploymentStringParts[1]
	modelDeploymentName := modelDeploymentStringParts[2]

	// Find the corresponding CRD and match with its name
	err := server.client.Get(r.Context(),
		client.ObjectKey{Namespace: namespace, Name: modelDeploymentName},
		&kaimera.ModelDeployment{})

	if err != nil {
		server.logger.Info("Error retrieving model deployment", "namespace", namespace, "modelDeploymentName", modelDeploymentName, "err", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Route to the service at location nameoftheservice.namespace
	// http://mymodeldeployment.mynamespace/v1/chat
	pathFragment := strings.Join(modelDeploymentStringParts[3:], "/")
	targetUrl := fmt.Sprintf("http://%s.%s/%s", modelDeploymentName, namespace, pathFragment)
	url, err := url.Parse(targetUrl)
	if err != nil {
		server.logger.Info("Unable to parse URL", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	reverseProxy := NewSingleHostReverseProxy(url)
	reverseProxy.ServeHTTP(w, r)
}

func NewSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		req.URL.RawPath = target.RawPath
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	return &httputil.ReverseProxy{Director: director}
}
