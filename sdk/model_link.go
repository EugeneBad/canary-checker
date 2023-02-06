/*
 * Canary Checker API
 *
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * API version: 1..1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package swagger

type Link struct {
	Icon string `json:"icon,omitempty"`
	Label string `json:"label,omitempty"`
	Text string `json:"text,omitempty"`
	Tooltip string `json:"tooltip,omitempty"`
	// e.g. documentation, support, playbook
	Type_ string `json:"type,omitempty"`
	Url string `json:"url,omitempty"`
}