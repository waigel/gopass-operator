package gp_kubernetes

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	errs "errors"
	"fmt"
	"github.com/gopasspw/gopass/pkg/gopass"
	_ "github.com/gopasspw/gopass/pkg/gopass/secrets"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeValidate "k8s.io/apimachinery/pkg/util/validation"
	"reflect"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log
var ErrCannotUpdateSecretType = errs.New("cannot change secret type. Secret type is immutable")

const GopassPrefix = "operator.gopass.waigel.com"
const SecretHashAnnotation = GopassPrefix + "/SecretHash"

func CreateSecretHash(bv []byte) string {
	sha256 := sha256.New()
	sha256.Write(bv)
	return base64.URLEncoding.EncodeToString(sha256.Sum(nil))
}

func CreateKubernetesSecretFromGopassSecret(client client.Client, secretName, secretNamespace string, sec gopass.Secret, autoRestart string,
	labels map[string]string, secretType string, ownerRef *metav1.OwnerReference) error {
	secretAnnotations := map[string]string{
		SecretHashAnnotation: CreateSecretHash(sec.Bytes()),
	}
	secret := BuildKubernetesSecretFromGopassItem(secretName, secretNamespace, secretAnnotations, labels, secretType, sec, ownerRef)
	currentSecret := &corev1.Secret{}
	err := client.Get(context.Background(), types.NamespacedName{Name: secretName, Namespace: secretNamespace}, currentSecret)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Secret", "Secret.Namespace", secretNamespace, "Secret.Name", secretName)
		return client.Create(context.Background(), secret)
	} else if err != nil {
		return err
	}

	//check for changes in secret type
	wantSecretType := secretType
	if wantSecretType == "" {
		wantSecretType = string(corev1.SecretTypeOpaque)
	}
	currentSecretType := string(currentSecret.Type)
	if currentSecretType == "" {
		currentSecretType = string(corev1.SecretTypeOpaque)
	}
	if currentSecretType != wantSecretType {
		return ErrCannotUpdateSecretType
	}

	currentAnnotations := currentSecret.Annotations
	currentLabels := currentSecret.Labels
	if !reflect.DeepEqual(currentAnnotations, secretAnnotations) || !reflect.DeepEqual(currentLabels, labels) {
		log.Info(fmt.Sprintf("Updating Secret %v at namespace '%v'", secret.Name, secret.Namespace))
		currentSecret.ObjectMeta.Annotations = secretAnnotations
		currentSecret.ObjectMeta.Labels = labels
		currentSecret.Data = secret.Data
		if err := client.Update(context.Background(), currentSecret); err != nil {
			return fmt.Errorf("kubernetes secret update failed: %w", err)
		}
		return nil
	}

	log.Info(fmt.Sprintf("Secret with name %v and version %v already exists", secret.Name, secret.Annotations[SecretHashAnnotation]))
	return nil

}

func BuildKubernetesSecretFromGopassItem(name string, namespace string, annotations map[string]string, labels map[string]string, secretType string, sec gopass.Secret, ownerRef *metav1.OwnerReference) *corev1.Secret {
	var ownerRefs []metav1.OwnerReference
	if ownerRef != nil {
		ownerRefs = []metav1.OwnerReference{*ownerRef}
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:            formatSecretName(name),
			Namespace:       namespace,
			Annotations:     annotations,
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		Data: BuildKubernetesSecretData(sec),
		Type: corev1.SecretType(secretType),
	}
}

// ToDo: Build a better type for key / values
func BuildKubernetesSecretData(sec gopass.Secret) map[string][]byte {
	return map[string][]byte{
		"secret": sec.Bytes(),
	}
}

// formatSecretName - format the secret name to valid dns subdomain name
func formatSecretName(name string) string {
	if errs := kubeValidate.IsDNS1123Subdomain(name); len(errs) == 0 {
		return name
	}
	return createValidSecretName(name)
}

// create subdomain dns name from name
var invalidDataChars = regexp.MustCompile("[^a-zA-Z0-9-._]+")
var invalidStartEndChars = regexp.MustCompile("(^[^a-zA-Z0-9-._]+|[^a-zA-Z0-9-._]+$)")

func createValidSecretName(name string) string {
	result := invalidStartEndChars.ReplaceAllString(name, "")
	result = invalidDataChars.ReplaceAllString(result, "-")
	if len(result) > kubeValidate.DNS1123SubdomainMaxLength {
		result = result[0:kubeValidate.DNS1123SubdomainMaxLength]
	}
	return result
}
