package configurator_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	k8sSecretNamespace = "test-ns"
	k8sSecretName      = "my-secret"
	k8sSecretKey       = "password"
	k8sSecretValue     = "super-secret-value"
)

func setupK3STestContainer(t *testing.T) (*typedcorev1.CoreV1Client, string, error) {
	t.Helper()
	k3sContainer, err := k3s.Run(t.Context(), "rancher/k3s:v1.35.2-k3s1")
	if err != nil {
		return nil, "", err
	}
	testcontainers.CleanupContainer(t, k3sContainer)

	kubeConfigYaml, err := k3sContainer.GetKubeConfig(t.Context())
	if err != nil {
		return nil, "", err
	}

	kubeconfigFile := filepath.Join(t.TempDir(), "kubeconfig")
	if err := os.WriteFile(kubeconfigFile, kubeConfigYaml, 0o600); err != nil {
		return nil, "", err
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeConfigYaml)
	if err != nil {
		return nil, "", err
	}

	k8sClient, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, "", err
	}

	return k8sClient, kubeconfigFile, nil
}

func createK8sTestSecrets(t *testing.T, k8sClient *typedcorev1.CoreV1Client) {
	t.Helper()
	_, err := k8sClient.Namespaces().Create(t.Context(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: k8sSecretNamespace,
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	_, err = k8sClient.Secrets(k8sSecretNamespace).Create(t.Context(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: k8sSecretName,
		},
		Data: map[string][]byte{
			k8sSecretKey: []byte(k8sSecretValue),
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	_, err = k8sClient.Secrets("default").Create(t.Context(), &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: k8sSecretName,
		},
		Data: map[string][]byte{
			k8sSecretKey: []byte(k8sSecretValue),
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
}

func TestPlugin_Kubernetes_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Kubernetes E2E integration test in short mode")
	}

	bin := buildCLI(t)
	k8sClient, kubeconfigFile, err := setupK3STestContainer(t)
	require.NoError(t, err)
	createK8sTestSecrets(t, k8sClient)

	env := cleanEnv(t,
		fmt.Sprintf("KUBECONFIG=%s", kubeconfigFile),
	)

	ctx := context.Background()

	t.Run("dig specific key in namespace", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"dig",
			fmt.Sprintf("k8s://%s/%s/%s", k8sSecretNamespace, k8sSecretName, k8sSecretKey),
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, k8sSecretValue, res.Stdout)
	})

	t.Run("default dig without subcommand in default namespace", func(t *testing.T) {
		res := runCLI(ctx, bin, env, fmt.Sprintf("k8s://%s/%s", k8sSecretName, k8sSecretKey))
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, k8sSecretValue, res.Stdout)
	})

	t.Run("dig with jsonpath modifier on whole secret", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"dig",
			fmt.Sprintf("k8s://%s/%s/?jp=$.%s", k8sSecretNamespace, k8sSecretName, k8sSecretKey),
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
		require.Equal(t, k8sSecretValue, res.Stdout)
	})

	t.Run("exist secret found", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"exists",
			fmt.Sprintf("k8s://%s/%s/%s", k8sSecretNamespace, k8sSecretName, k8sSecretKey),
		)
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})

	t.Run("exist secret missing", func(t *testing.T) {
		res := runCLI(
			ctx,
			bin,
			env,
			"exists",
			fmt.Sprintf("k8s://%s/missing-secret/%s", k8sSecretNamespace, k8sSecretKey),
		)
		require.NotEqual(t, 0, res.ExitCode)
	})

	t.Run("creds check", func(t *testing.T) {
		res := runCLI(ctx, bin, env, "creds")
		require.Equal(t, 0, res.ExitCode, res.Stderr)
	})
}
