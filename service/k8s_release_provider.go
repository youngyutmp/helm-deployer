package service

import (
	"fmt"
	"strings"
	"time"

	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/entwico/helm-deployer/domain"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	extensionsV1 "k8s.io/api/extensions/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type managedRelease struct {
	cfg    *domain.DeployConfig
	images []string
}

type k8sReleaseProvider struct {
	client          *kubernetes.Clientset
	managedReleases map[string]*managedRelease
	mutex           sync.Mutex
}

//NewK8SReleaseProvider returns new instance of K8SReleaseProvider
func NewK8SReleaseProvider(k8sConfigPath string) (domain.K8SReleaseProvider, error) {
	client, err := getClient(k8sConfigPath)
	if err != nil {
		return nil, err
	}

	return &k8sReleaseProvider{client: client, managedReleases: make(map[string]*managedRelease)}, nil
}

func (s *k8sReleaseProvider) Start() {
	stop := make(chan struct{})
	logrus.Debug("start informers...")
	for _, item := range s.getInformers() {
		go func(informer cache.Controller) {
			informer.Run(stop)
		}(item)
	}
}

func (s *k8sReleaseProvider) getInformers() []cache.Controller {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc:    s.handleAdd,
		UpdateFunc: s.handleUpdate,
		DeleteFunc: s.handleDelete,
	}
	_, deploymentsInformer := cache.NewInformer(
		cache.NewListWatchFromClient(s.client.AppsV1().RESTClient(), "deployments", coreV1.NamespaceAll,
			fields.Everything()),
		&appsV1.Deployment{},
		time.Second*30,
		handlers,
	)
	_, ingressInformer := cache.NewInformer(
		cache.NewListWatchFromClient(s.client.ExtensionsV1beta1().RESTClient(), "ingresses", coreV1.NamespaceAll,
			fields.Everything()),
		&extensionsV1.Ingress{},
		time.Second*30,
		handlers,
	)

	_, serviceInformer := cache.NewInformer(
		cache.NewListWatchFromClient(s.client.CoreV1().RESTClient(), "services", coreV1.NamespaceAll,
			fields.Everything()),
		&coreV1.Service{},
		time.Second*30,
		handlers,
	)

	informers := []cache.Controller{deploymentsInformer, ingressInformer, serviceInformer}

	return informers
}

func (s *k8sReleaseProvider) GetDeployConfigsForImagePath(path string) ([]*domain.DeployConfig, error) {
	logrus.Debugf("searching for deploy configs for image %s", path)
	results := make([]*domain.DeployConfig, 0)
	for name, managedRelease := range s.managedReleases {
		for _, image := range managedRelease.images {
			if strings.HasSuffix(image, path) {
				logrus.Debugf("release %s found for image path %s", name, path)
				results = append(results, managedRelease.cfg)
			}
		}
	}
	if len(results) == 0 {
		logrus.Warnf("no deploy configs found for image path %s", path)
	}

	return results, nil
}
func extractDeployConfig(labels map[string]string) (*domain.DeployConfig, error) {
	var chart, release string
	if val, ok := labels["chart"]; ok {
		chart = val
	}
	if val, ok := labels["release"]; ok {
		release = val
	}
	if chart == "" || release == "" {
		return nil, fmt.Errorf("could not extract deploy config from labels %v", labels)
	}

	var chartName, chartVersion string

	if index := strings.LastIndex(chart, "-"); index != -1 {
		chartName = chart[:index]
		chartVersion = chart[index+1:]
	}

	cfg := &domain.DeployConfig{
		ReleaseName:  release,
		ChartName:    chartName,
		ChartVersion: chartVersion,
	}

	return cfg, nil
}

func getClient(k8sConfigPath string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if k8sConfigPath == "" {
		logrus.Info("using in cluster config")
		config, err = rest.InClusterConfig()
		// in cluster access
	} else {
		logrus.Info("using out of cluster config")
		config, err = clientcmd.BuildConfigFromFlags("", k8sConfigPath)
	}
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func isManagedObject(obj interface{}) bool {
	if acc, ok := obj.(metaV1.ObjectMetaAccessor); ok {
		meta := acc.GetObjectMeta()
		annotations := meta.GetAnnotations()
		if val, ok := annotations["helm-deployer/enabled"]; ok {
			if val == "\"true\"" {
				return true
			}
		}
	}
	return false
}

func (s *k8sReleaseProvider) extractManagedRelease(obj interface{}) *managedRelease {
	if acc, ok := obj.(metaV1.ObjectMetaAccessor); ok {
		meta := acc.GetObjectMeta()
		cfg, err := extractDeployConfig(meta.GetLabels())
		if err != nil {
			logrus.Warnf("could not extract deploy config: %s", err)
			return nil
		}
		return &managedRelease{cfg: cfg}
	}
	return nil
}

func (s *k8sReleaseProvider) getImages(obj interface{}) []string {
	images := make([]string, 0)
	switch item := obj.(type) {
	case *appsV1.Deployment:
		for _, container := range item.Spec.Template.Spec.Containers {
			logrus.Debugf("found image %s", container.Image)
			images = append(images, container.Image)
		}
	case *extensionsV1.Ingress:
		images = append(images, s.extractImagesFromIngres(item)...)
	case *coreV1.Service:
		images = append(images, s.extractImagesFromService(item)...)
	default:
		logrus.Warnf("unable to get images from %T!", item)
	}
	return images
}

func (s *k8sReleaseProvider) extractImagesFromIngres(item *extensionsV1.Ingress) []string {
	imgMap := map[string]bool{}
	for _, rule := range item.Spec.Rules {
		for _, path := range rule.HTTP.Paths {
			serviceName := path.Backend.ServiceName
			svc, err := s.client.CoreV1().Services(item.Namespace).Get(serviceName, metaV1.GetOptions{})
			if err != nil {
				logrus.Warnf("could not get service for name %s: %v", serviceName, err)
			}
			for _, image := range s.extractImagesFromService(svc) {
				imgMap[image] = true
			}
		}
	}
	images := make([]string, 0)
	for image := range imgMap {
		images = append(images, image)
	}
	return images
}

func (s *k8sReleaseProvider) extractImagesFromService(svc *coreV1.Service) []string {
	imgMap := map[string]bool{}
	set := labels.Set(svc.Spec.Selector)
	pods, err := s.client.CoreV1().Pods(svc.Namespace).List(metaV1.ListOptions{LabelSelector: set.AsSelector().String()})
	if err != nil {
		logrus.Warnf("could not select pods: %v", err)
	}
	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			imgMap[container.Image] = true
		}
	}
	images := make([]string, 0)
	for image := range imgMap {
		images = append(images, image)
	}
	return images
}

func (s *k8sReleaseProvider) handleAdd(obj interface{}) {
	if isManagedObject(obj) {
		if release := s.extractManagedRelease(obj); release != nil {
			s.mutex.Lock()
			if _, ok := s.managedReleases[release.cfg.ReleaseName]; !ok {
				release.images = s.getImages(obj)
				s.managedReleases[release.cfg.ReleaseName] = release
				logrus.Debugf("managed release [%s] added", release.cfg.ReleaseName)
				logrus.Debugf("total managed %d", len(s.managedReleases))
			}
			s.mutex.Unlock()
		}
	}
}

func (s *k8sReleaseProvider) handleUpdate(old, current interface{}) {
	if !isManagedObject(old) && isManagedObject(current) {
		s.handleAdd(current)
	}
	if isManagedObject(old) && !isManagedObject(current) {
		s.handleDelete(old)
	}
	if isManagedObject(current) {
		if release := s.extractManagedRelease(current); release != nil {
			if _, ok := s.managedReleases[release.cfg.ReleaseName]; !ok {
				s.handleAdd(current)
			}
		}
	}
}

func (s *k8sReleaseProvider) handleDelete(obj interface{}) {
	if isManagedObject(obj) {
		if release := s.extractManagedRelease(obj); release != nil {
			s.mutex.Lock()
			if _, ok := s.managedReleases[release.cfg.ReleaseName]; ok {
				delete(s.managedReleases, release.cfg.ReleaseName)
				logrus.Infof("managed release [%s] deleted", release.cfg.ReleaseName)
				logrus.Debugf("total managed %d", len(s.managedReleases))
			}
			s.mutex.Unlock()
		}
	}
}
