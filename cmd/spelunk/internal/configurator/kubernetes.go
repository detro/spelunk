package configurator

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/detro/spelunk/cmd/spelunk/internal"
	"github.com/detro/spelunk/cmd/spelunk/internal/logger"
	spelunkk8s "github.com/detro/spelunk/plugin/source/kubernetes/v2"
	"github.com/detro/spelunk/v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesConfigurator struct {
	Kubeconfig string `name:"kubeconfig" env:"KUBECONFIG" help:"Path to Kubeconfig file."`
}

var _ internal.SecretSourceConfigurator = (*KubernetesConfigurator)(nil)

func (c *KubernetesConfigurator) Type() string {
	return spelunkk8s.Type
}

func (c *KubernetesConfigurator) loadConfig() (*rest.Config, error) {
	// 1. Explicit flag or env var
	if c.Kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	}
	if envKube := os.Getenv("KUBECONFIG"); envKube != "" {
		return clientcmd.BuildConfigFromFlags("", envKube)
	}
	// 2. In-cluster config
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}
	// 3. Canonical default ~/.kube/config
	defaultKubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
	if defaultKubeconfig != "" {
		if _, err := os.Stat(defaultKubeconfig); err == nil {
			return clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
		}
	}
	return nil, fmt.Errorf("no kubernetes configuration found")
}

func (c *KubernetesConfigurator) CredentialsDetected() bool {
	if c.Kubeconfig != "" || os.Getenv("KUBECONFIG") != "" {
		return true
	}
	if _, err := c.loadConfig(); err == nil {
		return true
	}
	return false
}

func (c *KubernetesConfigurator) newClient() (*kubernetes.Clientset, error) {
	restConfig, err := c.loadConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(restConfig)
}

func (c *KubernetesConfigurator) SpelunkerOption(
	ctx context.Context,
) (spelunk.SpelunkerOption, error) {
	if !c.CredentialsDetected() {
		slog.Log(ctx, logger.LevelTrace, "skipped (no credentials detected)", "plugin", c.Type())
		return nil, nil
	}
	slog.Log(ctx, logger.LevelTrace, "detected credentials", "plugin", c.Type())

	clientset, err := c.newClient()
	if err != nil {
		return nil, err
	}
	slog.Log(ctx, logger.LevelTrace, "configured client", "plugin", c.Type())

	return spelunkk8s.WithKubernetes(clientset.CoreV1()), nil
}

func (c *KubernetesConfigurator) CredentialsValid(_ context.Context) error {
	if !c.CredentialsDetected() {
		return fmt.Errorf("%w for plugin %s", ErrCredentialsNotDetected, c.Type())
	}
	clientset, err := c.newClient()
	if err != nil {
		return err
	}
	_, err = clientset.Discovery().ServerVersion()
	return err
}
