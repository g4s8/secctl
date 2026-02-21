package main

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	var cfg Config
	cfg.Parse()

	if cfg.showVersion {
		fmt.Printf("k8s-secret-editor version %s ("+
			"commit: %s, built at: %s, built by: %s"+
			")\n", version, commit, date, builtBy)
	}

	k8sClient, err := NewK8SClient(cfg.KubeConfig)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	editor, err := NewEditor(cfg.EditorPath)
	if err != nil {
		log.Fatalf("Error initializing editor: %v", err)
	}

	fmt.Printf("Loading namespaces...\n")
	namespaces, err := withTimeoutCtx(func(ctx context.Context) ([]string, error) {
		return k8sClient.ListNamespaces(ctx)
	})
	if err != nil {
		log.Fatalf("Error loading namespaces: %v", err)
	}
	selectedNamespace := runPrompt("Select namespace", namespaces)

	fmt.Printf("Loading secrets in namespace '%s'...\n", selectedNamespace)
	secrets, err := withTimeoutCtx(func(ctx context.Context) ([]string, error) {
		return k8sClient.ListSecrets(ctx, selectedNamespace)
	})
	if err != nil {
		log.Fatalf("Error loading secrets: %v", err)
	}
	selectedSecret := runPrompt(fmt.Sprintf("Select secret in '%s'", selectedNamespace), secrets)

	fmt.Printf("Loading secret '%s' in namespace '%s'...\n", selectedSecret, selectedNamespace)
	secret, err := withTimeoutCtx(func(ctx context.Context) (SecretData, error) {
		return k8sClient.GetSecret(ctx, selectedNamespace, selectedSecret)
	})
	if err != nil {
		log.Fatalf("Error loading secret: %v", err)
	}
	keys := make([]string, 0, len(secret))
	for k := range secret {
		keys = append(keys, k)
	}
	selectedKey := runPrompt(fmt.Sprintf("Select key in secret '%s'", selectedSecret), keys)
	data, ok := secret[selectedKey]
	if !ok {
		log.Fatalf("Key '%s' not found in secret '%s' in namespace '%s'", selectedKey, selectedSecret, selectedNamespace)
	}

	tmpFile, err := NewTmpFile()
	if err != nil {
		log.Fatalf("Error creating temp file: %v", err)
	}
	defer tmpFile.Close()
	if err := tmpFile.Write(data); err != nil {
		log.Fatalf("Error writing secret data to temp file: %v", err)
	}
	if err := tmpFile.OpenEditor(editor); err != nil {
		log.Fatalf("Error opening editor: %v", err)
	}
	editedData, err := tmpFile.Read()
	if err != nil {
		log.Fatalf("Error reading edited data from temp file: %v", err)
	}

	confirmPrompt := promptui.Prompt{
		Label:     fmt.Sprintf("Save changes to secret '%s/%s' key '%s'", selectedNamespace, selectedSecret, selectedKey),
		IsConfirm: true,
	}
	_, err = confirmPrompt.Run()
	if err != nil {
		fmt.Println("Save cancelled")
		return
	}

	_, err = withTimeoutCtx(func(ctx context.Context) (struct{}, error) {
		err := k8sClient.SaveSecret(ctx, selectedNamespace, selectedSecret, selectedKey, editedData)
		return struct{}{}, err
	})
	if err != nil {
		log.Fatalf("Error saving secret '%s' in namespace '%s': %v", selectedSecret, selectedNamespace, err)
	}

	fmt.Printf("Secret '%s' in namespace '%s' updated successfully.\n", selectedSecret, selectedNamespace)
}

func runPrompt(title string, items []string) string {
	slices.Sort(items)
	prompt := promptui.Select{
		Label:             title,
		Items:             items,
		Size:              10,
		StartInSearchMode: true,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . }}",
		},
		Searcher: func(input string, index int) bool {
			lower := strings.ToLower(input)
			return strings.HasPrefix(strings.ToLower(items[index]), lower)
		},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed: %v", err)
	}
	return result
}

func withTimeoutCtx[T any](f func(context.Context) (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return f(ctx)
}
