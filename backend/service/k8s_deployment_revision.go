package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"

	"devops-backend/global"
	"devops-backend/utils"
	"go.uber.org/zap"
)

type RevisionInfo struct {
	Revision     int       `json:"revision"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	Replicas     int32     `json:"replicas"`
	Image        string    `json:"image"`
	ChangeReason string    `json:"change_reason"`
	IsCurrent    bool      `json:"is_current"`
	YAML         string    `json:"yaml"`
}

type DiffLine struct {
	Type    string `json:"type"`
	OldNum  int    `json:"old_num"`
	NewNum  int    `json:"new_num"`
	Marker  string `json:"marker"`
	Content string `json:"content"`
}

type DiffResult struct {
	Lines        []DiffLine `json:"lines"`
	AddedCount   int        `json:"added_count"`
	RemovedCount int        `json:"removed_count"`
}

func (s *DeploymentDetailService) GetDeploymentRevisions(clusterID uint, namespace, deploymentName string) ([]RevisionInfo, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deploy, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	labelSelector := labels.Set(deploy.Spec.Selector.MatchLabels).AsSelectorPreValidated()
	rsList, err := client.Clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}

	revisions := make([]RevisionInfo, 0)
	currentRevision := deploy.Annotations["deployment.kubernetes.io/revision"]

	for _, rs := range rsList.Items {
		revisionStr := rs.Annotations["deployment.kubernetes.io/revision"]
		if revisionStr == "" {
			continue
		}

		revision := 0
		for _, c := range revisionStr {
			if c >= '0' && c <= '9' {
				revision = revision*10 + int(c-'0')
			}
		}

		if revision == 0 {
			continue
		}

		image := ""
		if len(rs.Spec.Template.Spec.Containers) > 0 {
			image = rs.Spec.Template.Spec.Containers[0].Image
		}

		changeReason := "更新"
		if revision == 1 {
			changeReason = "创建"
		}

		rsYAML, err := yaml.Marshal(&rs)
		if err != nil {
			continue
		}

		revisions = append(revisions, RevisionInfo{
			Revision:     revision,
			Name:         rs.Name,
			CreatedAt:    rs.CreationTimestamp.Time,
			Replicas:     *rs.Spec.Replicas,
			Image:        image,
			ChangeReason: changeReason,
			IsCurrent:    revisionStr == currentRevision,
			YAML:         string(rsYAML),
		})
	}

	sort.Slice(revisions, func(i, j int) bool {
		return revisions[i].Revision > revisions[j].Revision
	})

	if len(revisions) > 10 {
		revisions = revisions[:10]
	}

	return revisions, nil
}

func (s *DeploymentDetailService) CompareRevisions(yaml1, yaml2 string) (*DiffResult, error) {
	lines1 := splitLines(yaml1)
	lines2 := splitLines(yaml2)

	result := &DiffResult{
		Lines: make([]DiffLine, 0),
	}

	lcs := computeLCS(lines1, lines2)

	i, j, lineNum1, lineNum2 := 0, 0, 1, 1

	for k := 0; k < len(lcs); k++ {
		for i < len(lines1) && lines1[i] != lcs[k] {
			result.Lines = append(result.Lines, DiffLine{
				Type:    "added",
				OldNum:  lineNum1,
				NewNum:  0,
				Marker:  "+",
				Content: lines1[i],
			})
			result.AddedCount++
			i++
			lineNum1++
		}

		for j < len(lines2) && lines2[j] != lcs[k] {
			result.Lines = append(result.Lines, DiffLine{
				Type:    "removed",
				OldNum:  0,
				NewNum:  lineNum2,
				Marker:  "-",
				Content: lines2[j],
			})
			result.RemovedCount++
			j++
			lineNum2++
		}

		if i < len(lines1) && j < len(lines2) {
			result.Lines = append(result.Lines, DiffLine{
				Type:    "normal",
				OldNum:  lineNum1,
				NewNum:  lineNum2,
				Marker:  " ",
				Content: lcs[k],
			})
			i++
			j++
			lineNum1++
			lineNum2++
		}
	}

	for i < len(lines1) {
		result.Lines = append(result.Lines, DiffLine{
			Type:    "added",
			OldNum:  lineNum1,
			NewNum:  0,
			Marker:  "+",
			Content: lines1[i],
		})
		result.AddedCount++
		i++
		lineNum1++
	}

	for j < len(lines2) {
		result.Lines = append(result.Lines, DiffLine{
			Type:    "removed",
			OldNum:  0,
			NewNum:  lineNum2,
			Marker:  "-",
			Content: lines2[j],
		})
		result.RemovedCount++
		j++
		lineNum2++
	}

	return result, nil
}

func splitLines(s string) []string {
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func computeLCS(lines1, lines2 []string) []string {
	m := len(lines1)
	n := len(lines2)

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if lines1[i-1] == lines2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	lcs := make([]string, 0)
	i, j := m, n
	for i > 0 && j > 0 {
		if lines1[i-1] == lines2[j-1] {
			lcs = append([]string{lines1[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}

func (s *DeploymentDetailService) RollbackToRevision(clusterID uint, namespace, deploymentName string, revision int) error {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deploy, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	labelSelector := labels.Set(deploy.Spec.Selector.MatchLabels).AsSelectorPreValidated()
	rsList, err := client.Clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return err
	}

	var targetRS *appsv1.ReplicaSet
	for _, rs := range rsList.Items {
		revStr := rs.Annotations["deployment.kubernetes.io/revision"]
		if revStr == fmt.Sprintf("%d", revision) {
			targetRS = &rs
			break
		}
	}

	if targetRS == nil {
		return fmt.Errorf("revision %d not found", revision)
	}

	deploy.Spec.Template = targetRS.Spec.Template

	_, err = client.Clientset.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		global.GVA_LOG.Error("回退失败", zap.Error(err))
		return err
	}

	return nil
}
