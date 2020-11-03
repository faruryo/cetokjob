package cetokjob

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewJobGenerator(t *testing.T) {
	t.Parallel()
	clientset := fake.NewSimpleClientset()
	data := []struct {
		namespace string
	}{
		{
			namespace: "default",
		},
	}

	for _, single := range data {
		t.Run("", func(single struct {
			namespace string
		}) func(t *testing.T) {
			return func(t *testing.T) {
				generator := NewJobGenerator(single.namespace, clientset, []JobConfig{})
				if generator == nil {
					t.Error("expected generator, got nil")
				}
			}
		}(single))
	}
}

func convertToEnvMapFromEnvs(envs []corev1.EnvVar) map[string]string {
	envMap := map[string]string{}
	for _, env := range envs {
		envMap[env.Name] = env.Value
	}
	return envMap
}

func compareContainer(ideal corev1.Container, real corev1.Container) error {
	if ideal.Name != real.Name {
		return fmt.Errorf("Name are different %s := %s", ideal.Name, real.Name)
	}
	if ideal.Image != real.Image {
		return fmt.Errorf("Image are different %s := %s", ideal.Image, real.Image)
	}
	iEnv := convertToEnvMapFromEnvs(ideal.Env)
	rEnv := convertToEnvMapFromEnvs(real.Env)
	if !reflect.DeepEqual(iEnv, rEnv) {
		return fmt.Errorf("Env are different %s := %s", iEnv, rEnv)
	}
	if !reflect.DeepEqual(ideal.EnvFrom, real.EnvFrom) {
		return fmt.Errorf("EnvFrom are different %s := %s", ideal.EnvFrom, real.EnvFrom)
	}
	return nil
}

func compareJob(ideal batchv1.Job, real batchv1.Job) error {
	if strings.HasPrefix(real.Name, ideal.Name) {
		return fmt.Errorf("Generated job name(%s) should start with %s", real.Name, ideal.Name)
	}
	if ideal.Namespace != real.Namespace {
		return fmt.Errorf("Namespaces are different %s := %s", ideal.Namespace, real.Namespace)
	}
	if len(ideal.Spec.Template.Spec.Containers) != len(real.Spec.Template.Spec.Containers) {
		return fmt.Errorf("Containers length are different %d := %d", len(ideal.Spec.Template.Spec.Containers), len(real.Spec.Template.Spec.Containers))
	}
	for i := 0; i < len(ideal.Spec.Template.Spec.Containers); i++ {
		err := compareContainer(ideal.Spec.Template.Spec.Containers[i], real.Spec.Template.Spec.Containers[i])
		if err != nil {
			return err
		}
	}
	if ideal.Spec.Template.Spec.RestartPolicy != real.Spec.Template.Spec.RestartPolicy {
		return fmt.Errorf("RestartPolicy are different %s := %s", ideal.Spec.Template.Spec.RestartPolicy, real.Spec.Template.Spec.RestartPolicy)
	}
	return nil
}

func compareJobs(ideal []batchv1.Job, real []batchv1.Job) error {
	if len(ideal) != len(real) {
		return fmt.Errorf("Length varies len(ideal):%d len(real):%d", len(ideal), len(real))
	}
	for i := 0; i < len(ideal); i++ {
		err := compareJob(ideal[i], real[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func TestGenerateJob(t *testing.T) {
	t.Parallel()
	clientset := fake.NewSimpleClientset()
	data := []struct {
		generator *JobGenerator
		envMap    map[string]string
		jobs      []batchv1.Job
		err       error
	}{
		{
			generator: NewJobGenerator("default", clientset, []JobConfig{}),
		},
		{
			generator: NewJobGenerator("test", clientset, []JobConfig{
				{
					Name:      "test",
					Image:     "test/image",
					Command:   "ls -lt",
					Configmap: "cm",
					Secret:    "sec",
				},
			}),
			envMap: map[string]string{
				"KEY1":       "VALUE1",
				"camelCase":  "Camel",
				"Snake_Case": "Snake",
			},
			jobs: []batchv1.Job{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-job",
						Namespace: "test",
					},
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:    "job",
										Image:   "test/image",
										Command: []string{"ls", "-lt"},
										Env: []corev1.EnvVar{
											{
												Name:  "KEY1",
												Value: "VALUE1",
											},
											{
												Name:  "camelCase",
												Value: "Camel",
											},
											{
												Name:  "Snake_Case",
												Value: "Snake",
											},
										},
										EnvFrom: []corev1.EnvFromSource{
											{
												ConfigMapRef: &corev1.ConfigMapEnvSource{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "cm",
													},
												},
											},
											{
												SecretRef: &corev1.SecretEnvSource{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "sec",
													},
												},
											},
										},
									},
								},
								RestartPolicy: corev1.RestartPolicyNever,
							},
						},
					},
				},
			},
		},
		{
			generator: NewJobGenerator("default", clientset, []JobConfig{}),
			envMap: map[string]string{
				"KE Y1": "VALUE1",
			},
			err: errors.New("\"KE Y1\" : a valid environment variable name must consist of alphabetic characters, digits, '_', '-', or '.', and must not start with a digit"),
		},
	}

	for _, single := range data {
		t.Run("", func(single struct {
			generator *JobGenerator
			envMap    map[string]string
			jobs      []batchv1.Job
			err       error
		}) func(t *testing.T) {
			return func(t *testing.T) {
				jobs, err := single.generator.GenerateJob(single.envMap)
				if err != nil {
					if single.err == nil {
						t.Fatalf(err.Error())
					}
					if !strings.EqualFold(single.err.Error(), err.Error()) {
						t.Fatalf("expected err: %s got err: %s", single.err, err)
					}
				} else {
					if err := compareJobs(single.jobs, jobs); err != nil {
						t.Error(err)
					}
				}
			}
		}(single))
	}
}
