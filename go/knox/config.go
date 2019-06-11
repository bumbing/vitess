package knox

import (
	"encoding/json"
	"flag"
)

var (
	pinterestAuthConfig authConfig
)

type subjectList struct {
	TLSSubjects []string `json:"tls_subjects"`
	KnoxRoles   []string `json:"knox_roles"`
}

type authConfig struct {
	// Map of user name to what roles they have
	UserGroups map[string][]string    `json:"user_groups"`
	GroupAuthz map[string]subjectList `json:"group_authz"`
}

func (value *authConfig) Set(v string) error {
	return json.Unmarshal([]byte(v), value)
}
func (value *authConfig) Get() interface{} {
	return value
}

func (value *authConfig) String() string {
	result, err := json.Marshal(value)
	if err != nil {
		return err.Error()
	}
	return string(result)
}

func init() {
	flag.Var(&pinterestAuthConfig, "pinterest_auth_config", (&authConfig{
		UserGroups: map[string][]string{
			"scriptrw": []string{"reader", "writer"}},
		GroupAuthz: map[string]subjectList{"reader": subjectList{TLSSubjects: []string{"spiffe://foo/bar", "foo-%.ec2.pin220.com"}, KnoxRoles: []string{"scriptro", "scriptrw"}}},
	}).String()+" is the format, describing what users exist, which TLS subjects allow them, and which knox roles allow them")
}
