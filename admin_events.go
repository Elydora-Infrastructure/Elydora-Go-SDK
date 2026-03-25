package elydora

import "fmt"

// ListAdminEvents retrieves recent administrative events for the organization.
// An optional limit (1–100) may be provided; pass 0 to use the server default (20).
func (c *Client) ListAdminEvents(limit int) (*ListAdminEventsResponse, error) {
	path := "/v1/admin/events"
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
	}
	var result ListAdminEventsResponse
	if err := c.doGet(path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
