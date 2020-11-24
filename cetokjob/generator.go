package cetokjob

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"regexp"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/batch/v1"
)

var envVarNameFmt = regexp.MustCompile(`^[-._a-zA-Z][-._a-zA-Z0-9]*$`)

func makeKJobEnv(data map[string]string) (*[]corev1.EnvVar, error) {
	retEnv := []corev1.EnvVar{}
	for k, v := range data {
		if !envVarNameFmt.MatchString(k) {
			return nil, fmt.Errorf("\"%s\" : a valid environment variable name must consist of alphabetic characters, digits, '_', '-', or '.', and must not start with a digit", k)
		}
		retEnv = append(retEnv, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return &retEnv, nil
}

func makePodTemplateSpec(jobconfig JobConfig, envs *[]corev1.EnvVar) corev1.PodTemplateSpec {

	envFrom := []corev1.EnvFromSource{}
	if len(jobconfig.Configmap) != 0 {
		envFrom = append(envFrom, corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: jobconfig.Configmap,
				},
			},
		})
	}
	if len(jobconfig.Secret) != 0 {
		envFrom = append(envFrom, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: jobconfig.Secret,
				},
			},
		})
	}
	if len(envFrom) == 0 {
		envFrom = nil
	}

	newPodTemplateSpec := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "job",
					Image:   jobconfig.Image,
					Command: jobconfig.Command,
					Args:    jobconfig.Args,
					Env:     *envs,
					EnvFrom: envFrom,
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	return newPodTemplateSpec
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Reset()
	_, _ = h.Write([]byte(s))
	sum := h.Sum32()
	return sum
}

func makeJob(jc JobConfig, envs *[]corev1.EnvVar) *batchv1.Job {
	newPodTemplateSpec := makePodTemplateSpec(jc, envs)
	job := &batchv1.Job{}
	job.Spec.Template = newPodTemplateSpec
	job.Name = fmt.Sprintf("%s-%x-%d", jc.Name, hash(fmt.Sprint(newPodTemplateSpec)), time.Now().Unix())
	return job
}

// JobGenerator is an object to generate jobs
type JobGenerator struct {
	// Configs of job to be Generated
	JobConfigs []JobConfig
	jobClient  v1.JobInterface
}

// NewJobGenerator generate JobGenerator
func NewJobGenerator(ns string, clientset kubernetes.Interface, jcs []JobConfig) *JobGenerator {

	jobClient := clientset.BatchV1().Jobs(ns)

	jc := &JobGenerator{
		JobConfigs: jcs,
		jobClient:  jobClient,
	}

	return jc
}

// GenerateJob creates a job by adding environment variables to a specified job
func (creator *JobGenerator) GenerateJob(envMap map[string]string) ([]batchv1.Job, error) {

	jobEnv, err := makeKJobEnv(envMap)
	if err != nil {
		return nil, err
	}

	generatedJobs := []batchv1.Job{}
	for i, jc := range creator.JobConfigs {
		log.Printf("%d jobconfig: %v", i, jc)

		job := makeJob(jc, jobEnv)

		generatedJob, err := creator.jobClient.Create(context.TODO(), job, metav1.CreateOptions{})
		if err != nil {
			return generatedJobs, err
		}

		generatedJobs = append(generatedJobs, *generatedJob)
	}

	return generatedJobs, nil
}
