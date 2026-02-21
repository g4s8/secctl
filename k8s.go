package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8SClient struct {
	clientset kubernetes.Interface
}

func NewK8SClient(cfgPath string) (*K8SClient, error) {
	if cfgPath == "" {
		cfgPath = clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
	}
	config, err := clientcmd.BuildConfigFromFlags("", cfgPath)
	if err != nil {
		return nil, fmt.Errorf("build config from flags: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8SClient{clientset: clientset}, nil
}

func (k *K8SClient) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := k.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}

	items := make([]string, len(namespaces.Items))
	for i, ns := range namespaces.Items {
		items[i] = ns.Name
	}
	return items, nil
}

func (k *K8SClient) ListSecrets(ctx context.Context, namespace string) ([]string, error) {
	secrets, err := k.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list secrets in namespace '%s': %w", namespace, err)
	}

	items := make([]string, len(secrets.Items))
	for i, secret := range secrets.Items {
		items[i] = secret.Name
	}
	return items, nil
}

type SecretData map[string][]byte

func (k *K8SClient) GetSecret(ctx context.Context, namespace, name string) (SecretData, error) {
	secret, err := k.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get secret '%s' in namespace '%s': %w", name, namespace, err)
	}
	return secret.Data, nil
}

func (k *K8SClient) SaveSecret(ctx context.Context, namespace, name, key string, data []byte) error {
	secret, err := k.clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if secret.Data == nil {
		secret.Data = make(map[string][]byte)
	}
	secret.Data[key] = data

	if _, err = k.clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("update secret '%s' in namespace '%s': %w", name, namespace, err)
	}
	return nil
}
