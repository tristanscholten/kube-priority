package priorityclass

import (
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestNameForValue(t *testing.T) {
	tests := map[int32]string{500000: "kube-priority-hstr-nl-500000", 0: "kube-priority-hstr-nl-0", -10: "kube-priority-hstr-nl-neg-10", 999999999: "kube-priority-hstr-nl-999999999"}
	for v, want := range tests {
		t.Run(want, func(t *testing.T) {
			got, err := NameForValue(v)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	policy := corev1.PreemptLowerPriority
	pc, err := Build(750000, policy)
	if err != nil {
		t.Fatal(err)
	}
	if pc.Name != "kube-priority-hstr-nl-750000" || pc.Value != 750000 || pc.GlobalDefault {
		t.Fatalf("unexpected PriorityClass: %#v", pc)
	}
	if pc.PreemptionPolicy == nil || *pc.PreemptionPolicy != policy {
		t.Fatalf("bad preemption policy")
	}
	if !IsManaged(pc) {
		t.Fatalf("expected managed labels")
	}
	if pc.Annotations[ValueAnnotation] != "750000" {
		t.Fatalf("missing value annotation")
	}
}
