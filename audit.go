package elydora

import (
	"fmt"
	"net/url"
	"strconv"
)

// QueryAudit queries the audit log with the given parameters.
func (c *Client) QueryAudit(params *AuditQueryRequest) (*AuditQueryResponse, error) {
	q := url.Values{}
	if params.OrgID != "" {
		q.Set("org_id", params.OrgID)
	}
	if params.AgentID != "" {
		q.Set("agent_id", params.AgentID)
	}
	if params.OperationType != "" {
		q.Set("operation_type", params.OperationType)
	}
	if params.StartTime != nil {
		q.Set("start_time", strconv.FormatInt(*params.StartTime, 10))
	}
	if params.EndTime != nil {
		q.Set("end_time", strconv.FormatInt(*params.EndTime, 10))
	}
	if params.Cursor != "" {
		q.Set("cursor", params.Cursor)
	}
	if params.Limit != nil {
		q.Set("limit", strconv.Itoa(*params.Limit))
	}

	path := "/audit"
	if encoded := q.Encode(); encoded != "" {
		path = fmt.Sprintf("/audit?%s", encoded)
	}

	var result AuditQueryResponse
	if err := c.doGet(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
