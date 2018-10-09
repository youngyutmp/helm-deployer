package service

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/entwico/helm-deployer/domain"
	log "github.com/sirupsen/logrus"
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
	logger          *log.Entry
	managedReleases map[string]*managedRelease
	mutex           sync.Mutex
}

//NewK8SReleaseProvider returns new instance of K8SReleaseProvider
func NewK8SReleaseProvider(k8sConfigPath string, logger *log.Entry) (domain.K8SReleaseProvider, error) {
	client, err := getClient(k8sConfigPath, logger)
	if err != nil {
		return nil, err
	}

	return &k8sReleaseProvider{client: client, logger: logger, managedReleases: make(map[string]*managedRelease)}, nil
}

func (s *k8sReleaseProvider) Start() {
	stop := make(chan struct{})
	s.logger.Debug("start informers...")
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
	s.logger.WithField("image_path", path).Debug("searching for deploy configs")
	results := make([]*domain.DeployConfig, 0)
	for name, managedRelease := range s.managedReleases {
		for _, image := range managedRelease.images {
			if strings.HasSuffix(image, path) {
				s.logger.WithFields(log.Fields{
					"name": name,
					"path": path,
				}).Debug("release found for image path")
				results = append(results, managedRelease.cfg)
			}
		}
	}
	if len(results) == 0 {
		s.logger.WithField("image_path", path).Warn("no deploy configs found for image path")
	}

	return results, nil
}
func extractDeployConfig(labels map[string]string) (*domain.DeployConfig, error) {
	var chart, release string
	if val, ok := labels["chart"]; ok {
		chart = val
	}
	if val, ok := labels["helm.sh/chart"]; ok {
		chart = val
	}
	if val, ok := labels["release"]; ok {
		release = val
	}
	if val, ok := labels["app.kubernetes.io/instance"]; ok {
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

func getClient(k8sConfigPath string, logger *log.Entry) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if k8sConfigPath == "" {
		logger.Info("using in cluster config")
		config, err = rest.InClusterConfig()
		// in cluster access
	} else {
		logger.Info("using out of cluster config")
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
			s.logger.WithField("error", err).Warning("could not extract deploy config")
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
			s.logger.WithField("image", container.Image).Debug("image found")
			images = append(images, container.Image)
		}
	case *extensionsV1.Ingress:
		images = append(images, s.extractImagesFromIngres(item)...)
	case *coreV1.Service:
		images = append(images, s.extractImagesFromService(item)...)
	default:
		s.logger.WithField("item", fmt.Sprintf("%T!", item)).Warn("unable to get images")
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
				s.logger.WithFields(log.Fields{
					"service_name": serviceName,
					"error":        err,
				}).Warn("could not get service")
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
		s.logger.WithField("error", err).Warn("could not select pod")
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
				s.logger.WithFields(log.Fields{
					"release":       release.cfg.ReleaseName,
					"chart_name":    release.cfg.ChartName,
					"chart_version": release.cfg.ChartVersion,
				}).Info("managed release added")
				//s.logger.WithField("total", len(s.managedReleases)).Debug("managed releases found")
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
				s.logger.WithField("release", release.cfg.ReleaseName).Info("managed release deleted")
				s.logger.WithField("total", len(s.managedReleases)).Debug("managed releases found")
			}
			s.mutex.Unlock()
		}
	}
}
