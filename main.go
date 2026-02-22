package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() { //nolint:gocyclo
	var cfg Config
	cfg.Parse()

	if cfg.showVersion {
		fmt.Printf("k8s-secret-editor version %s ("+
			"commit: %s, built at: %s, built by: %s"+
			")\n", version, commit, date, builtBy)
		return
	}

	k8sClient, err := NewK8SClient(cfg.KubeConfig)
	if err != nil {
		fatalf("Error creating Kubernetes client: %v", err)
	}

	editor, err := NewEditor(cfg.EditorPath)
	if err != nil {
		fatalf("Error initializing editor: %v", err)
	}

	namespaces, err := withTimeoutCtx(func(ctx context.Context) ([]string, error) {
		return k8sClient.ListNamespaces(ctx)
	})
	if err != nil {
		fatalf("Error loading namespaces: %v", err)
	}
	selectedNamespace := runPrompt("Select namespace", namespaces)

	secrets, err := withTimeoutCtx(func(ctx context.Context) ([]string, error) {
		return k8sClient.ListSecrets(ctx, selectedNamespace)
	})
	if err != nil {
		fatalf("Error loading secrets: %v", err)
	}
	selectedSecret := runPrompt(fmt.Sprintf("Select secret in '%s'", selectedNamespace), secrets)

	secret, err := withTimeoutCtx(func(ctx context.Context) (SecretData, error) {
		return k8sClient.GetSecret(ctx, selectedNamespace, selectedSecret)
	})
	if err != nil {
		fatalf("Error loading secret: %v", err)
	}
	keys := make([]string, 0, len(secret))
	for k := range secret {
		keys = append(keys, k)
	}
	selectedKey := runPrompt(fmt.Sprintf("Select key in secret '%s'", selectedSecret), keys)
	originData, ok := secret[selectedKey]
	if !ok {
		fatalf("Key '%s' not found in secret '%s' in namespace '%s'", selectedKey, selectedSecret, selectedNamespace)
	}

	tmpFile, err := NewTmpFile()
	if err != nil {
		fatalf("Error creating temp file: %v", err)
	}
	defer tmpFile.Close()
	if err := tmpFile.Write(originData); err != nil {
		fatalf("Error writing secret data to temp file: %v", err)
	}
	if err := tmpFile.OpenEditor(editor); err != nil {
		fatalf("Error opening editor: %v", err)
	}
	editedData, err := tmpFile.Read()
	if err != nil {
		fatalf("Error reading edited data from temp file: %v", err)
	}

	if slices.Equal(originData, editedData) {
		fmt.Println("No changes detected, exiting.")
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(originData), string(editedData), false)
	fmt.Println(dmp.DiffPrettyText(diffs))

	confirmPrompt := promptui.Prompt{
		Label:     fmt.Sprintf("Apply changes to secret '%s/%s' key '%s'", selectedNamespace, selectedSecret, selectedKey),
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
		fatalf("Error saving secret '%s' in namespace '%s': %v", selectedSecret, selectedNamespace, err)
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
		fatalf("Prompt failed: %v", err)
	}
	return result
}

func withTimeoutCtx[T any](f func(context.Context) (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return f(ctx)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, args...))
	os.Exit(1)
}
