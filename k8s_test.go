package main

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewK8SClient_MissingConfig(t *testing.T) {
	// Test with non-existent config path
	_, err := NewK8SClient("/nonexistent/path/kubeconfig")
	if err == nil {
		t.Error("expected error for non-existent config path, got nil")
	}
}

func TestListNamespaces(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-system"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kube-public"}},
	)

	client := &K8SClient{clientset: fakeClientset}
	ctx := context.Background()

	namespaces, err := client.ListNamespaces(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(namespaces) != 3 {
		t.Errorf("expected 3 namespaces, got %d", len(namespaces))
	}

	expectedNamespaces := map[string]bool{
		"default":     true,
		"kube-system": true,
		"kube-public": true,
	}

	for _, ns := range namespaces {
		if !expectedNamespaces[ns] {
			t.Errorf("unexpected namespace: %s", ns)
		}
	}
}

func TestListSecrets(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret1",
				Namespace: "default",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret2",
				Namespace: "default",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret3",
				Namespace: "kube-system",
			},
		},
	)

	client := &K8SClient{clientset: fakeClientset}
	ctx := context.Background()

	// List secrets in default namespace
	secrets, err := client.ListSecrets(ctx, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(secrets) != 2 {
		t.Errorf("expected 2 secrets in default namespace, got %d", len(secrets))
	}

	expectedSecrets := map[string]bool{
		"secret1": true,
		"secret2": true,
	}

	for _, s := range secrets {
		if !expectedSecrets[s] {
			t.Errorf("unexpected secret: %s", s)
		}
	}
}

func TestGetSecret(t *testing.T) {
	secretData := map[string][]byte{
		"username": []byte("admin"),
		"password": []byte("secret123"),
	}

	fakeClientset := fake.NewSimpleClientset(
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysecret",
				Namespace: "default",
			},
			Data: secretData,
		},
	)

	client := &K8SClient{clientset: fakeClientset}
	ctx := context.Background()

	secret, err := client.GetSecret(ctx, "default", "mysecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(secret) != 2 {
		t.Errorf("expected 2 keys, got %d", len(secret))
	}

	if string(secret["username"]) != "admin" {
		t.Errorf("expected username='admin', got '%s'", string(secret["username"]))
	}

	if string(secret["password"]) != "secret123" {
		t.Errorf("expected password='secret123', got '%s'", string(secret["password"]))
	}
}

func TestGetSecret_NotFound(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	client := &K8SClient{clientset: fakeClientset}
	ctx := context.Background()

	_, err := client.GetSecret(ctx, "default", "nonexistent")
	if err == nil {
		t.Error("expected error for non-existent secret, got nil")
	}
}

func TestSaveSecret(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysecret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"key1": []byte("value1"),
			},
		},
	)

	client := &K8SClient{clientset: fakeClientset}
	ctx := context.Background()

	// Update existing key
	newValue := []byte("updated_value")
	err := client.SaveSecret(ctx, "default", "mysecret", "key1", newValue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the update
	secret, err := client.GetSecret(ctx, "default", "mysecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(secret["key1"]) != "updated_value" {
		t.Errorf("expected key1='updated_value', got '%s'", string(secret["key1"]))
	}
}

func TestSaveSecret_NewKey(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset(
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysecret",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"key1": []byte("value1"),
			},
		},
	)

	client := &K8SClient{clientset: fakeClientset}
	ctx := context.Background()

	// Add new key
	newValue := []byte("value2")
	err := client.SaveSecret(ctx, "default", "mysecret", "key2", newValue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the update
	secret, err := client.GetSecret(ctx, "default", "mysecret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(secret) != 2 {
		t.Errorf("expected 2 keys, got %d", len(secret))
	}

	if string(secret["key2"]) != "value2" {
		t.Errorf("expected key2='value2', got '%s'", string(secret["key2"]))
	}
}

func TestSaveSecret_NotFound(t *testing.T) {
	fakeClientset := fake.NewSimpleClientset()

	client := &K8SClient{clientset: fakeClientset}
	ctx := context.Background()

	err := client.SaveSecret(ctx, "default", "nonexistent", "key", []byte("value"))
	if err == nil {
		t.Error("expected error for non-existent secret, got nil")
	}
}
