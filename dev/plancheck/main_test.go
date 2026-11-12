package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const planJSON = `{"resource_changes":[
 {"address":"google_cloud_run_v2_service.this","change":{"actions":["delete","create"]}},
 {"address":"google_service_account.runtime","change":{"actions":["no-op"]}},
 {"address":"google_project_iam_member.runtime_run_viewer","change":{"actions":["create"]}},
 {"address":"google_cloud_run_v2_service_iam_member.invoker[\"allUsers\"]","change":{"actions":["delete"]}}
]}`

func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		plan    string
		allowed string
		want    string // substring of the error, empty for success
	}{
		{"nothing allowed refuses a replacement", planJSON, "# nothing\n", "google_cloud_run_v2_service.this"},
		{"partially allowed still refuses", planJSON, "google_cloud_run_v2_service.this\n", "invoker"},
		{"all allowed passes", planJSON,
			"# comment\n\ngoogle_cloud_run_v2_service.this\ngoogle_cloud_run_v2_service_iam_member.invoker[\"allUsers\"]\n", ""},
		{"no destroy passes with an empty list", `{"resource_changes":[{"address":"a","change":{"actions":["create"]}}]}`, "", ""},
		{"empty plan passes", `{}`, "", ""},
		{"invalid json fails closed", `not json`, "", "decode plan"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			var out bytes.Buffer
			err := run(write(t, dir, "plan.json", tc.plan), write(t, dir, "allowed.txt", tc.allowed), &out)
			if tc.want == "" {
				if err != nil || !strings.Contains(out.String(), "all allowed") {
					t.Fatalf("err = %v, out = %q", err, out.String())
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestRunMissingFiles(t *testing.T) {
	dir := t.TempDir()
	allowed := write(t, dir, "allowed.txt", "")
	if err := run(filepath.Join(dir, "absent.json"), allowed, &bytes.Buffer{}); err == nil {
		t.Error("a missing plan must fail")
	}
	plan := write(t, dir, "plan.json", "{}")
	if err := run(plan, filepath.Join(dir, "absent.txt"), &bytes.Buffer{}); err == nil {
		t.Error("a missing allow-list must fail closed")
	}
}
