package tags

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Base(project, stack string, extra ...map[string]string) pulumi.StringMap {
	base := pulumi.StringMap{
		"Project": pulumi.String(project),
		"Stack":   pulumi.String(stack),
	}
	if len(extra) > 0 {
		for k, v := range extra[0] {
			base[k] = pulumi.String(v)
		}
	}
	return base
}